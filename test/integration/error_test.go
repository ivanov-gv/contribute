//go:build !integration

package integration

import (
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: "data":null response is NOT tested here. When the GitHub API returns
// {"data":null}, mapPR dereferences a nil URL field and panics — a known bug
// to be fixed separately (nil-safe URL handling in mapPR).

// TestGraphQL_ServerError — 500 from the GraphQL endpoint propagates as an error (no panic).
func (s *EdgeCaseSuite) TestGraphQL_ServerError() {
	s.server.OnGraphQLRaw(prQueryPattern, http.StatusInternalServerError,
		`{"message":"Internal Server Error"}`)

	_, err := s.prService.Get(42)
	require.Error(s.T(), err)
}

// TestGraphQL_GraphQLErrors — 200 with {"errors":[...]} propagates as an error.
func (s *EdgeCaseSuite) TestGraphQL_GraphQLErrors() {
	s.server.OnGraphQLError(prQueryPattern, http.StatusOK,
		"Could not resolve to a PullRequest with the number of 999.", "NOT_FOUND")

	_, err := s.prService.Get(999)
	require.Error(s.T(), err)
}

// TestREST_NotFound — 404 from REST propagates as a wrapped error.
func (s *EdgeCaseSuite) TestREST_NotFound() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/", http.StatusNotFound,
		map[string]interface{}{"message": "Not Found"})

	_, err := s.commentService.Post(42, "hello")
	require.Error(s.T(), err)
}

// TestREST_UnprocessableEntity — 422 from REST propagates as a wrapped error.
func (s *EdgeCaseSuite) TestREST_UnprocessableEntity() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusUnprocessableEntity,
		map[string]interface{}{"message": "Validation Failed"})

	err := s.reactionService.AddToIssueComment(12345, "+1")
	require.Error(s.T(), err)
}

// TestREST_ServerError — 500 from REST propagates as a wrapped error.
func (s *EdgeCaseSuite) TestREST_ServerError() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/pulls/", http.StatusInternalServerError,
		map[string]interface{}{"message": "Internal Server Error"})

	_, err := s.commentService.ReplyToReviewComment(42, 5001, "reply")
	require.Error(s.T(), err)
}

// TestGraphQL_NoMatchingRule — verifies the mock server returns a structured error
// when no rule matches, so tests fail clearly rather than silently.
func (s *EdgeCaseSuite) TestGraphQL_NoMatchingRule() {
	// register NO rules — any query should return MOCK_NO_MATCH error
	_, err := s.prService.Get(42)
	require.Error(s.T(), err)
}

// TestGraphQL_MultipleRules_FirstMatchWins — rules are checked in order; first match wins.
func (s *EdgeCaseSuite) TestGraphQL_MultipleRules_FirstMatchWins() {
	s.server.OnGraphQL(prQueryPattern, prData(42, "First rule wins", "OPEN"))
	s.server.OnGraphQL(prQueryPattern, prData(42, "Second rule loses", "CLOSED"))

	info, err := s.prService.Get(42)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "First rule wins", info.Title)
}
