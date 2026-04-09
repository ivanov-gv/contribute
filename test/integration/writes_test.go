//go:build !integration

package integration

import (
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Write operation tests: POST comment, reply, inline comment, submit review,
// resolve/unresolve thread. These use the mock server because PR #1 is locked
// and write operations must not reach the real GitHub API.

func (s *EdgeCaseSuite) TestPostComment_Success() {
	restResp := map[string]interface{}{
		"id":         5001,
		"body":       "Hello world",
		"created_at": "2024-01-15T12:00:00Z",
		"user":       map[string]interface{}{"login": "testviewer"},
	}
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/issues/", http.StatusCreated, restResp)

	c, err := s.commentService.Post(42, "Hello world")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(5001), c.DatabaseID)
	assert.Equal(s.T(), "Hello world", c.Body)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/issues/42/comments")
}

func (s *EdgeCaseSuite) TestReplyToReviewComment_Success() {
	// go-github CreateCommentInReplyTo uses POST /repos/{owner}/{repo}/pulls/{number}/comments
	// with in_reply_to in the body (not a /replies sub-path)
	restResp := map[string]interface{}{
		"id":         6001,
		"body":       "Reply text",
		"created_at": "2024-01-15T12:30:00Z",
		"user":       map[string]interface{}{"login": "testviewer"},
	}
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/pulls/", http.StatusCreated, restResp)

	c, err := s.commentService.ReplyToReviewComment(42, 5001, "Reply text")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(6001), c.DatabaseID)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/pulls/42/comments")
}

func (s *EdgeCaseSuite) TestPostInlineComment_Success() {
	restResp := map[string]interface{}{
		"id":         7001,
		"body":       "Inline comment",
		"created_at": "2024-01-15T13:00:00Z",
		"user":       map[string]interface{}{"login": "testviewer"},
	}
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/pulls/", http.StatusCreated, restResp)

	c, err := s.commentService.PostInlineComment(42, "abc123", "main.go", "Inline comment", 10, "RIGHT")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(7001), c.DatabaseID)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/pulls/42/comments")
}

func (s *EdgeCaseSuite) TestSubmitReview_Success() {
	restResp := map[string]interface{}{
		"id":           8001,
		"body":         "LGTM",
		"state":        "APPROVED",
		"user":         map[string]interface{}{"login": "testviewer"},
		"submitted_at": "2024-01-15T14:00:00Z",
	}
	s.server.OnREST(http.MethodPost, "/repos/testowner/testrepo/pulls/", http.StatusOK, restResp)

	id, err := s.commentService.SubmitReview(42, "APPROVE", "LGTM")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), int64(8001), id)

	reqs := s.server.Requests()
	require.Len(s.T(), reqs, 1)
	assert.Contains(s.T(), reqs[0].Path, "/repos/testowner/testrepo/pulls/42/reviews")
}

func (s *EdgeCaseSuite) TestResolveThread_Success() {
	const threadID = int64(5001)
	// findThreadNodeID makes one threadsQuery, then Mutate sends a mutation
	s.server.OnGraphQL(threadsQueryPattern, threadsQueryData("testviewer", []interface{}{
		threadNode("thread-node-abc", false, []interface{}{
			threadCommentNode(threadID, 3001, nil, "alice", "comment"),
		}),
	}))
	s.server.OnGraphQL("resolveReviewThread", map[string]interface{}{
		"resolveReviewThread": map[string]interface{}{
			"thread": map[string]interface{}{"isResolved": true},
		},
	})

	err := s.threadService.Resolve(42, threadID)
	require.NoError(s.T(), err)

	// two GraphQL requests: one query (findThreadNodeID) + one mutation
	assert.Equal(s.T(), 2, s.server.RequestCount())
}

func (s *EdgeCaseSuite) TestUnresolveThread_Success() {
	const threadID = int64(5001)
	s.server.OnGraphQL(threadsQueryPattern, threadsQueryData("testviewer", []interface{}{
		threadNode("thread-node-xyz", true, []interface{}{
			threadCommentNode(threadID, 3001, nil, "alice", "comment"),
		}),
	}))
	s.server.OnGraphQL("unresolveReviewThread", map[string]interface{}{
		"unresolveReviewThread": map[string]interface{}{
			"thread": map[string]interface{}{"isResolved": false},
		},
	})

	err := s.threadService.Unresolve(42, threadID)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, s.server.RequestCount())
}

func (s *EdgeCaseSuite) TestResolveThread_ThreadNotFound() {
	s.server.OnGraphQL(threadsQueryPattern, threadsQueryData("testviewer", []interface{}{}))

	err := s.threadService.Resolve(42, 9999)
	require.Error(s.T(), err)
	assert.Contains(s.T(), err.Error(), "not found")
}
