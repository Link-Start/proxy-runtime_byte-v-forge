package runtimehttp

import (
	"context"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type retryPolicy struct {
	maxAttempts int
	baseBackoff time.Duration
	maxBackoff  time.Duration
}

var defaultRetryPolicy = retryPolicy{
	maxAttempts: 3,
	baseBackoff: 200 * time.Millisecond,
	maxBackoff:  5 * time.Second,
}

// retryTransport retries only idempotent requests on transient transport
// errors and 429/5xx responses. It deliberately logs nothing: a request URL or
// header can carry api keys or proxy credentials, so the retry stays silent and
// leaves all reporting to the (already sanitized) caller.
type retryTransport struct {
	base   http.RoundTripper
	policy retryPolicy
}

func newRetryTransport(base http.RoundTripper, policy retryPolicy) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if policy.maxAttempts < 1 {
		policy.maxAttempts = 1
	}
	return &retryTransport{base: base, policy: policy}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !retryableRequest(req) {
		return t.base.RoundTrip(req)
	}
	ctx := req.Context()
	var resp *http.Response
	var err error
	for attempt := 1; attempt <= t.policy.maxAttempts; attempt++ {
		if attempt > 1 {
			wait := t.backoff(attempt, resp)
			drainResponse(resp)
			if cerr := sleepContext(ctx, wait); cerr != nil {
				return nil, cerr
			}
			if req.GetBody != nil {
				body, berr := req.GetBody()
				if berr != nil {
					return nil, berr
				}
				req.Body = body
			}
		}
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, err
			}
			continue
		}
		if !retryableStatus(resp.StatusCode) {
			return resp, nil
		}
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (t *retryTransport) backoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if ra := ParseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
			if ra > t.policy.maxBackoff {
				return t.policy.maxBackoff
			}
			return ra
		}
	}
	shift := attempt - 2
	if shift < 0 {
		shift = 0
	}
	if shift > 16 {
		shift = 16
	}
	exp := t.policy.baseBackoff << shift
	if exp <= 0 || exp > t.policy.maxBackoff {
		exp = t.policy.maxBackoff
	}
	return rand.N(exp)
}

func retryableRequest(req *http.Request) bool {
	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodPut:
	default:
		return false
	}
	if req.Body != nil && req.Body != http.NoBody && req.GetBody == nil {
		return false
	}
	return true
}

func retryableStatus(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func drainResponse(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	_ = resp.Body.Close()
}

// ParseRetryAfter parses a Retry-After header value expressed as a delay in
// seconds or as an HTTP date, returning 0 when absent or already elapsed.
func ParseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
		return 0
	}
	if at, err := http.ParseTime(value); err == nil {
		if d := time.Until(at); d > 0 {
			return d
		}
	}
	return 0
}
