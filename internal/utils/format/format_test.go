package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReactionEmoji(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"THUMBS_UP", "👍"},
		{"THUMBS_DOWN", "👎"},
		{"LAUGH", "😄"},
		{"HOORAY", "🎉"},
		{"CONFUSED", "😕"},
		{"HEART", "❤️"},
		{"ROCKET", "🚀"},
		{"EYES", "👀"},
		{"UNKNOWN", "UNKNOWN"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, ReactionEmoji(tt.input))
		})
	}
}

func TestIsViewer(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		viewerLogin string
		expected    bool
	}{
		{"exact match", "alice", "alice", true},
		{"different users", "alice", "bob", false},
		{"bot suffix on viewer", "myapp", "myapp[bot]", true},
		{"bot suffix on login", "myapp[bot]", "myapp", true},
		{"both bot suffix", "myapp[bot]", "myapp[bot]", true},
		{"different with bot", "myapp[bot]", "other[bot]", false},
		{"empty strings", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsViewer(tt.login, tt.viewerLogin))
		})
	}
}

func TestAuthor(t *testing.T) {
	tests := []struct {
		name        string
		login       string
		viewerLogin string
		expected    string
	}{
		{"viewer", "alice", "alice", "you (@alice)"},
		{"other user", "bob", "alice", "@bob"},
		{"bot viewer", "myapp[bot]", "myapp", "you (@myapp[bot])"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Author(tt.login, tt.viewerLogin))
		})
	}
}

func TestDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-03-11T11:33:27Z", "2026-03-11 11:33:27"},
		{"2026-01-01T00:00:00Z", "2026-01-01 00:00:00"},
		{"2026-12-31T23:59:59Z", "2026-12-31 23:59:59"},
		// no Z suffix
		{"2026-03-11T11:33:27", "2026-03-11 11:33:27"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, Date(tt.input))
		})
	}
}

func TestEnumLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"OFF_TOPIC", "Off topic"},
		{"SPAM", "Spam"},
		{"OUTDATED", "Outdated"},
		{"RESOLVED", "Resolved"},
		{"CHANGES_REQUESTED", "Changes requested"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, EnumLabel(tt.input))
		})
	}
}

func TestReactions(t *testing.T) {
	t.Run("empty reactions", func(t *testing.T) {
		assert.Equal(t, "", Reactions(nil, "alice"))
		assert.Equal(t, "", Reactions([]Reaction{}, "alice"))
	})

	t.Run("single reaction by viewer", func(t *testing.T) {
		reactions := []Reaction{
			{Content: "THUMBS_UP", Author: "alice"},
		}
		result := Reactions(reactions, "alice")
		assert.Contains(t, result, "1 👍")
		assert.Contains(t, result, "reactions by you:")
		assert.Contains(t, result, "1 👍")
	})

	t.Run("reaction by other user", func(t *testing.T) {
		reactions := []Reaction{
			{Content: "ROCKET", Author: "bob"},
		}
		result := Reactions(reactions, "alice")
		assert.Contains(t, result, "1 🚀")
		assert.Contains(t, result, "reactions by you:")
		// should NOT contain viewer-specific count after "reactions by you:"
		assert.NotContains(t, result, "reactions by you: (")
	})

	t.Run("mixed reactions", func(t *testing.T) {
		reactions := []Reaction{
			{Content: "THUMBS_UP", Author: "alice"},
			{Content: "THUMBS_UP", Author: "bob"},
			{Content: "ROCKET", Author: "alice"},
		}
		result := Reactions(reactions, "alice")
		assert.Contains(t, result, "2 👍")
		assert.Contains(t, result, "1 🚀")
		assert.Contains(t, result, "reactions by you:")
	})

	t.Run("bot viewer match", func(t *testing.T) {
		reactions := []Reaction{
			{Content: "EYES", Author: "myapp[bot]"},
		}
		result := Reactions(reactions, "myapp")
		assert.Contains(t, result, "1 👀")
		assert.Contains(t, result, "reactions by you: (")
	})
}
