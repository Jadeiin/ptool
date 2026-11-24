package site

import (
	"context"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/Noooste/azuretls-client"
	"golang.org/x/net/publicsuffix"
)

type siteSessionKey struct {
	name, url, navigator, ja3, http2, proxy string
	insecure                                bool
	timeout, maxRedirects                   int64
}

type siteRedirectStateKey struct{}

type siteRedirectState struct {
	headers azuretls.OrderedHeaders
	origins map[string]*redirectCookieState
}

type redirectCookieState struct {
	jar    *cookiejar.Jar
	scopes *cookiejar.Jar
}

func prepareSiteRequest(request *azuretls.Context) error {
	req := request.Request
	req.NoCookie = true
	ctx := req.Context()
	if ctx == nil {
		ctx = request.Session.Context()
	}
	if ctx.Value(siteRedirectStateKey{}) == nil {
		state := &siteRedirectState{headers: req.OrderedHeaders.Clone()}
		req.SetContext(context.WithValue(ctx, siteRedirectStateKey{}, state))
	}
	return nil
}

func followSiteRedirect(next *azuretls.Request, via []*azuretls.Request) error {
	if uint(len(via)) >= next.MaxRedirects {
		return azuretls.ErrTooManyRedirects
	}
	destination, err := url.Parse(next.Url)
	if err != nil {
		return err
	}
	first := via[0]
	previous := via[len(via)-1]
	source := previous.HttpRequest.URL
	redirect := next.Context().Value(siteRedirectStateKey{}).(*siteRedirectState)
	sourceOrigin := siteOrigin(source)
	destinationOrigin := siteOrigin(destination)
	if destinationOrigin != "" && siteOrigin(first.HttpRequest.URL) == destinationOrigin {
		next.OrderedHeaders = redirect.headers.Clone()
	} else {
		next.OrderedHeaders = redirectHeadersForOtherOrigin(redirect.headers)
		next.Header = nil
	}
	if next.Method != first.Method {
		next.OrderedHeaders = next.OrderedHeaders.Del("Content-Type")
		next.OrderedHeaders = next.OrderedHeaders.Del("Content-Length")
	}
	if referer := azuretls.RefererForURL(source, destination); sourceOrigin != "" && sourceOrigin == destinationOrigin &&
		next.OrderedHeaders.Get("Referer") == "" && referer != "" {
		next.OrderedHeaders.Set("Referer", referer)
	}

	// Keep cookies in this redirect chain for origins eligible for credentials.
	response := &http.Response{Header: http.Header(previous.Response.Header)}
	if cookies := response.Cookies(); sourceOrigin != "" && len(cookies) > 0 {
		if redirect.origins == nil {
			redirect.origins = make(map[string]*redirectCookieState)
		}
		state := redirect.origins[sourceOrigin]
		if state == nil {
			options := &cookiejar.Options{PublicSuffixList: publicsuffix.List}
			jar, _ := cookiejar.New(options)
			scopes, _ := cookiejar.New(options)
			state = &redirectCookieState{jar: jar, scopes: scopes}
			redirect.origins[sourceOrigin] = state
		}
		state.jar.SetCookies(source, cookies)
		// Keep a scope marker even when Set-Cookie removes a value. The jar
		// applies the same domain, path, and Secure rules to each marker.
		for _, scope := range cookies {
			scope.Value = ""
			scope.MaxAge = 0
			scope.Expires = time.Time{}
		}
		state.scopes.SetCookies(source, cookies)
	}
	state := redirect.origins[destinationOrigin]
	if state == nil {
		return nil
	}
	cookieRequest := &http.Request{Header: http.Header{"Cookie": {next.OrderedHeaders.Get("Cookie")}}}
	explicit := cookieRequest.Cookies()
	cookieRequest.Header.Del("Cookie")
	updated := make(map[string]bool)
	for _, scope := range state.scopes.Cookies(destination) {
		updated[scope.Name] = true
	}
	for _, cookie := range explicit {
		if !updated[cookie.Name] {
			cookieRequest.AddCookie(cookie)
		}
	}
	for _, cookie := range state.jar.Cookies(destination) {
		cookieRequest.AddCookie(cookie)
	}
	if cookie := cookieRequest.Header.Get("Cookie"); cookie != "" {
		next.OrderedHeaders.Set("Cookie", cookie)
	} else {
		next.OrderedHeaders = next.OrderedHeaders.Del("Cookie")
	}
	return nil
}

func redirectHeadersForOtherOrigin(headers azuretls.OrderedHeaders) azuretls.OrderedHeaders {
	safe := make(azuretls.OrderedHeaders, 0, len(headers))
	for _, header := range headers {
		if len(header) == 0 {
			continue
		}
		switch http.CanonicalHeaderKey(header[0]) {
		case "Accept", "Accept-Encoding", "Accept-Language", "Cache-Control", "Content-Type", "Range", "User-Agent":
			safe = append(safe, header)
		}
	}
	return safe.Clone()
}

// siteOrigin returns a credential scope for URLs with ASCII hostnames.
func siteOrigin(u *url.URL) string {
	host := u.Hostname()
	if host == "" || strings.ContainsFunc(host, func(r rune) bool { return r > unicode.MaxASCII }) {
		return ""
	}
	port := u.Port()
	if port == "" {
		port = "80"
		if u.Scheme == "https" {
			port = "443"
		}
	}
	return u.Scheme + "://" + net.JoinHostPort(strings.ToLower(host), port)
}
