package review

import (
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	graphql_model "github.com/ivanov-gv/contribute/internal/model/graphql"
)

func TestCollectGroupsNoDiff(t *testing.T) {
	t.Run("filters by review ID", func(t *testing.T) {
		nodes := []reviewThreadNodeNoDiff{
			{
				Path: "a.go",
				Comments: struct {
					Nodes []graphql_model.ThreadCommentNode
				}{
					Nodes: []graphql_model.ThreadCommentNode{
						{DatabaseID: 1, PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 100}},
						{DatabaseID: 2, PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 200}},
					},
				},
			},
		}

		groups := collectGroupsNoDiff(nodes, 100)
		require.Len(t, groups, 1)
		assert.Len(t, groups[0].Comments, 1)
		assert.Equal(t, int64(1), groups[0].Comments[0].DatabaseID)
	})

	t.Run("skips threads with no matching comments", func(t *testing.T) {
		nodes := []reviewThreadNodeNoDiff{
			{
				Path: "a.go",
				Comments: struct {
					Nodes []graphql_model.ThreadCommentNode
				}{
					Nodes: []graphql_model.ThreadCommentNode{
						{DatabaseID: 1, PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 200}},
					},
				},
			},
		}

		groups := collectGroupsNoDiff(nodes, 100)
		assert.Empty(t, groups)
	})

	t.Run("thread ID from first comment", func(t *testing.T) {
		nodes := []reviewThreadNodeNoDiff{
			{
				Path: "a.go",
				Comments: struct {
					Nodes []graphql_model.ThreadCommentNode
				}{
					Nodes: []graphql_model.ThreadCommentNode{
						{DatabaseID: 10, PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 200}},
						{DatabaseID: 20, PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 100}},
					},
				},
			},
		}

		groups := collectGroupsNoDiff(nodes, 100)
		require.Len(t, groups, 1)
		// ThreadID should be the first comment in the full thread (10), not the first matching comment
		assert.Equal(t, int64(10), groups[0].ThreadID)
	})
}

func TestCollectGroupsWithDiff(t *testing.T) {
	t.Run("filters by review ID and populates diff hunk", func(t *testing.T) {
		nodes := []reviewThreadNodeWithDiff{
			{
				Path: "b.go",
				Comments: struct {
					Nodes []threadCommentNodeWithDiff
				}{
					Nodes: []threadCommentNodeWithDiff{
						{
							ThreadCommentNode: graphql_model.ThreadCommentNode{
								DatabaseID:        1,
								PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 100},
							},
							DiffHunk: "@@ -1,3 +1,4 @@\n+added line",
						},
						{
							ThreadCommentNode: graphql_model.ThreadCommentNode{
								DatabaseID:        2,
								PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 100},
							},
							DiffHunk: "@@ -1,3 +1,4 @@\n+added line",
						},
					},
				},
			},
		}

		groups := collectGroupsWithDiff(nodes, 100)
		require.Len(t, groups, 1)
		assert.Len(t, groups[0].Comments, 2)
		assert.Equal(t, "@@ -1,3 +1,4 @@\n+added line", groups[0].DiffHunk)
	})

	t.Run("skips threads with no matching comments", func(t *testing.T) {
		nodes := []reviewThreadNodeWithDiff{
			{
				Path: "b.go",
				Comments: struct {
					Nodes []threadCommentNodeWithDiff
				}{
					Nodes: []threadCommentNodeWithDiff{
						{
							ThreadCommentNode: graphql_model.ThreadCommentNode{
								DatabaseID:        1,
								PullRequestReview: &struct{ DatabaseID int64 }{DatabaseID: 200},
							},
						},
					},
				},
			},
		}

		groups := collectGroupsWithDiff(nodes, 100)
		assert.Empty(t, groups)
	})
}

func TestSortedGroups(t *testing.T) {
	groups := []ReviewThreadGroup{
		{
			ThreadID: 1,
			Comments: []ReviewComment{
				{DatabaseID: 1, CreatedAt: "2026-03-14T12:00:00Z"},
			},
		},
		{
			ThreadID: 2,
			Comments: []ReviewComment{
				{DatabaseID: 2, CreatedAt: "2026-03-14T10:00:00Z"},
			},
		},
		{
			ThreadID: 3,
			Comments: []ReviewComment{
				{DatabaseID: 3, CreatedAt: "2026-03-14T11:00:00Z"},
			},
		},
	}

	sorted := sortedGroups(groups)
	require.Len(t, sorted, 3)
	assert.Equal(t, int64(2), sorted[0].ThreadID)
	assert.Equal(t, int64(3), sorted[1].ThreadID)
	assert.Equal(t, int64(1), sorted[2].ThreadID)
}

func TestMapReviewComment(t *testing.T) {
	t.Run("thread root", func(t *testing.T) {
		rc := fromReviewCommentNode(
			100,
			githubv4.String("alice"),
			githubv4.String("Comment body"),
			githubv4.DateTime{},
			githubv4.Boolean(false),
			githubv4.String(""),
			nil,
			nil,
			map[int64]struct{}{100: {}},
		)

		assert.Equal(t, int64(100), rc.DatabaseID)
		assert.Equal(t, "alice", rc.Author)
		assert.Equal(t, "Comment body", rc.Body)
		assert.Equal(t, int64(0), rc.ReplyToID)
		assert.False(t, rc.ReplyToIsExternal)
	})

	t.Run("reply within same review", func(t *testing.T) {
		reviewIDs := map[int64]struct{}{100: {}, 200: {}}
		rc := fromReviewCommentNode(
			200,
			githubv4.String("bob"),
			githubv4.String("Reply"),
			githubv4.DateTime{},
			githubv4.Boolean(false),
			githubv4.String(""),
			&struct{ DatabaseID int64 }{DatabaseID: 100},
			nil,
			reviewIDs,
		)

		assert.Equal(t, int64(100), rc.ReplyToID)
		assert.False(t, rc.ReplyToIsExternal)
	})

	t.Run("reply to external comment", func(t *testing.T) {
		reviewIDs := map[int64]struct{}{200: {}}
		rc := fromReviewCommentNode(
			200,
			githubv4.String("bob"),
			githubv4.String("Reply to external"),
			githubv4.DateTime{},
			githubv4.Boolean(false),
			githubv4.String(""),
			&struct{ DatabaseID int64 }{DatabaseID: 100},
			nil,
			reviewIDs,
		)

		assert.Equal(t, int64(100), rc.ReplyToID)
		assert.True(t, rc.ReplyToIsExternal)
	})

	t.Run("minimized comment", func(t *testing.T) {
		rc := fromReviewCommentNode(
			100,
			githubv4.String("spam"),
			githubv4.String("Spam"),
			githubv4.DateTime{},
			githubv4.Boolean(true),
			githubv4.String("SPAM"),
			nil,
			nil,
			nil,
		)

		assert.True(t, rc.IsMinimized)
		assert.Equal(t, "SPAM", rc.MinimizedReason)
	})

	t.Run("reactions mapped", func(t *testing.T) {
		reactions := []graphql_model.ReactionNode{
			{Content: "THUMBS_UP"},
			{Content: "ROCKET"},
		}
		reactions[0].User.Login = "alice"
		reactions[1].User.Login = "bob"

		rc := fromReviewCommentNode(
			100,
			githubv4.String("bob"),
			githubv4.String("Text"),
			githubv4.DateTime{},
			githubv4.Boolean(false),
			githubv4.String(""),
			nil,
			reactions,
			nil,
		)

		require.Len(t, rc.Reactions, 2)
		assert.Equal(t, "THUMBS_UP", rc.Reactions[0].Content)
		assert.Equal(t, "alice", rc.Reactions[0].Author)
	})
}

func TestFindReviewMeta(t *testing.T) {
	nodes := []reviewMetaNode{
		{DatabaseID: 100},
		{DatabaseID: 200},
		{DatabaseID: 300},
	}

	t.Run("found", func(t *testing.T) {
		meta := findReviewMeta(nodes, 200)
		require.NotNil(t, meta)
		assert.Equal(t, int64(200), meta.DatabaseID)
	})

	t.Run("not found", func(t *testing.T) {
		meta := findReviewMeta(nodes, 999)
		assert.Nil(t, meta)
	})

	t.Run("empty nodes", func(t *testing.T) {
		meta := findReviewMeta(nil, 100)
		assert.Nil(t, meta)
	})
}
