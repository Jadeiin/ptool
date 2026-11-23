package reseed

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/googollee/go-socket.io/engineio/frame"
	"github.com/googollee/go-socket.io/engineio/packet"
	"github.com/googollee/go-socket.io/engineio/payload"
	"github.com/googollee/go-socket.io/engineio/session"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/utils"
)

type reseedPolling struct {
	endpoint  url.URL
	token     string
	firstPing time.Time
	poll      *reseedPoll
}

func (p *reseedPolling) open(ctx context.Context) (transport.ConnParameters, error) {
	poll, err := p.startPoll(ctx)
	if err != nil {
		return transport.ConnParameters{}, err
	}
	p.poll = poll
	first, err := poll.nextPacket()
	if err != nil {
		return transport.ConnParameters{}, err
	}
	if first.PType != packet.OPEN {
		return transport.ConnParameters{}, fmt.Errorf("missing Engine.IO open packet")
	}
	params, err := transport.ReadConnParameters(bytes.NewReader(first.Data))
	if err != nil {
		return transport.ConnParameters{}, err
	}
	if params.SID == "" || params.PingInterval <= 0 || params.PingTimeout <= 0 {
		return transport.ConnParameters{}, fmt.Errorf("invalid Engine.IO session parameters")
	}
	p.firstPing = time.Now().Add(params.PingInterval)
	// Closing the OPEN frame can wait for the next header or the end of the
	// response. Do this before starting the result idle timer.
	if err := poll.closeFrame(); err != nil {
		return transport.ConnParameters{}, err
	}
	query := p.endpoint.Query()
	query.Set("sid", params.SID)
	p.endpoint.RawQuery = query.Encode()
	return params, nil
}

func (p *reseedPolling) request(ctx context.Context, method string, body io.Reader) (*http.Response, error) {
	endpoint := p.endpoint
	query := endpoint.Query()
	query.Set("t", utils.Timestamp())
	endpoint.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/octet-stream")
	if p.token != "" {
		request.Header.Set("Authorization", "Bearar "+p.token)
	}
	client := *http.DefaultClient
	if method == http.MethodGet && (client.Timeout <= 0 || client.Timeout > time.Minute) {
		client.Timeout = time.Minute
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		_ = response.Body.Close()
		return nil, fmt.Errorf("Reseed polling %s returned HTTP %d", method, response.StatusCode)
	}
	return response, nil
}

func (p *reseedPolling) send(ctx context.Context, data []byte) error {
	// Outgoing Reseed packets are ASCII. Binary polling prefixes the text frame with
	// zero, decimal length digits as bytes 0–9, and the 0xff delimiter.
	header := strconv.AppendInt([]byte{0}, int64(len(data)), 10)
	for i := 1; i < len(header); i++ {
		header[i] -= '0'
	}
	body := io.MultiReader(bytes.NewReader(append(header, 0xff)), bytes.NewReader(data))
	response, err := p.request(ctx, http.MethodPost, body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// The status already confirmed the POST. Drain the acknowledgment for reuse.
	_, _ = io.Copy(io.Discard, response.Body)
	return nil
}

func (p *reseedPolling) write(ctx context.Context, params transport.ConnParameters, filePacket []byte) error {
	startup, stopStartup := context.WithTimeout(ctx, time.Minute)
	defer stopStartup()
	if err := p.send(startup, []byte("40")); err != nil {
		return fmt.Errorf("failed to open Reseed namespace: %w", err)
	}
	if err := p.send(startup, filePacket); err != nil {
		return fmt.Errorf("failed to send Reseed file list: %w", err)
	}
	stopStartup()

	timer := time.NewTimer(time.Until(p.firstPing))
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			ping, cancel := context.WithTimeout(ctx, params.PingTimeout)
			nextPing := time.Now().Add(params.PingInterval)
			err := p.send(ping, []byte("2"))
			cancel()
			if err != nil {
				return fmt.Errorf("failed to send Reseed heartbeat: %w", err)
			}
			timer.Reset(time.Until(nextPing))
		}
	}
}

func (p *reseedPolling) startPoll(ctx context.Context) (*reseedPoll, error) {
	response, err := p.request(ctx, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
	if err != nil {
		_ = response.Body.Close()
		return nil, err
	}
	binary := mediaType == "application/octet-stream"
	if !binary && mediaType != "text/plain" {
		_ = response.Body.Close()
		return nil, fmt.Errorf("invalid Reseed polling content type %q", mediaType)
	}

	poll := &reseedPoll{
		body:  response.Body,
		codec: payload.New(binary),
		done:  make(chan struct{}),
	}
	go func() {
		defer close(poll.done)
		defer poll.codec.Close()
		_ = poll.codec.FeedIn(poll.body, binary)
	}()
	return poll, nil
}

type reseedPoll struct {
	body     io.ReadCloser
	codec    *payload.Payload
	done     chan struct{}
	frame    io.ReadCloser
	finished bool
}

func (p *reseedPoll) closeFrame() error {
	if p.frame == nil {
		return nil
	}
	err := p.frame.Close()
	p.frame = nil
	p.finished = err != nil
	return err
}

func (p *reseedPoll) advance() (frame.Type, packet.Type, error) {
	if err := p.closeFrame(); err != nil {
		return 0, 0, err
	}
	ft, pt, reader, err := p.codec.NextReader()
	p.frame = reader
	p.finished = err != nil
	return ft, pt, err
}

func (p *reseedPoll) nextPacket() (packet.Packet, error) {
	ft, pt, err := p.advance()
	if err != nil {
		return packet.Packet{}, err
	}
	data, err := io.ReadAll(p.frame)
	// Close prefetches the next header and can block. Advance on the next read.
	return packet.Packet{FType: ft, PType: pt, Data: data}, err
}

func (p *reseedPolling) closePoll() {
	if p.poll == nil {
		return
	}
	poll := p.poll
	p.poll = nil
	_ = poll.body.Close()
	// FeedIn exits once the decoder consumes the buffered frames.
	for !poll.finished {
		_, _, _ = poll.advance()
	}
	<-poll.done
}

func (p *reseedPolling) nextPacket(ctx context.Context) (packet.Packet, error) {
	for ctx.Err() == nil {
		firstPacket := p.poll == nil
		if firstPacket {
			poll, err := p.startPoll(ctx)
			if err != nil {
				return packet.Packet{}, err
			}
			p.poll = poll
		}
		frame, err := p.poll.nextPacket()
		if err == nil {
			return frame, nil
		}
		p.closePoll()
		if !errors.Is(err, io.EOF) {
			return packet.Packet{}, err
		}
		if firstPacket {
			return packet.Packet{}, errors.New("empty Reseed polling response")
		}
	}
	return packet.Packet{}, ctx.Err()
}

type reseedFrameReader struct {
	ctx     context.Context
	polling *reseedPolling
}

func (r reseedFrameReader) NextReader() (session.FrameType, io.ReadCloser, error) {
	for {
		frame, err := r.polling.nextPacket(r.ctx)
		if err != nil {
			return 0, nil, err
		}
		switch frame.PType {
		case packet.MESSAGE:
			return session.FrameType(frame.FType), io.NopCloser(bytes.NewReader(frame.Data)), nil
		case packet.CLOSE:
			return 0, nil, io.EOF
		}
	}
}
