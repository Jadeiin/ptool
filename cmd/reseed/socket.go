package reseed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/googollee/go-socket.io/parser"
)

func requestReseedResults(baseURL, token string, files File, idleTimeout time.Duration,
	onResult func(ReseedResult)) error {
	endpoint, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	if idleTimeout <= 0 {
		return fmt.Errorf("Reseed timeout must be positive")
	}
	endpoint.Path = "/socket.io/"
	endpoint.RawPath = ""
	query := endpoint.Query()
	query.Set("EIO", "3")
	query.Set("transport", "polling")
	endpoint.RawQuery = query.Encode()
	polling := &reseedPolling{endpoint: *endpoint, token: token}
	event, err := json.Marshal([]any{"file", files})
	if err != nil {
		return fmt.Errorf("failed to encode Reseed file list: %w", err)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	var workers sync.WaitGroup
	defer func() {
		cancel(nil)
		workers.Wait()
		polling.closePoll()
	}()
	startup := time.AfterFunc(time.Minute, func() { cancel(context.DeadlineExceeded) })
	params, err := polling.open(ctx)
	startup.Stop()
	if err != nil {
		return fmt.Errorf("failed to connect to Reseed backend: %w", err)
	}
	type readOutcome struct {
		result ReseedResult
		err    error
	}
	outcomes := make(chan readOutcome, 1)
	workers.Add(2)
	go func() {
		defer workers.Done()
		if err := polling.write(ctx, params, append([]byte("42"), event...)); err != nil {
			cancel(err)
		}
	}()
	go func() {
		defer workers.Done()
		decoder := parser.NewDecoder(reseedFrameReader{ctx: ctx, polling: polling})
		defer decoder.Close()
		for range len(files) {
			result, err := readReseedResult(decoder)
			select {
			case outcomes <- readOutcome{result: result, err: err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()

	for range len(files) {
		idle := time.NewTimer(idleTimeout)
		var outcome readOutcome
		select {
		case outcome = <-outcomes:
		case <-ctx.Done():
			idle.Stop()
			return context.Cause(ctx)
		case <-idle.C:
			// A match can already be queued when the timer fires.
			select {
			case outcome = <-outcomes:
			default:
				return nil // Return the matches received so far.
			}
		}
		idle.Stop()
		if outcome.err != nil {
			if cause := context.Cause(ctx); cause != nil {
				return cause
			}
			return outcome.err
		}
		onResult(outcome.result)
	}
	return nil
}

func readReseedResult(decoder *parser.Decoder) (ReseedResult, error) {
	for {
		var packet parser.Header
		var event string
		if err := decoder.DecodeHeader(&packet, &event); err != nil {
			return ReseedResult{}, fmt.Errorf("failed to read Reseed event: %w", err)
		}
		if packet.Type != parser.Event || event != "reseed result" ||
			(packet.Namespace != "" && packet.Namespace != "/") {
			_ = decoder.DiscardLast()
			continue
		}
		args, err := decoder.DecodeArgs([]reflect.Type{reflect.TypeFor[ReseedResult]()})
		if err != nil {
			return ReseedResult{}, fmt.Errorf("failed to decode Reseed result: %w", err)
		}
		if len(args) != 1 {
			return ReseedResult{}, fmt.Errorf("Reseed result is missing its payload")
		}
		return args[0].Interface().(ReseedResult), nil
	}
}
