// Package testserver provides a mock GitHub API server for integration tests.
// It handles both GraphQL (/graphql) and REST (/repos/...) requests,
// returning canned JSON responses matched by pattern.
package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
)

// Request records an incoming HTTP request for test assertions.
type Request struct {
	Method string
	Path   string
	Body   string
}

// graphQLRule matches a GraphQL request by query substring and returns a canned response.
type graphQLRule struct {
	queryContains string
	statusCode    int
	response      json.RawMessage
}

// restRule matches a REST request by method and path prefix, returning a canned response.
type restRule struct {
	method     string
	path       string
	statusCode int
	response   json.RawMessage
}

// Server is a mock GitHub API server that returns canned responses.
// Register responses with OnGraphQL / OnREST before each test, then Reset between tests.
type Server struct {
	*httptest.Server
	mu           sync.Mutex
	graphqlRules []graphQLRule
	restRules    []restRule
	requests     []Request
}

// New creates and starts a new mock server. Call Close() when done.
func New() *Server {
	s := &Server{}
	s.Server = httptest.NewServer(http.HandlerFunc(s.handler))
	return s
}

// Reset clears all registered rules and request log — call in SetupTest.
func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graphqlRules = nil
	s.restRules = nil
	s.requests = nil
}

// OnGraphQL registers a canned GraphQL response returned when the query body contains queryContains.
// The response is automatically wrapped in {"data": response} if not already wrapped.
// Rules are checked in registration order; first match wins.
func (s *Server) OnGraphQL(queryContains string, response interface{}) {
	s.onGraphQLStatus(queryContains, http.StatusOK, response)
}

// OnGraphQLError registers a GraphQL error response returned when the query body contains queryContains.
// statusCode can be 200 (for GraphQL-layer errors) or 4xx/5xx for transport errors.
func (s *Server) OnGraphQLError(queryContains string, statusCode int, errorMessage, errorType string) {
	body, _ := json.Marshal(map[string]interface{}{
		"errors": []map[string]interface{}{
			{"message": errorMessage, "type": errorType},
		},
	})
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graphqlRules = append(s.graphqlRules, graphQLRule{
		queryContains: queryContains,
		statusCode:    statusCode,
		response:      body,
	})
}

// OnGraphQLStatus registers a GraphQL response with a specific HTTP status code.
func (s *Server) onGraphQLStatus(queryContains string, statusCode int, response interface{}) {
	wrapped := wrapData(response)
	data, _ := json.Marshal(wrapped)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graphqlRules = append(s.graphqlRules, graphQLRule{
		queryContains: queryContains,
		statusCode:    statusCode,
		response:      data,
	})
}

// OnGraphQLRaw registers a raw JSON GraphQL response (no wrapping applied).
func (s *Server) OnGraphQLRaw(queryContains string, statusCode int, rawJSON string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.graphqlRules = append(s.graphqlRules, graphQLRule{
		queryContains: queryContains,
		statusCode:    statusCode,
		response:      json.RawMessage(rawJSON),
	})
}

// OnREST registers a canned REST response for the given method and path prefix.
// Rules are checked in registration order; first match wins.
func (s *Server) OnREST(method, path string, statusCode int, response interface{}) {
	data, _ := json.Marshal(response)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.restRules = append(s.restRules, restRule{
		method:     method,
		path:       path,
		statusCode: statusCode,
		response:   data,
	})
}

// Requests returns a copy of all logged requests for test assertions.
func (s *Server) Requests() []Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Request, len(s.requests))
	copy(out, s.requests)
	return out
}

// RequestCount returns how many requests have been received.
func (s *Server) RequestCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.requests)
}

// GraphQLURL returns the URL of the mock server's GraphQL endpoint.
func (s *Server) GraphQLURL() string {
	return s.URL + "/graphql"
}

// handler dispatches incoming requests to GraphQL or REST handlers.
func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	s.mu.Lock()
	s.requests = append(s.requests, Request{Method: r.Method, Path: r.URL.Path, Body: string(body)})
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	// add a dummy rate-limit header so the transport doesn't warn
	w.Header().Set("X-RateLimit-Remaining", "4999")

	if r.URL.Path == "/graphql" {
		s.handleGraphQL(w, body)
		return
	}
	s.handleREST(w, r)
}

// handleGraphQL matches the query body against registered rules and writes the response.
func (s *Server) handleGraphQL(w http.ResponseWriter, body []byte) {
	var req struct {
		Query string `json:"query"`
	}
	_ = json.Unmarshal(body, &req)

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rule := range s.graphqlRules {
		if strings.Contains(req.Query, rule.queryContains) {
			w.WriteHeader(rule.statusCode)
			_, _ = w.Write(rule.response)
			return
		}
	}

	// no match — return a structured error so tests fail clearly
	w.WriteHeader(http.StatusOK)
	msg := "no registered rule matched query"
	if rule := shortQuery(req.Query); rule != "" {
		msg += ": " + rule
	}
	resp, _ := json.Marshal(map[string]interface{}{
		"errors": []map[string]interface{}{
			{"message": msg, "type": "MOCK_NO_MATCH"},
		},
	})
	_, _ = w.Write(resp)
}

// handleREST matches method+path against registered rules and writes the response.
func (s *Server) handleREST(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rule := range s.restRules {
		if rule.method == r.Method && strings.HasPrefix(r.URL.Path, rule.path) {
			w.WriteHeader(rule.statusCode)
			_, _ = w.Write(rule.response)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"message":"Not Found"}`))
}

// wrapData wraps a response value in {"data": v} if it isn't already wrapped.
func wrapData(v interface{}) interface{} {
	if v == nil {
		return map[string]interface{}{"data": nil}
	}
	// if already a map with "data" key, return as-is
	if m, ok := v.(map[string]interface{}); ok {
		if _, hasData := m["data"]; hasData {
			return v
		}
		if _, hasErrors := m["errors"]; hasErrors {
			return v
		}
	}
	return map[string]interface{}{"data": v}
}

// shortQuery returns the first 80 chars of a query for diagnostic messages.
func shortQuery(q string) string {
	q = strings.TrimSpace(q)
	if len(q) > 80 { //nolint:mnd // diagnostic truncation length
		return q[:80] + "..."
	}
	return q
}
