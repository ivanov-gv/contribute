package integration

import (
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *EdgeCaseSuite) TestAddReaction_IssueComment() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusCreated, reactionResponse)

	err := s.reactionService.AddToIssueComment(12345, "+1")
	require.NoError(s.T(), err)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/issues/comments/12345/reactions")
}

func (s *EdgeCaseSuite) TestAddReaction_ReviewComment() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/pulls/comments/", http.StatusCreated, reactionResponse)

	err := s.reactionService.AddToReviewComment(99999, "rocket")
	require.NoError(s.T(), err)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/pulls/comments/99999/reactions")
}

func (s *EdgeCaseSuite) TestAddReaction_InvalidType_IssueComment() {
	err := s.reactionService.AddToIssueComment(12345, "notareaction")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid reaction")
	// invalid reaction type is rejected before any HTTP call
	assert.Equal(s.T(), 0, s.server.RequestCount())
}

func (s *EdgeCaseSuite) TestAddReaction_InvalidType_ReviewComment() {
	err := s.reactionService.AddToReviewComment(99999, "🚀")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid reaction")
	assert.Equal(s.T(), 0, s.server.RequestCount())
}

func (s *EdgeCaseSuite) TestAddReaction_AllValidTypes() {
	validReactions := []string{"+1", "-1", "laugh", "confused", "heart", "hooray", "rocket", "eyes"}
	for _, r := range validReactions {
		s.server.Reset()
		s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusCreated, reactionResponse)

		err := s.reactionService.AddToIssueComment(1, r)
		require.NoError(s.T(), err, "reaction %q should be valid", r)
	}
}

func (s *EdgeCaseSuite) TestAddReaction_ServerError() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusInternalServerError,
		map[string]interface{}{"message": "Internal Server Error"})

	err := s.reactionService.AddToIssueComment(12345, "+1")
	require.Error(s.T(), err)
}
