package github

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"

	"github.com/ivanov-gv/gh-contribute/internal/client/auth"
)

const (
	maxRetries       = 4
	rateLimitHeader  = "X-RateLimit-Remaining"
	retryAfterHeader = "Retry-After"
	rateLimitWarning = 100 // warn when remaining drops below this
)

// initialBackoff is a var so tests can override it to avoid slow retries.
var initialBackoff = 2 * time.Second

// rateLimitTransport wraps a token getter with rate limit handling and exponential backoff.
type rateLimitTransport struct {
	getToken func() (string, error) // called on every request to get a fresh token
	inner    http.RoundTripper      // nil means use http.DefaultTransport
}

func (t *rateLimitTransport) transport() http.RoundTripper {
	if t.inner != nil {
		return t.inner
	}
	return http.DefaultTransport
}

func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	token, err := t.getToken()
	if err != nil {
		return nil, fmt.Errorf("getToken: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	var resp *http.Response
	backoff := initialBackoff

	for attempt := range maxRetries {
		// restore body before each attempt — Clone shallow-copies the Body reference,
		// which is drained after the first round-trip, so POST retries send an empty body.
		if req.GetBody != nil {
			body, bodyErr := req.GetBody()
			if bodyErr != nil {
				return nil, fmt.Errorf("req.GetBody: %w", bodyErr)
			}
			req.Body = body
		}

		resp, err = t.transport().RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// log rate limit headers at warning level when running low
		if remaining := resp.Header.Get(rateLimitHeader); remaining != "" {
			rem, _ := strconv.Atoi(remaining)
			if rem < rateLimitWarning && rem > 0 {
				log.Warn().Int("remaining", rem).Msg("GitHub API rate limit running low")
			}
		}

		// retry on 429 (always rate-limited) or 403 only when GitHub signals rate limiting
		// via X-RateLimit-Remaining:0 or Retry-After header. Other 403s are permission errors
		// that will not resolve on retry.
		isRateLimitSignaled := resp.Header.Get(retryAfterHeader) != "" ||
			resp.Header.Get(rateLimitHeader) == "0"
		shouldRetry := resp.StatusCode == http.StatusTooManyRequests ||
			(resp.StatusCode == http.StatusForbidden && isRateLimitSignaled)

		if shouldRetry {
			resp.Body.Close() //nolint:errcheck,gosec // best-effort close before retry

			// respect Retry-After header if present
			if retryAfter := resp.Header.Get(retryAfterHeader); retryAfter != "" {
				if seconds, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					backoff = time.Duration(seconds) * time.Second
				}
			}

			log.Warn().
				Int("status", resp.StatusCode).
				Int("attempt", attempt+1).
				Dur("backoff", backoff).
				Msg("rate limited, retrying")

			time.Sleep(backoff)
			backoff *= 2 //nolint:mnd // exponential backoff multiplier
			continue
		}

		return resp, nil
	}

	// return last response after exhausting retries
	return resp, nil
}

// NewGraphQLClient creates a githubv4 client with a static token.
func NewGraphQLClient(token string) *githubv4.Client {
	getToken := func() (string, error) { return token, nil }
	httpClient := &http.Client{Transport: &rateLimitTransport{getToken: getToken}}
	return githubv4.NewClient(httpClient)
}

// NewGraphQLClientWithProvider creates a githubv4 client that refreshes tokens automatically
// via the given TokenProvider. Use this for GitHub App installation tokens that expire after 1h.
func NewGraphQLClientWithProvider(provider *auth.TokenProvider) *githubv4.Client {
	httpClient := &http.Client{Transport: &rateLimitTransport{getToken: provider.Token}}
	return githubv4.NewClient(httpClient)
}
