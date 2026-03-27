package integration

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *EdgeCaseSuite) TestGetIssue_Success() {
	s.server.OnGraphQL(issueGetPattern, issueData(10, "Test Issue", "OPEN"))

	info, err := s.issueService.Get(10)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 10, info.Number)
	assert.Equal(s.T(), "Test Issue", info.Title)
	assert.Equal(s.T(), "OPEN", info.State)
	assert.Equal(s.T(), "issueauthor", info.Author)
	assert.Equal(s.T(), "Issue body", info.Body)
}

func (s *EdgeCaseSuite) TestGetIssue_WithLabelsAndAssignees() {
	s.server.OnGraphQL(issueGetPattern, map[string]interface{}{
		"repository": map[string]interface{}{
			"issue": map[string]interface{}{
				"number": 5,
				"title":  "Labeled Issue",
				"state":  "OPEN",
				"body":   "",
				"url":    "https://github.com/testowner/testrepo/issues/5",
				"author": map[string]interface{}{"login": "author1"},
				"labels": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{"name": "bug"},
						map[string]interface{}{"name": "help wanted"},
					},
				},
				"assignees": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{"login": "dev1"},
					},
				},
				"comments": map[string]interface{}{
					"totalCount": 1,
					"nodes": []interface{}{
						map[string]interface{}{
							"databaseId": 100,
							"author":     map[string]interface{}{"login": "commenter"},
							"body":       "Comment text",
							"createdAt":  "2024-01-15T10:00:00Z",
						},
					},
				},
				// CrossReferencedEvent inline fragment: fields appear at node level (no wrapper).
				// PullRequest inline fragment within source: fields appear at source level.
				"timelineItems": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{
							"source": map[string]interface{}{
								"number": 42,
								"title":  "Fix issue",
								"state":  "OPEN",
							},
						},
					},
				},
			},
		},
	})

	info, err := s.issueService.Get(5)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), []string{"bug", "help wanted"}, info.Labels)
	assert.Equal(s.T(), []string{"@dev1"}, info.Assignees)
	require.Len(s.T(), info.Comments, 1)
	assert.Equal(s.T(), "commenter", info.Comments[0].Author)
	require.Len(s.T(), info.LinkedPRs, 1)
	assert.Equal(s.T(), 42, info.LinkedPRs[0].Number)
}

func (s *EdgeCaseSuite) TestListIssues_Success() {
	s.server.OnGraphQL(issueListPattern, map[string]interface{}{
		"repository": map[string]interface{}{
			"issues": map[string]interface{}{
				"nodes": []interface{}{
					map[string]interface{}{
						"number": 3,
						"title":  "Issue 3",
						"state":  "OPEN",
						"author": map[string]interface{}{"login": "author1"},
						"labels": map[string]interface{}{
							"nodes": []interface{}{
								map[string]interface{}{"name": "bug"},
							},
						},
						"comments": map[string]interface{}{"totalCount": 2},
					},
					map[string]interface{}{
						"number":   1,
						"title":    "Issue 1",
						"state":    "OPEN",
						"author":   map[string]interface{}{"login": "author2"},
						"labels":   map[string]interface{}{"nodes": []interface{}{}},
						"comments": map[string]interface{}{"totalCount": 0},
					},
				},
			},
		},
	})

	items, err := s.issueService.List(10, nil)
	require.NoError(s.T(), err)
	require.Len(s.T(), items, 2)
	assert.Equal(s.T(), 3, items[0].Number)
	assert.Equal(s.T(), []string{"bug"}, items[0].Labels)
	assert.Equal(s.T(), 2, items[0].Comments)
	assert.Equal(s.T(), 1, items[1].Number)
}

func (s *EdgeCaseSuite) TestListIssues_Empty() {
	s.server.OnGraphQL(issueListPattern, map[string]interface{}{
		"repository": map[string]interface{}{
			"issues": map[string]interface{}{
				"nodes": []interface{}{},
			},
		},
	})

	items, err := s.issueService.List(10, nil)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), items)
	assert.Len(s.T(), items, 0)
}
