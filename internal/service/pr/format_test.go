package pr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormat(t *testing.T) {
	t.Run("standard PR", func(t *testing.T) {
		info := &Info{
			Number:       42,
			Title:        "Add feature X",
			State:        "open",
			IsDraft:      false,
			Mergeable:    "MERGEABLE",
			Body:         "This PR adds feature X.",
			URL:          "https://github.com/owner/repo/pull/42",
			Head:         "feature-x",
			Base:         "main",
			Author:       "alice",
			CommitCount:  3,
			CommentCount: 5,
			Reviewers:    []string{"@bob"},
			Assignees:    []string{"@alice"},
			Labels:       []string{"enhancement"},
			Projects:     []string{"Board"},
			Milestone:    "v1.0",
			Issues:       []LinkedIssue{{Number: 10, Title: "Feature request"}},
		}

		output := info.Format()

		assert.Contains(t, output, "# Add feature X #42")
		assert.Contains(t, output, "open, by @alice, 3 commits")
		assert.Contains(t, output, "`feature-x` -> `main`")
		assert.Contains(t, output, "no merge conflict")
		assert.Contains(t, output, "https://github.com/owner/repo/pull/42")
		assert.Contains(t, output, "Reviewers: @bob")
		assert.Contains(t, output, "Assignees: @alice")
		assert.Contains(t, output, "Labels: enhancement")
		assert.Contains(t, output, "Projects: Board")
		assert.Contains(t, output, "Milestone: v1.0")
		assert.Contains(t, output, "#10 Feature request")
		assert.Contains(t, output, "5 comments")
		assert.Contains(t, output, "This PR adds feature X.")
	})

	t.Run("draft PR with conflict", func(t *testing.T) {
		info := &Info{
			Number:       1,
			Title:        "WIP",
			State:        "open",
			IsDraft:      true,
			Mergeable:    "CONFLICTING",
			Body:         "",
			URL:          "https://github.com/owner/repo/pull/1",
			Head:         "wip",
			Base:         "main",
			Author:       "alice",
			CommitCount:  1,
			CommentCount: 1,
		}

		output := info.Format()

		assert.Contains(t, output, "draft, by @alice, 1 commit")
		assert.Contains(t, output, "merge conflict")
		assert.Contains(t, output, "1 comment")
		assert.Contains(t, output, "No description provided.")
	})

	t.Run("unknown merge status", func(t *testing.T) {
		info := &Info{
			Number:    1,
			Title:     "Test",
			State:     "open",
			Mergeable: "UNKNOWN",
			URL:       "https://example.com",
			Author:    "a",
		}

		output := info.Format()
		assert.Contains(t, output, "merge status unknown")
	})

	t.Run("single commit uses singular", func(t *testing.T) {
		info := &Info{
			Number:      1,
			Title:       "Test",
			State:       "open",
			CommitCount: 1,
			URL:         "https://example.com",
			Author:      "a",
		}
		output := info.Format()
		assert.Contains(t, output, "1 commit ")
		assert.NotContains(t, output, "1 commits")
	})
}
