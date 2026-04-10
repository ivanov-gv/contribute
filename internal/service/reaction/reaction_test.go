package reaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	// all valid reactions
	for _, r := range ValidReactions {
		t.Run("valid_"+r, func(t *testing.T) {
			assert.True(t, isValid(r))
		})
	}

	// invalid reactions — includes old +1/-1 which are no longer accepted
	invalid := []string{"", "+1", "-1", "ROCKET", "smile", "🚀", "invalid"}
	for _, r := range invalid {
		name := r
		if name == "" {
			name = "empty"
		}
		t.Run("invalid_"+name, func(t *testing.T) {
			assert.False(t, isValid(r))
		})
	}
}

func TestToRESTContent(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"thumbsup", "+1"},
		{"thumbsdown", "-1"},
		{"laugh", "laugh"},
		{"confused", "confused"},
		{"heart", "heart"},
		{"hooray", "hooray"},
		{"rocket", "rocket"},
		{"eyes", "eyes"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, toRESTContent(tc.input))
		})
	}
}

func TestToGraphQLContent(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"thumbsup", "THUMBS_UP"},
		{"thumbsdown", "THUMBS_DOWN"},
		{"laugh", "LAUGH"},
		{"confused", "CONFUSED"},
		{"heart", "HEART"},
		{"hooray", "HOORAY"},
		{"rocket", "ROCKET"},
		{"eyes", "EYES"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(toGraphQLContent(tc.input)))
		})
	}
}
