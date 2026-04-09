//go:build integration

package integration

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Suite) TestGetThread_2918002761() {
	t, err := s.threadService.Get(realPR, 2918002761)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "thread-2918002761.md")
	assert.Equal(s.T(), expected, normalize(t.Format(false)))
}

func (s *Suite) TestGetThread_2918002761_ShowHidden() {
	t, err := s.threadService.Get(realPR, 2918002761)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "thread-2918002761-unhidden.md")
	assert.Equal(s.T(), expected, normalize(t.Format(true)))
}

func (s *Suite) TestGetThread_2918006660() {
	t, err := s.threadService.Get(realPR, 2918006660)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "thread-2918006660.md")
	assert.Equal(s.T(), expected, normalize(t.Format(false)))
}

func (s *Suite) TestGetThread_2918006660_ShowHidden() {
	t, err := s.threadService.Get(realPR, 2918006660)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "thread-2918006660-unhidden.md")
	assert.Equal(s.T(), expected, normalize(t.Format(true)))
}
