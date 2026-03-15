package github

import (
	"net/http"

	"github.com/shurcooL/githubv4"
)

// tokenTransport injects an OAuth token as a Bearer header on every request
type tokenTransport struct {
	token string
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)
	return http.DefaultTransport.RoundTrip(req)
}

// NewGraphQLClient creates a githubv4 client authenticated with the given token
func NewGraphQLClient(token string) *githubv4.Client {
	httpClient := &http.Client{Transport: &tokenTransport{token: token}}
	return githubv4.NewClient(httpClient)
}
