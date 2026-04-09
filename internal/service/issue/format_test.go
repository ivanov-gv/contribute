package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfo_Format(t *testing.T) {
	t.Run("full issue", func(t *testing.T) {
		info := &Info{
			Number:       42,
			Title:        "Add dark mode",
			State:        "OPEN",
			Body:         "We need dark mode support.",
			URL:          "https://github.com/owner/repo/issues/42",
			Author:       "alice",
			Labels:       []string{"enhancement", "ui"},
			Assignees:    []string{"@bob"},
			CommentCount: 2,
			Comments: []Comment{
				{DatabaseID: 100, Author: "bob", Body: "I'll take this", CreatedAt: "2026-03-11T10:00:00Z"},
			},
			LinkedPRs: []LinkedPR{
				{Number: 50, Title: "Implement dark mode", State: "OPEN"},
			},
		}

		output := info.Format()

		assert.Contains(t, output, "# Add dark mode #42")
		assert.Contains(t, output, "open, by @alice")
		assert.Contains(t, output, "Labels: enhancement, ui")
		assert.Contains(t, output, "Assignees: @bob")
		assert.Contains(t, output, "Linked PRs: #50 Implement dark mode (open)")
		assert.Contains(t, output, "2 comments")
		assert.Contains(t, output, "We need dark mode support.")
		assert.Contains(t, output, "comment #100 by @bob")
		assert.Contains(t, output, "I'll take this")
	})

	t.Run("no body", func(t *testing.T) {
		info := &Info{
			Number: 1,
			Title:  "Bug",
			State:  "OPEN",
			URL:    "https://example.com",
			Author: "a",
		}
		output := info.Format()
		assert.Contains(t, output, "No description provided.")
	})

	t.Run("no linked PRs", func(t *testing.T) {
		info := &Info{
			Number: 1,
			Title:  "Bug",
			State:  "OPEN",
			URL:    "https://example.com",
			Author: "a",
		}
		output := info.Format()
		assert.NotContains(t, output, "Linked PRs:")
	})

	t.Run("singular comment", func(t *testing.T) {
		info := &Info{
			Number:       1,
			Title:        "Bug",
			State:        "OPEN",
			URL:          "https://example.com",
			Author:       "a",
			CommentCount: 1,
		}
		output := info.Format()
		assert.Contains(t, output, "1 comment")
		assert.NotContains(t, output, "1 comments")
	})
}

func TestFormatList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		assert.Equal(t, "No open issues found.\n", FormatList(nil))
	})

	t.Run("issues with labels and comments", func(t *testing.T) {
		items := []ListItem{
			{Number: 1, Title: "Bug fix", Author: "alice", Labels: []string{"bug"}, Comments: 3},
			{Number: 2, Title: "Feature", Author: "bob"},
		}

		output := FormatList(items)
		assert.Contains(t, output, "#1  Bug fix [bug]  by @alice (3 comments)")
		assert.Contains(t, output, "#2  Feature  by @bob")
	})
}
