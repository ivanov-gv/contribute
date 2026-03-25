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

	// invalid reactions
	invalid := []string{"", "thumbsup", "ROCKET", "smile", "🚀", "invalid"}
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
