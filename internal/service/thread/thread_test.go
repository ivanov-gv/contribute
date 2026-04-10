package thread

import (
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	graphql_model "github.com/ivanov-gv/contribute/internal/model/graphql"
)

func TestBuildThread(t *testing.T) {
	t.Run("basic thread with replies", func(t *testing.T) {
		line := githubv4.Int(42)
		node := reviewThreadNode{
			IsOutdated: false,
			Path:       "main.go",
			Line:       &line,
			Comments: struct {
				Nodes []graphql_model.ThreadCommentNode
			}{
				Nodes: []graphql_model.ThreadCommentNode{
					{
						DatabaseID: 500,
						Body:       "Root comment",
						PullRequestReview: &struct {
							DatabaseID int64
						}{DatabaseID: 100},
					},
					{
						DatabaseID: 600,
						Body:       "Reply",
						ReplyTo: &struct {
							DatabaseID int64
						}{DatabaseID: 500},
						PullRequestReview: &struct {
							DatabaseID int64
						}{DatabaseID: 200},
					},
				},
			},
		}
		node.Comments.Nodes[0].Author.Login = "bob"
		node.Comments.Nodes[1].Author.Login = "alice"

		thread := buildThread(node, "alice", 500)

		assert.Equal(t, int64(500), thread.ThreadID)
		assert.Equal(t, "main.go", thread.Path)
		assert.Equal(t, 42, thread.Line)
		assert.Equal(t, "alice", thread.ViewerLogin)
		require.Len(t, thread.Comments, 2)

		// root comment
		assert.Equal(t, int64(500), thread.Comments[0].DatabaseID)
		assert.Equal(t, "bob", thread.Comments[0].Author)
		assert.Equal(t, int64(100), thread.Comments[0].ReviewDatabaseID)
		assert.Equal(t, int64(0), thread.Comments[0].ReplyToID)

		// reply
		assert.Equal(t, int64(600), thread.Comments[1].DatabaseID)
		assert.Equal(t, "alice", thread.Comments[1].Author)
		assert.Equal(t, int64(200), thread.Comments[1].ReviewDatabaseID)
		assert.Equal(t, int64(500), thread.Comments[1].ReplyToID)
	})

	t.Run("outdated thread with original lines", func(t *testing.T) {
		originalLine := githubv4.Int(88)
		originalStartLine := githubv4.Int(80)
		node := reviewThreadNode{
			IsOutdated:        true,
			Path:              "old.go",
			OriginalLine:      &originalLine,
			OriginalStartLine: &originalStartLine,
			Comments: struct {
				Nodes []graphql_model.ThreadCommentNode
			}{
				Nodes: []graphql_model.ThreadCommentNode{
					{DatabaseID: 500, Body: "Comment"},
				},
			},
		}

		thread := buildThread(node, "alice", 500)

		assert.True(t, thread.IsOutdated)
		assert.Equal(t, 88, thread.OriginalLine)
		assert.Equal(t, 80, thread.OriginalStartLine)
	})

	t.Run("nil line fields", func(t *testing.T) {
		node := reviewThreadNode{
			Path: "file.go",
			Comments: struct {
				Nodes []graphql_model.ThreadCommentNode
			}{
				Nodes: []graphql_model.ThreadCommentNode{
					{DatabaseID: 500},
				},
			},
		}

		thread := buildThread(node, "alice", 500)
		assert.Equal(t, 0, thread.Line)
		assert.Equal(t, 0, thread.StartLine)
		assert.Equal(t, 0, thread.OriginalLine)
		assert.Equal(t, 0, thread.OriginalStartLine)
	})

	t.Run("minimized comment preserved", func(t *testing.T) {
		node := reviewThreadNode{
			Path: "file.go",
			Comments: struct {
				Nodes []graphql_model.ThreadCommentNode
			}{
				Nodes: []graphql_model.ThreadCommentNode{
					{
						DatabaseID:      500,
						IsMinimized:     true,
						MinimizedReason: "OFF_TOPIC",
					},
				},
			},
		}

		thread := buildThread(node, "alice", 500)
		require.Len(t, thread.Comments, 1)
		assert.True(t, thread.Comments[0].IsMinimized)
		assert.Equal(t, "OFF_TOPIC", thread.Comments[0].MinimizedReason)
	})
}

func TestMapReactions(t *testing.T) {
	nodes := []graphql_model.ReactionNode{
		{Content: "THUMBS_UP"},
		{Content: "ROCKET"},
	}
	nodes[0].User.Login = "alice"
	nodes[1].User.Login = "bob"

	reactions := graphql_model.MapReactions(nodes)
	require.Len(t, reactions, 2)
	assert.Equal(t, "THUMBS_UP", reactions[0].Content)
	assert.Equal(t, "alice", reactions[0].Author)
	assert.Equal(t, "ROCKET", reactions[1].Content)
	assert.Equal(t, "bob", reactions[1].Author)
}
