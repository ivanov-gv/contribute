package github

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
)

const (
	maxRetries       = 4
	rateLimitHeader  = "X-RateLimit-Remaining"
	retryAfterHeader = "Retry-After"
	rateLimitWarning = 100 // warn when remaining drops below this
)

// initialBackoff is a var so tests can override it to avoid slow retries.
var initialBackoff = 2 * time.Second

// rateLimitTransport wraps tokenTransport with rate limit handling and exponential backoff
type rateLimitTransport struct {
	token string
	inner http.RoundTripper // nil means use http.DefaultTransport
}

func (t *rateLimitTransport) transport() http.RoundTripper {
	if t.inner != nil {
		return t.inner
	}
	return http.DefaultTransport
}

func (t *rateLimitTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)

	var resp *http.Response
	var err error
	backoff := initialBackoff

	for attempt := range maxRetries {
		resp, err = t.transport().RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// log rate limit headers at debug level
		if remaining := resp.Header.Get(rateLimitHeader); remaining != "" {
			rem, _ := strconv.Atoi(remaining)
			if rem < rateLimitWarning && rem > 0 {
				log.Warn().Int("remaining", rem).Msg("GitHub API rate limit running low")
			}
		}

		// retry on rate limit (403 or 429)
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
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

// NewGraphQLClient creates a githubv4 client with rate limit handling
func NewGraphQLClient(token string) *githubv4.Client {
	httpClient := &http.Client{Transport: &rateLimitTransport{token: token}}
	return githubv4.NewClient(httpClient)
}
