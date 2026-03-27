package integration

import (
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadPRAndComments fetches PR metadata then comments for the same PR.
// Both use the mock server; responses are consistent.
func (s *EdgeCaseSuite) TestReadPRAndComments() {
	s.server.OnGraphQL(prQueryPattern, prData(42, "PR with Comments", "OPEN"))
	s.server.OnGraphQL(commentsQueryPattern, commentsQueryData(
		[]interface{}{issueCommentNode(1001, "alice", "Comment on PR 42", false)},
		nil, nil,
	))

	// fetch PR
	info, err := s.prService.Get(42)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "PR with Comments", info.Title)

	// fetch comments for the same PR
	comments, err := s.commentService.List(42)
	require.NoError(s.T(), err)
	require.Len(s.T(), comments.IssueComments, 1)
	assert.Equal(s.T(), "Comment on PR 42", comments.IssueComments[0].Body)

	// both made one GraphQL request each
	assert.Equal(s.T(), 2, s.server.RequestCount())
}

// TestReviewWithThreadDetails fetches a review then looks up a thread it references.
func (s *EdgeCaseSuite) TestReviewWithThreadDetails() {
	const reviewID = int64(3001)
	const threadID = int64(5001)

	// first call: review service fetches all reviews + threads
	s.server.OnGraphQL(allReviewsQueryPattern, allReviewsData(reviewID, []interface{}{
		buildReviewThreadNode(reviewID, false, []int64{threadID}),
	}))
	// second call: thread service fetches all threads to find threadID
	s.server.OnGraphQL(threadsQueryPattern, threadsQueryData("testviewer", []interface{}{
		threadNode("thread-node-1", false, []interface{}{
			threadCommentNode(threadID, reviewID, nil, "reviewer1", "Thread comment"),
		}),
	}))

	// get review detail — threadGroups[0].ThreadID should equal threadID
	detail, err := s.reviewService.Get(42, reviewID, false)
	require.NoError(s.T(), err)
	require.Len(s.T(), detail.ThreadGroups, 1)
	assert.Equal(s.T(), threadID, detail.ThreadGroups[0].ThreadID)

	// fetch full thread using the thread ID from the review
	t, err := s.threadService.Get(42, threadID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), threadID, t.ThreadID)
	require.Len(s.T(), t.Comments, 1)
	assert.Equal(s.T(), "Thread comment", t.Comments[0].Body)
}

// TestCommentThenReact posts a comment via REST then adds a reaction — verifying both calls land.
func (s *EdgeCaseSuite) TestCommentThenReact() {
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/", http.StatusCreated, map[string]interface{}{
		"id":         9001,
		"body":       "New comment",
		"created_at": "2024-01-15T15:00:00Z",
		"user":       map[string]interface{}{"login": "testviewer"},
	})
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/comments/", http.StatusCreated, reactionResponse)

	// post comment
	c, err := s.commentService.Post(42, "New comment")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(9001), c.DatabaseID)

	// react to that comment
	err = s.reactionService.AddToIssueComment(c.DatabaseID, "rocket")
	require.NoError(s.T(), err)

	// two REST requests, in order
	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 2)
	assert.Contains(s.T(), reqs[0].Path, "issues/42/comments")
	assert.Contains(s.T(), reqs[1].Path, "issues/comments/9001/reactions")
}

// TestReviewWorkflow reads a review, replies to the thread, then resolves it.
func (s *EdgeCaseSuite) TestReviewWorkflow() {
	const reviewID = int64(3001)
	const threadID = int64(5001)

	// 1. read review
	s.server.OnGraphQL(allReviewsQueryPattern, allReviewsData(reviewID, []interface{}{
		buildReviewThreadNode(reviewID, false, []int64{threadID}),
	}))
	// 2. reply to thread comment (REST)
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/pulls/", http.StatusCreated, map[string]interface{}{
		"id":         6001,
		"body":       "Fixed in abc123",
		"created_at": "2024-01-15T16:00:00Z",
		"user":       map[string]interface{}{"login": "testviewer"},
	})
	// 3. resolve thread (findThreadNodeID query + mutation)
	s.server.OnGraphQL(threadsQueryPattern, threadsQueryData("testviewer", []interface{}{
		threadNode("thread-node-resolve", false, []interface{}{
			threadCommentNode(threadID, reviewID, nil, "reviewer1", "comment"),
		}),
	}))
	s.server.OnGraphQL("resolveReviewThread", map[string]interface{}{
		"resolveReviewThread": map[string]interface{}{
			"thread": map[string]interface{}{"isResolved": true},
		},
	})

	// step 1: read review
	detail, err := s.reviewService.Get(42, reviewID, false)
	require.NoError(s.T(), err)
	require.Len(s.T(), detail.ThreadGroups, 1)

	// step 2: reply to the root comment of the thread
	reply, err := s.commentService.ReplyToReviewComment(42, threadID, "Fixed in abc123")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(6001), reply.DatabaseID)

	// step 3: resolve the thread
	err = s.threadService.Resolve(42, threadID)
	require.NoError(s.T(), err)

	// 4 total requests: 1 GraphQL (review) + 1 REST (reply) + 2 GraphQL (find+mutate)
	assert.Equal(s.T(), 4, s.server.RequestCount())
}
