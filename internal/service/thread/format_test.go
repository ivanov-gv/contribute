package thread

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ivanov-gv/gh-contribute/internal/utils/format"
)

func TestThread_Format(t *testing.T) {
	t.Run("basic thread with comments", func(t *testing.T) {
		thread := &Thread{
			ThreadID:    500,
			Path:        "main.go",
			Line:        42,
			ViewerLogin: "alice",
			Comments: []ThreadComment{
				{
					DatabaseID:       500,
					Author:           "bob",
					Body:             "Fix this line",
					CreatedAt:        "2026-03-14T11:18:37Z",
					ReviewDatabaseID: 100,
				},
				{
					DatabaseID:       600,
					Author:           "alice",
					Body:             "Fixed in abc123",
					CreatedAt:        "2026-03-14T12:37:24Z",
					ReviewDatabaseID: 200,
					ReplyToID:        500,
				},
			},
		}

		output := thread.Format()

		assert.Contains(t, output, "# thread #500")
		assert.Contains(t, output, "main.go on line +42")
		assert.Contains(t, output, "comment #500 by @bob  review #100")
		assert.Contains(t, output, "Fix this line")
		assert.Contains(t, output, "reply #600 to #500  by you (@alice)  review #200")
		assert.Contains(t, output, "Fixed in abc123")
		assert.Contains(t, output, "---")
	})

	t.Run("outdated thread", func(t *testing.T) {
		thread := &Thread{
			ThreadID:          500,
			Path:              "old.go",
			IsOutdated:        true,
			OriginalLine:      88,
			OriginalStartLine: 80,
			ViewerLogin:       "alice",
			Comments: []ThreadComment{
				{
					DatabaseID:       500,
					Author:           "bob",
					Body:             "Comment",
					CreatedAt:        "2026-03-14T11:18:37Z",
					ReviewDatabaseID: 100,
				},
			},
		}

		output := thread.Format()
		assert.Contains(t, output, "old.go on original lines 80 to 88 (outdated)")
	})

	t.Run("minimized comment", func(t *testing.T) {
		thread := &Thread{
			ThreadID:    500,
			Path:        "main.go",
			ViewerLogin: "alice",
			Comments: []ThreadComment{
				{
					DatabaseID:       500,
					Author:           "spam",
					Body:             "Buy stuff",
					CreatedAt:        "2026-03-14T11:18:37Z",
					ReviewDatabaseID: 100,
					IsMinimized:      true,
					MinimizedReason:  "OFF_TOPIC",
				},
			},
		}

		output := thread.Format()
		assert.Contains(t, output, "hidden: Off topic")
		assert.NotContains(t, output, "Buy stuff")
	})

	t.Run("reactions on comment", func(t *testing.T) {
		thread := &Thread{
			ThreadID:    500,
			Path:        "main.go",
			ViewerLogin: "alice",
			Comments: []ThreadComment{
				{
					DatabaseID:       500,
					Author:           "bob",
					Body:             "Nice",
					CreatedAt:        "2026-03-14T11:18:37Z",
					ReviewDatabaseID: 100,
					Reactions: []format.Reaction{
						{Content: "ROCKET", Author: "alice"},
					},
				},
			},
		}

		output := thread.Format()
		assert.Contains(t, output, "🚀")
		assert.Contains(t, output, "reactions by you:")
	})

	t.Run("empty thread", func(t *testing.T) {
		thread := &Thread{
			ThreadID:    500,
			Path:        "main.go",
			ViewerLogin: "alice",
		}

		output := thread.Format()
		assert.Contains(t, output, "# thread #500")
	})
}

func TestFormatLocation(t *testing.T) {
	tests := []struct {
		name     string
		thread   Thread
		expected string
	}{
		{
			name:     "empty path",
			thread:   Thread{},
			expected: "",
		},
		{
			name:     "current single line",
			thread:   Thread{Path: "main.go", Line: 42},
			expected: "main.go on line +42",
		},
		{
			name:     "current line range",
			thread:   Thread{Path: "main.go", Line: 50, StartLine: 40},
			expected: "main.go on lines +40 to +50",
		},
		{
			name:     "current same start and end",
			thread:   Thread{Path: "main.go", Line: 42, StartLine: 42},
			expected: "main.go on line +42",
		},
		{
			name:     "outdated single line",
			thread:   Thread{Path: "main.go", IsOutdated: true, OriginalLine: 22},
			expected: "main.go on original line 22 (outdated)",
		},
		{
			name:     "outdated line range",
			thread:   Thread{Path: "main.go", IsOutdated: true, OriginalLine: 30, OriginalStartLine: 20},
			expected: "main.go on original lines 20 to 30 (outdated)",
		},
		{
			name:     "outdated no line info",
			thread:   Thread{Path: "main.go", IsOutdated: true},
			expected: "main.go (outdated)",
		},
		{
			name:     "current no line info",
			thread:   Thread{Path: "main.go"},
			expected: "main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatLocation(&tt.thread))
		})
	}
}
