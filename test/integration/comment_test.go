//go:build integration

package integration

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListComments fetches all comments for PR #1 and compares the hidden-collapsed output.
func (s *Suite) TestListComments() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "comments.md")
	assert.Equal(s.T(), expected, normalize(result.Format(false)))
}

// TestListComments_ShowHidden fetches all comments for PR #1 and compares the fully-expanded output.
func (s *Suite) TestListComments_ShowHidden() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "comments-unhidden.md")
	assert.Equal(s.T(), expected, normalize(result.Format(true)))
}

// Single-comment tests: CLI always calls Format(true) for single-item lookup.

func (s *Suite) TestGetComment_4038597073() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4038597073)
	require.NotNil(s.T(), filtered, "comment 4038597073 not found")
	expected := readExpected(s.T(), "1-comments-4038597073.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4038597073_ShowHidden() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4038597073)
	require.NotNil(s.T(), filtered, "comment 4038597073 not found")
	expected := readExpected(s.T(), "1-comments-4038597073-unhidden.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4038819817() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4038819817)
	require.NotNil(s.T(), filtered, "comment 4038819817 not found")
	expected := readExpected(s.T(), "2-comments-4038819817.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4039142865() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4039142865)
	require.NotNil(s.T(), filtered, "comment 4039142865 not found")
	expected := readExpected(s.T(), "5-comments-4039142865.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4039221478() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4039221478)
	require.NotNil(s.T(), filtered, "comment 4039221478 not found")
	expected := readExpected(s.T(), "6-comments-4039221478.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4039593663() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4039593663)
	require.NotNil(s.T(), filtered, "comment 4039593663 not found")
	expected := readExpected(s.T(), "8-comments-4039593663.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4041153603() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4041153603)
	require.NotNil(s.T(), filtered, "comment 4041153603 not found")
	expected := readExpected(s.T(), "10-comments-4041153603.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4042410800() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4042410800)
	require.NotNil(s.T(), filtered, "comment 4042410800 not found")
	expected := readExpected(s.T(), "11-comments-4042410800.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}

func (s *Suite) TestGetComment_4067633036() {
	result, err := s.commentService.List(realPR)
	require.NoError(s.T(), err)

	filtered := result.FilterByID(4067633036)
	require.NotNil(s.T(), filtered, "comment 4067633036 not found")
	expected := readExpected(s.T(), "12-comments-4067633036.md")
	assert.Equal(s.T(), expected, normalize(filtered.Format(true)))
}
