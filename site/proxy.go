package site

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/Noooste/azuretls-client"
	"github.com/Noooste/fhttp/httptrace"
	"golang.org/x/net/http2"
)

func setSiteProxy(session *azuretls.Session, proxy string) error {
	if err := session.SetProxy(proxy); err != nil {
		return err
	}
	proxyURL := session.ProxyDialer.ProxyChain[0]
	if proxyURL.Scheme == "http" || proxyURL.Scheme == "https" {
		authorization := session.ProxyDialer.DefaultHeader.Get("Proxy-Authorization")
		session.Dial = func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialSiteProxyTunnel(ctx, network, address, proxyURL, authorization, session)
		}
		session.PreHooksWithContext = append(session.PreHooksWithContext, traceSiteProxyRequest)
	}
	return nil
}

func traceSiteProxyRequest(request *azuretls.Context) error {
	req := request.Request
	if req.Response != nil {
		return nil // A redirect inherits the trace from the preceding request.
	}
	trace := &httptrace.ClientTrace{GotConn: func(info httptrace.GotConnInfo) {
		conn := info.Conn
		if secure, ok := conn.(interface{ NetConn() net.Conn }); ok {
			conn = secure.NetConn()
		}
		if tunnel, ok := conn.(*proxyTunnel); ok {
			tunnel.finishSetup()
		}
	}}
	req.SetContext(httptrace.WithClientTrace(req.Context(), trace))
	return nil
}

func dialSiteProxyTunnel(ctx context.Context, network, address string, proxy *url.URL, authorization string,
	session *azuretls.Session) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, session.TimeOut)

	dialer := &net.Dialer{}
	dial := dialer.DialContext
	if proxy.Scheme == "https" {
		dial = (&tls.Dialer{
			NetDialer: dialer,
			Config: &tls.Config{
				NextProtos:         []string{http2.NextProtoTLS, "http/1.1"},
				InsecureSkipVerify: session.InsecureSkipVerify,
			},
		}).DialContext
	}
	conn, err := dial(ctx, network, proxy.Host)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}
	raw := conn
	protocol := ""
	if secure, ok := conn.(*tls.Conn); ok {
		protocol = secure.ConnectionState().NegotiatedProtocol
		raw = secure.NetConn()
	}
	// CONNECT only opens the tunnel. Keep its cancellation active until the
	// transport acquires it, after any tracker TLS handshake.
	interrupt := context.AfterFunc(ctx, func() { _ = raw.Close() })
	finishSetup := sync.OnceFunc(func() {
		interrupt()
		cancel()
	})

	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Scheme: proxy.Scheme, Host: address},
		Host:   address,
		Header: http.Header{"User-Agent": {session.UserAgent}},
	}
	if authorization != "" {
		req.Header.Set("Proxy-Authorization", authorization)
	}

	var tunnel net.Conn
	if protocol == http2.NextProtoTLS {
		tunnel, err = connectHTTP2Proxy(req, conn)
	} else {
		tunnel, err = connectHTTP1Proxy(req, conn)
	}
	if err != nil {
		finishSetup()
		_ = conn.Close()
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		finishSetup()
		_ = tunnel.Close()
		return nil, err
	}
	return &proxyTunnel{Conn: tunnel, finishSetup: finishSetup}, nil
}

type proxyTunnel struct {
	net.Conn
	finishSetup func()
}

func (c *proxyTunnel) Close() error {
	c.finishSetup()
	return c.Conn.Close()
}

func connectHTTP1Proxy(req *http.Request, conn net.Conn) (net.Conn, error) {
	if err := req.Write(conn); err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	res, err := http.ReadResponse(reader, req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("proxy tunnel failed: %s", res.Status)
	}
	if reader.Buffered() != 0 {
		return &bufferedProxyConn{Conn: conn, reader: reader}, nil
	}
	return conn, nil
}

func connectHTTP2Proxy(req *http.Request, conn net.Conn) (net.Conn, error) {
	transport := &http2.Transport{DisableCompression: true}
	client, err := transport.NewClientConn(conn)
	if err != nil {
		return nil, err
	}
	input, output := io.Pipe()
	req.Body = input
	res, err := client.RoundTrip(req)
	if err == nil && res.StatusCode >= 200 && res.StatusCode < 300 {
		return &http2ProxyConn{Conn: conn, output: output, input: res.Body}, nil
	}
	_ = client.Close()
	_ = input.Close()
	_ = output.Close()
	if err != nil {
		return nil, err
	}
	_ = res.Body.Close()
	return nil, fmt.Errorf("proxy tunnel failed: %s", res.Status)
}

type bufferedProxyConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedProxyConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

type http2ProxyConn struct {
	net.Conn
	output *io.PipeWriter
	input  io.ReadCloser
}

func (c *http2ProxyConn) Read(p []byte) (int, error) {
	return c.input.Read(p)
}

func (c *http2ProxyConn) Write(p []byte) (int, error) {
	return c.output.Write(p)
}

func (c *http2ProxyConn) Close() error {
	err := c.Conn.Close()
	_ = c.output.Close()
	_ = c.input.Close()
	return err
}
