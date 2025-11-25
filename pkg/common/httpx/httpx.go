package httpx

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

const DefaultUserAgent = "copilot-worker/0.1"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Doer matches http.Client.Do; used for dependency injection/testing.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// New creates an HTTP client with sane defaults (timeouts, pooling).
func New() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        64,
		MaxIdleConnsPerHost: 16,
		IdleConnTimeout:     60 * time.Second,
		DisableCompression:  false,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}
}

// SetUserAgent ensures every outbound request carries a descriptive UA header.
func SetUserAgent(req *http.Request) {
	if req == nil {
		return
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", DefaultUserAgent)
	}
}

// DoWithRetry retries 5xx/429 responses with exponential backoff + jitter.
func DoWithRetry(ctx context.Context, c Doer, req *http.Request, maxAttempts int) (*http.Response, error) {
	var resp *http.Response
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err = c.Do(req)
		if err == nil && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return resp, nil
		}
		var wait time.Duration
		if resp != nil {
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, convErr := strconv.Atoi(ra); convErr == nil {
					wait = time.Duration(secs) * time.Second
				}
			}
			_ = resp.Body.Close()
		}
		if wait == 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond
			wait = backoff + jitter
		}
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if err == nil && resp != nil {
		return resp, fmt.Errorf("max retries exceeded: %s", resp.Status)
	}
	return nil, err
}
