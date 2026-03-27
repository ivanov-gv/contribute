//go:build integration

package integration

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetPR fetches PR #1 from the real GitHub API and compares the formatted
// output to the expected file — the same file used by the E2E binary test.
func (s *Suite) TestGetPR() {
	info, err := s.prService.Get(realPR)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "pr-description.md")
	assert.Equal(s.T(), expected, normalize(info.Format()))
}
