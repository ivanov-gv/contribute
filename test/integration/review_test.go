//go:build integration

package integration

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Review test IDs and their expected files mirror the E2E test cases exactly.
// showDiff=false matches the CLI default (no --diff flag in E2E tests).

func (s *Suite) TestGetReview_3929204495() {
	detail, err := s.reviewService.Get(realPR, 3929204495, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "3-review-3929204495.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, false)))
}

func (s *Suite) TestGetReview_3929204495_ShowHidden() {
	detail, err := s.reviewService.Get(realPR, 3929204495, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "3-review-3929204495-unhidden.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, true)))
}

func (s *Suite) TestGetReview_3929240428() {
	detail, err := s.reviewService.Get(realPR, 3929240428, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "3-3.2.1-review-3929240428.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, false)))
}

func (s *Suite) TestGetReview_3929240428_ShowHidden() {
	detail, err := s.reviewService.Get(realPR, 3929240428, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "3-3.2.1-review-3929240428-unhidden.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, true)))
}

func (s *Suite) TestGetReview_3929353771() {
	detail, err := s.reviewService.Get(realPR, 3929353771, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "4-review-3929353771.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, false)))
}

func (s *Suite) TestGetReview_3929353771_ShowHidden() {
	detail, err := s.reviewService.Get(realPR, 3929353771, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "4-review-3929353771-unhidden.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, true)))
}

func (s *Suite) TestGetReview_3929758963() {
	detail, err := s.reviewService.Get(realPR, 3929758963, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "7-review-3929758963.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, false)))
}

func (s *Suite) TestGetReview_3930039277() {
	detail, err := s.reviewService.Get(realPR, 3930039277, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "9-review-3930039277.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, false)))
}

func (s *Suite) TestGetReview_3930039277_ShowHidden() {
	detail, err := s.reviewService.Get(realPR, 3930039277, false)
	require.NoError(s.T(), err)

	expected := readExpected(s.T(), "9-review-3930039277-unhidden.md")
	assert.Equal(s.T(), expected, normalize(detail.Format(false, true)))
}
