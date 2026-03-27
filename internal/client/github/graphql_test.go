package github

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTransport creates a rateLimitTransport with a custom inner transport for testing.
func makeTransport(token string, inner http.RoundTripper) *rateLimitTransport {
	return &rateLimitTransport{token: token, inner: inner}
}

func TestRateLimitTransport_AuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	transport := makeTransport("my-secret-token", http.DefaultTransport)
	client := &http.Client{Transport: transport}

	resp, err := client.Get(srv.URL) //nolint:noctx // test helper, no request context needed
	require.NoError(t, err)
	resp.Body.Close() //nolint:errcheck,gosec // test cleanup

	assert.Equal(t, "Bearer my-secret-token", gotAuth)
}

func TestRateLimitTransport_SuccessPassthrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "4500")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	transport := makeTransport("token", http.DefaultTransport)
	client := &http.Client{Transport: transport}

	resp, err := client.Get(srv.URL) //nolint:noctx // test helper, no request context needed
	require.NoError(t, err)
	resp.Body.Close() //nolint:errcheck,gosec // test cleanup
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRateLimitTransport_RetryOn429_EventuallySucceeds(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n < 3 { //nolint:mnd // fail first two attempts
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	// use a fast backoff by injecting a transport with negligible delay — we override
	// initialBackoff via a short initial wait by patching the transport's inner with
	// a custom RoundTripper that responds instantly
	transport := makeTransport("token", http.DefaultTransport)
	// patch initialBackoff to 1ms for test speed
	origBackoff := initialBackoff
	const testBackoff = 1 * time.Millisecond
	initialBackoff = testBackoff
	t.Cleanup(func() { initialBackoff = origBackoff })

	client := &http.Client{Transport: transport}
	resp, err := client.Get(srv.URL) //nolint:noctx // test helper, no request context needed
	require.NoError(t, err)
	resp.Body.Close() //nolint:errcheck,gosec // test cleanup

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), callCount.Load()) //nolint:mnd // expects 3 calls: 2 failures + 1 success
}

func TestRateLimitTransport_RetryOn403(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	origBackoff := initialBackoff
	initialBackoff = 1 * time.Millisecond
	t.Cleanup(func() { initialBackoff = origBackoff })

	transport := makeTransport("token", http.DefaultTransport)
	client := &http.Client{Transport: transport}

	resp, err := client.Get(srv.URL) //nolint:noctx // test helper, no request context needed
	require.NoError(t, err)
	resp.Body.Close() //nolint:errcheck,gosec // test cleanup

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), callCount.Load()) //nolint:mnd // 1 failure + 1 success
}

func TestRateLimitTransport_MaxRetries_ReturnsLastResponse(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		// always rate-limit
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":"rate limited"}`))
	}))
	defer srv.Close()

	origBackoff := initialBackoff
	initialBackoff = 1 * time.Millisecond
	t.Cleanup(func() { initialBackoff = origBackoff })

	transport := makeTransport("token", http.DefaultTransport)
	client := &http.Client{Transport: transport}

	resp, err := client.Get(srv.URL) //nolint:noctx // test helper, no request context needed
	require.NoError(t, err)          // no Go-level error — HTTP error is in the response
	resp.Body.Close()                //nolint:errcheck,gosec // test cleanup

	// should have tried maxRetries times and returned the last 429
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, int32(maxRetries), callCount.Load())
}

func TestRateLimitTransport_RespectsRetryAfterHeader(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := callCount.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0") // 0 seconds — test stays fast
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	origBackoff := initialBackoff
	initialBackoff = 1 * time.Millisecond
	t.Cleanup(func() { initialBackoff = origBackoff })

	transport := makeTransport("token", http.DefaultTransport)
	client := &http.Client{Transport: transport}

	resp, err := client.Get(srv.URL) //nolint:noctx // test helper, no request context needed
	require.NoError(t, err)
	resp.Body.Close() //nolint:errcheck,gosec // test cleanup

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), callCount.Load()) //nolint:mnd // 1 failure + 1 success
}

func TestNewGraphQLClient_CreatesClient(t *testing.T) {
	// smoke test: NewGraphQLClient returns a non-nil client
	client := NewGraphQLClient("test-token")
	assert.NotNil(t, client)
}
