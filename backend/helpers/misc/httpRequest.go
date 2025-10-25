package misc

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type FetchResult struct {
	Status int
	Body   []byte
	Err    error
}

func computeBackoff(base, max time.Duration, attempt int) time.Duration {
	pow := math.Pow(2, float64(attempt))
	delay := time.Duration(float64(base) * pow)
	if delay > max {
		delay = max
	}

	jitter := 0.2 * (rand.Float64()*2 - 1) // [-0.2, +0.2]
	return time.Duration(float64(delay) * (1 + jitter))
}

func DoAllWithRetry(ctx context.Context, requests []*http.Request, concurrency int) ([]FetchResult, error) {
	if concurrency < 1 {
		concurrency = len(requests)
	}

	results := make([]FetchResult, len(requests))

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var firstErr atomic.Value

	for i, req := range requests {
		wg.Add(1)
		i, req := i, req

		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				firstErr.Store(ctx.Err())
				return
			}
			defer func() { <-sem }()

			results[i] = DoWithRetry(ctx, req)
		}()
	}

	wg.Wait()

	if v := firstErr.Load(); v != nil {
		return results, v.(error)
	}

	return results, nil
}

func DoWithRetry(ctx context.Context, req *http.Request) FetchResult {
	const (
		backoffBase = 500 * time.Millisecond
		maxBackoff  = 30 * time.Second
		maxErrBody  = 8192 // cap error text we propagate
	)

	// If the request has a body and you want to retry it, make sure it's rewindable.
	// For example, use req.GetBody (Go 1.13+) or buffer the bytes yourself.

	for attempt := 0; ; attempt++ {
		if ctx.Err() != nil {
			return FetchResult{Status: 0, Body: nil, Err: ctx.Err()}
		}

		// For retries with a body, reset it via GetBody if present.
		if attempt > 0 && req.GetBody != nil {
			if rc, err := req.GetBody(); err == nil {
				req.Body = rc
			} else {
				return FetchResult{Status: 0, Body: nil, Err: fmt.Errorf("reset request body: %w", err)}
			}
		}

		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			// Transport-level error; back off and retry unless context canceled.
			delay := computeBackoff(backoffBase, maxBackoff, attempt)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return FetchResult{Status: 0, Body: nil, Err: ctx.Err()}
			}
		}

		// Make sure we always close the body this iteration.
		bodyBytes, readErr := func() ([]byte, error) {
			defer resp.Body.Close()
			b, e := io.ReadAll(resp.Body)
			return b, e
		}()

		if readErr != nil {
			return FetchResult{Status: 0, Body: nil, Err: readErr}
		}

		// Success
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return FetchResult{Status: resp.StatusCode, Body: bodyBytes, Err: nil}
		}

		// Build a concise error message using the response body.
		errText := string(bodyBytes)
		if len(errText) > maxErrBody {
			errText = errText[:maxErrBody] + "…"
		}
		statusErr := fmt.Errorf("http %d %s: %s", resp.StatusCode, http.StatusText(resp.StatusCode), errText)

		// Retry on 429 and 5xx
		if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 599) {
			// Honor Retry-After if present (delta-seconds or HTTP-date)
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, parseErr := strconv.Atoi(strings.TrimSpace(ra)); parseErr == nil && secs >= 0 {
					select {
					case <-time.After(time.Duration(secs) * time.Second):
						continue
					case <-ctx.Done():
						return FetchResult{Status: resp.StatusCode, Body: bodyBytes, Err: ctx.Err()}
					}
				} else if t, parseErr := http.ParseTime(ra); parseErr == nil {
					delay := time.Until(t)
					if delay < 0 {
						delay = 0
					}
					select {
					case <-time.After(delay):
						continue
					case <-ctx.Done():
						return FetchResult{Status: resp.StatusCode, Body: bodyBytes, Err: ctx.Err()}
					}
				}
			}

			delay := computeBackoff(backoffBase, maxBackoff, attempt)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return FetchResult{Status: resp.StatusCode, Body: bodyBytes, Err: ctx.Err()}
			}
		}

		// Non-retryable (most 4xx): return status and the server error text.
		return FetchResult{Status: resp.StatusCode, Body: bodyBytes, Err: statusErr}
	}
}
