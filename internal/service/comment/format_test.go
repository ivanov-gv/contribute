package comment

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ivanov-gv/contribute/internal/utils/format"
)

func TestCommentsResult_Format(t *testing.T) {
	t.Run("mixed timeline sorted by date", func(t *testing.T) {
		result := &CommentsResult{
			ViewerLogin: "alice",
			IssueComments: []IssueComment{
				{
					DatabaseID: 100,
					Author:     "alice",
					Body:       "First comment",
					CreatedAt:  "2026-03-11T10:00:00Z",
				},
				{
					DatabaseID: 300,
					Author:     "bob",
					Body:       "Third comment",
					CreatedAt:  "2026-03-11T12:00:00Z",
				},
			},
			Reviews: []Review{
				{
					DatabaseID:   200,
					Author:       "bob",
					Body:         "Looks good",
					State:        "APPROVED",
					CreatedAt:    "2026-03-11T11:00:00Z",
					CommentCount: 0,
				},
			},
		}

		output := result.Format(false)

		// verify ordering: issue100 → review200 → issue300
		pos100 := indexOf(output, "issue #100")
		pos200 := indexOf(output, "review #200")
		pos300 := indexOf(output, "issue #300")
		assert.Greater(t, pos200, pos100, "review should come after first comment")
		assert.Greater(t, pos300, pos200, "third comment should come after review")
	})

	t.Run("minimized issue comment shows only header", func(t *testing.T) {
		result := &CommentsResult{
			ViewerLogin: "alice",
			IssueComments: []IssueComment{
				{
					DatabaseID:      100,
					Author:          "spam",
					Body:            "Buy stuff!",
					CreatedAt:       "2026-03-11T10:00:00Z",
					IsMinimized:     true,
					MinimizedReason: "SPAM",
				},
			},
		}

		output := result.Format(false)
		assert.Contains(t, output, "hidden: Spam")
		assert.NotContains(t, output, "Buy stuff!")
	})

	t.Run("dismissed review shows as hidden", func(t *testing.T) {
		result := &CommentsResult{
			ViewerLogin: "alice",
			Reviews: []Review{
				{
					DatabaseID: 200,
					Author:     "bob",
					Body:       "Old review",
					State:      "DISMISSED",
					CreatedAt:  "2026-03-11T10:00:00Z",
				},
			},
		}

		output := result.Format(false)
		assert.Contains(t, output, "hidden: Dismissed")
		assert.NotContains(t, output, "Old review")
	})

	t.Run("hidden review via thread resolution shows as hidden", func(t *testing.T) {
		result := &CommentsResult{
			ViewerLogin: "alice",
			Reviews: []Review{
				{
					DatabaseID:   200,
					Author:       "bob",
					Body:         "Resolved review",
					State:        "COMMENTED",
					CreatedAt:    "2026-03-11T10:00:00Z",
					IsHidden:     true,
					HiddenReason: "Resolved",
				},
			},
		}

		output := result.Format(false)
		assert.Contains(t, output, "hidden: Resolved")
		assert.NotContains(t, output, "Resolved review")

		// with showHidden, body is visible
		outputShown := result.Format(true)
		assert.Contains(t, outputShown, "hidden: Resolved")
		assert.Contains(t, outputShown, "Resolved review")
	})

	t.Run("viewer identified correctly", func(t *testing.T) {
		result := &CommentsResult{
			ViewerLogin: "alice",
			IssueComments: []IssueComment{
				{
					DatabaseID: 100,
					Author:     "alice",
					Body:       "My comment",
					CreatedAt:  "2026-03-11T10:00:00Z",
				},
			},
		}

		output := result.Format(false)
		assert.Contains(t, output, "you (@alice)")
	})

	t.Run("review with inline comments shows count", func(t *testing.T) {
		result := &CommentsResult{
			ViewerLogin: "alice",
			Reviews: []Review{
				{
					DatabaseID:   200,
					Author:       "bob",
					Body:         "Needs changes",
					State:        "CHANGES_REQUESTED",
					CreatedAt:    "2026-03-11T10:00:00Z",
					CommentCount: 3,
				},
			},
		}

		output := result.Format(false)
		assert.Contains(t, output, "comments: 3")
	})

	t.Run("reactions displayed", func(t *testing.T) {
		result := &CommentsResult{
			ViewerLogin: "alice",
			IssueComments: []IssueComment{
				{
					DatabaseID: 100,
					Author:     "bob",
					Body:       "Nice",
					CreatedAt:  "2026-03-11T10:00:00Z",
					Reactions: []format.Reaction{
						{Content: "THUMBS_UP", Author: "alice"},
						{Content: "ROCKET", Author: "bob"},
					},
				},
			},
		}

		output := result.Format(false)
		assert.Contains(t, output, "👍")
		assert.Contains(t, output, "🚀")
		assert.Contains(t, output, "reactions by you:")
	})

	t.Run("empty result", func(t *testing.T) {
		result := &CommentsResult{ViewerLogin: "alice"}
		output := result.Format(false)
		assert.Equal(t, "", output)
	})
}

func TestCommentsResult_FilterByID(t *testing.T) {
	result := &CommentsResult{
		ViewerLogin: "alice",
		IssueComments: []IssueComment{
			{DatabaseID: 100, Author: "alice", Body: "Comment 1"},
			{DatabaseID: 200, Author: "bob", Body: "Comment 2"},
		},
		Reviews: []Review{
			{DatabaseID: 300, Author: "bob", Body: "Review 1"},
		},
	}

	t.Run("find issue comment", func(t *testing.T) {
		filtered := result.FilterByID(100)
		assert.NotNil(t, filtered)
		assert.Len(t, filtered.IssueComments, 1)
		assert.Equal(t, int64(100), filtered.IssueComments[0].DatabaseID)
		assert.Empty(t, filtered.Reviews)
	})

	t.Run("find review", func(t *testing.T) {
		filtered := result.FilterByID(300)
		assert.NotNil(t, filtered)
		assert.Empty(t, filtered.IssueComments)
		assert.Len(t, filtered.Reviews, 1)
		assert.Equal(t, int64(300), filtered.Reviews[0].DatabaseID)
	})

	t.Run("not found", func(t *testing.T) {
		filtered := result.FilterByID(999)
		assert.Nil(t, filtered)
	})
}

// indexOf returns the position of substr in s, or -1 if not found
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
