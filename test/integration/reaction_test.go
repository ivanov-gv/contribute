//go:build !integration

package integration

import (
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// reviewNodeData builds a minimal GraphQL response for the reaction service's findReviewNodeID query.
func reviewNodeData(nodeID string, databaseID int64) map[string]interface{} {
	return map[string]interface{}{
		"repository": map[string]interface{}{
			"pullRequest": map[string]interface{}{
				"reviews": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{
							"id":         nodeID,
							"databaseId": databaseID,
						},
					},
				},
			},
		},
	}
}

// addReactionMutationResponse is the GraphQL mutation response for addReaction.
var addReactionMutationResponse = map[string]interface{}{
	"addReaction": map[string]interface{}{
		"reaction": map[string]interface{}{"content": "EYES"},
	},
}

// ── IssueComment reactions ────────────────────────────────────────────────────

func (s *EdgeCaseSuite) TestAddReaction_IssueComment() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusCreated, reactionResponse)

	err := s.reactionService.AddToIssueComment(12345, "thumbsup")
	require.NoError(s.T(), err)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/issues/comments/12345/reactions")
}

func (s *EdgeCaseSuite) TestAddReaction_IssueComment_InvalidReaction() {
	err := s.reactionService.AddToIssueComment(12345, "notareaction")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid reaction")
	// rejected before any HTTP call
	assert.Equal(s.T(), 0, s.server.RequestCount())
}

func (s *EdgeCaseSuite) TestAddReaction_IssueComment_ServerError() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusInternalServerError,
		map[string]interface{}{"message": "Internal Server Error"})

	err := s.reactionService.AddToIssueComment(12345, "thumbsup")
	require.Error(s.T(), err)
}

// ── ReviewComment reactions ───────────────────────────────────────────────────

func (s *EdgeCaseSuite) TestAddReaction_ReviewComment() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/pulls/comments/", http.StatusCreated, reactionResponse)

	err := s.reactionService.AddToReviewComment(99999, "rocket")
	require.NoError(s.T(), err)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/pulls/comments/99999/reactions")
}

func (s *EdgeCaseSuite) TestAddReaction_ReviewComment_InvalidReaction() {
	err := s.reactionService.AddToReviewComment(99999, "🚀")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid reaction")
	assert.Equal(s.T(), 0, s.server.RequestCount())
}

// ── Review (body) reactions ───────────────────────────────────────────────────

func (s *EdgeCaseSuite) TestAddReaction_Review_Success() {
	const reviewID = int64(4001)
	const prNumber = 42
	const nodeID = "PR_kwDO_review_node_abc"

	// findReviewNodeID query then addReaction mutation
	s.server.OnGraphQL(reviewNodeQueryPattern, reviewNodeData(nodeID, reviewID))
	s.server.OnGraphQL("addReaction", addReactionMutationResponse)

	err := s.reactionService.AddToReview(prNumber, reviewID, "eyes")
	require.NoError(s.T(), err)

	// two GraphQL requests: one query (findReviewNodeID) + one mutation (addReaction)
	assert.Equal(s.T(), 2, s.server.RequestCount())
}

func (s *EdgeCaseSuite) TestAddReaction_Review_ReviewNotFound() {
	// node query returns no reviews — review ID does not match
	s.server.OnGraphQL(reviewNodeQueryPattern, map[string]interface{}{
		"repository": map[string]interface{}{
			"pullRequest": map[string]interface{}{
				"reviews": map[string]interface{}{"nodes": []interface{}{}},
			},
		},
	})

	err := s.reactionService.AddToReview(42, 9999, "eyes")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not found")
	// only the lookup query fires; mutation is never called
	assert.Equal(s.T(), 1, s.server.RequestCount())
}

func (s *EdgeCaseSuite) TestAddReaction_Review_InvalidReaction() {
	err := s.reactionService.AddToReview(42, 4001, "notareaction")
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "invalid reaction")
	// rejected before any HTTP/GraphQL call
	assert.Equal(s.T(), 0, s.server.RequestCount())
}

// ── Cross-type: all valid reactions accepted for each entity type ─────────────

func (s *EdgeCaseSuite) TestAddReaction_AllValidTypes() {
	validReactions := []string{"thumbsup", "thumbsdown", "laugh", "confused", "heart", "hooray", "rocket", "eyes"}
	for _, r := range validReactions {
		s.server.Reset()
		s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusCreated, reactionResponse)

		err := s.reactionService.AddToIssueComment(1, r)
		require.NoError(s.T(), err, "reaction %q should be valid", r)
	}
}

// TestAddReaction_OldPlusMinusOneRejected verifies that the old +1/-1 names are no longer accepted.
func (s *EdgeCaseSuite) TestAddReaction_OldPlusMinusOneRejected() {
	for _, old := range []string{"+1", "-1"} {
		err := s.reactionService.AddToIssueComment(1, old)
		require.Error(s.T(), err, "old reaction %q should be rejected", old)
		assert.Contains(s.T(), err.Error(), "invalid reaction")
	}
}
