package review

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ivanov-gv/gh-contribute/internal/utils/format"
)

func TestReviewDetail_Format(t *testing.T) {
	t.Run("basic review with body", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "bob",
			Body:        "Needs changes",
			State:       "CHANGES_REQUESTED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
		}

		output := detail.Format(false)

		assert.Contains(t, output, "# review #100 by @bob")
		assert.Contains(t, output, "2026-03-14 11:13:03")
		assert.Contains(t, output, "Needs changes")
	})

	t.Run("review by viewer", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "alice",
			Body:        "",
			State:       "APPROVED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
		}

		output := detail.Format(false)
		assert.Contains(t, output, "you (@alice)")
	})

	t.Run("thread groups included", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "bob",
			Body:        "",
			State:       "COMMENTED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
			ThreadGroups: []ReviewThreadGroup{
				{
					ThreadID: 500,
					Path:     "main.go",
					Line:     42,
					Comments: []ReviewComment{
						{
							DatabaseID: 500,
							Author:     "bob",
							Body:       "Fix this",
							CreatedAt:  "2026-03-14T11:13:03Z",
						},
					},
				},
			},
		}

		output := detail.Format(false)
		assert.Contains(t, output, "thread #500")
		assert.Contains(t, output, "main.go on line +42")
		assert.Contains(t, output, "comment #500 by @bob")
		assert.Contains(t, output, "Fix this")
	})

	t.Run("diff hunk shown when requested", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "bob",
			Body:        "",
			State:       "COMMENTED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
			ThreadGroups: []ReviewThreadGroup{
				{
					ThreadID: 500,
					Path:     "main.go",
					Line:     10,
					DiffHunk: "@@ -1,5 +1,5 @@\n-old\n+new",
					Comments: []ReviewComment{
						{
							DatabaseID: 500,
							Author:     "bob",
							Body:       "Why this change?",
							CreatedAt:  "2026-03-14T11:13:03Z",
						},
					},
				},
			},
		}

		output := detail.Format(true)
		assert.Contains(t, output, "```diff")
		assert.Contains(t, output, "-old")
		assert.Contains(t, output, "+new")
	})

	t.Run("diff hunk hidden when not requested", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "bob",
			Body:        "",
			State:       "COMMENTED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
			ThreadGroups: []ReviewThreadGroup{
				{
					ThreadID: 500,
					Path:     "main.go",
					DiffHunk: "@@ -1,5 +1,5 @@\n-old\n+new",
					Comments: []ReviewComment{
						{DatabaseID: 500, Author: "bob", Body: "Test", CreatedAt: "2026-03-14T11:13:03Z"},
					},
				},
			},
		}

		output := detail.Format(false)
		assert.NotContains(t, output, "```diff")
	})

	t.Run("external reply flagged", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "bob",
			Body:        "",
			State:       "COMMENTED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
			ThreadGroups: []ReviewThreadGroup{
				{
					ThreadID: 500,
					Path:     "main.go",
					Comments: []ReviewComment{
						{
							DatabaseID:        600,
							Author:            "bob",
							Body:              "Reply to external",
							CreatedAt:         "2026-03-14T12:00:00Z",
							ReplyToID:         400,
							ReplyToIsExternal: true,
						},
					},
				},
			},
		}

		output := detail.Format(false)
		assert.Contains(t, output, "not in this review")
		assert.Contains(t, output, "reply #600 to #400")
	})

	t.Run("minimized comment", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "bob",
			Body:        "",
			State:       "COMMENTED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
			ThreadGroups: []ReviewThreadGroup{
				{
					ThreadID: 500,
					Path:     "main.go",
					Comments: []ReviewComment{
						{
							DatabaseID:      500,
							Author:          "spam",
							Body:            "Spam content",
							CreatedAt:       "2026-03-14T11:13:03Z",
							IsMinimized:     true,
							MinimizedReason: "SPAM",
						},
					},
				},
			},
		}

		output := detail.Format(false)
		assert.Contains(t, output, "hidden: Spam")
		assert.NotContains(t, output, "Spam content")
	})

	t.Run("reactions on review", func(t *testing.T) {
		detail := &ReviewDetail{
			DatabaseID:  100,
			Author:      "bob",
			Body:        "LGTM",
			State:       "APPROVED",
			CreatedAt:   "2026-03-14T11:13:03Z",
			ViewerLogin: "alice",
			Reactions: []format.Reaction{
				{Content: "THUMBS_UP", Author: "alice"},
			},
		}

		output := detail.Format(false)
		assert.Contains(t, output, "👍")
		assert.Contains(t, output, "reactions by you:")
	})
}

func TestFormatLocation(t *testing.T) {
	tests := []struct {
		name     string
		group    ReviewThreadGroup
		expected string
	}{
		{
			name:     "empty path",
			group:    ReviewThreadGroup{},
			expected: "",
		},
		{
			name:     "current single line",
			group:    ReviewThreadGroup{Path: "main.go", Line: 42},
			expected: "main.go on line +42",
		},
		{
			name:     "current line range",
			group:    ReviewThreadGroup{Path: "main.go", Line: 50, StartLine: 40},
			expected: "main.go on lines +40 to +50",
		},
		{
			name:     "current same start and end",
			group:    ReviewThreadGroup{Path: "main.go", Line: 42, StartLine: 42},
			expected: "main.go on line +42",
		},
		{
			name:     "outdated single line",
			group:    ReviewThreadGroup{Path: "main.go", IsOutdated: true, OriginalLine: 22},
			expected: "main.go on original line 22 (outdated)",
		},
		{
			name:     "outdated line range",
			group:    ReviewThreadGroup{Path: "main.go", IsOutdated: true, OriginalLine: 30, OriginalStartLine: 20},
			expected: "main.go on original lines 20 to 30 (outdated)",
		},
		{
			name:     "outdated no line info",
			group:    ReviewThreadGroup{Path: "main.go", IsOutdated: true},
			expected: "main.go (outdated)",
		},
		{
			name:     "current no line info",
			group:    ReviewThreadGroup{Path: "main.go"},
			expected: "main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatLocation(tt.group))
		})
	}
}
