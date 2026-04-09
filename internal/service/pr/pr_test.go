package pr

import (
	"net/url"
	"testing"

	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
)

func TestMapPR(t *testing.T) {
	t.Run("all fields populated", func(t *testing.T) {
		milestone := &prMilestone{Title: "v1.0"}
		node := &prNode{
			Number:      githubv4.Int(42),
			Title:       githubv4.String("Add feature X"),
			State:       githubv4.String("OPEN"),
			IsDraft:     githubv4.Boolean(false),
			Mergeable:   githubv4.String("MERGEABLE"),
			Body:        githubv4.String("Description"),
			HeadRefName: githubv4.String("feature-x"),
			BaseRefName: githubv4.String("main"),
			Milestone:   milestone,
		}
		node.Author.Login = "alice"
		node.Commits.TotalCount = 3
		node.TotalCommentsCount = 3
		node.Assignees.Nodes = []struct{ Login githubv4.String }{
			{Login: "alice"},
		}
		node.Labels.Nodes = []struct{ Name githubv4.String }{
			{Name: "enhancement"},
		}
		node.ProjectsV2.Nodes = []struct{ Title githubv4.String }{
			{Title: "Board"},
		}
		node.ClosingIssuesReferences.Nodes = []struct {
			Number githubv4.Int
			Title  githubv4.String
		}{
			{Number: 10, Title: "Feature request"},
		}

		// set URL
		parsedURL, _ := url.Parse("https://github.com/owner/repo/pull/42")
		node.URL = githubv4.URI{URL: parsedURL}

		info := fromPRNode(node)

		assert.Equal(t, 42, info.Number)
		assert.Equal(t, "Add feature X", info.Title)
		assert.Equal(t, "open", info.State)
		assert.False(t, info.IsDraft)
		assert.Equal(t, "MERGEABLE", info.Mergeable)
		assert.Equal(t, "Description", info.Body)
		assert.Equal(t, "feature-x", info.Head)
		assert.Equal(t, "main", info.Base)
		assert.Equal(t, "alice", info.Author)
		assert.Equal(t, 3, info.CommitCount)
		assert.Equal(t, 3, info.CommentCount)
		assert.Equal(t, []string{"@alice"}, info.Assignees)
		assert.Equal(t, []string{"enhancement"}, info.Labels)
		assert.Equal(t, []string{"Board"}, info.Projects)
		assert.Equal(t, "v1.0", info.Milestone)
		assert.Len(t, info.Issues, 1)
		assert.Equal(t, 10, info.Issues[0].Number)
		assert.Equal(t, "Feature request", info.Issues[0].Title)
	})

	t.Run("nil milestone", func(t *testing.T) {
		emptyURL, _ := url.Parse("")
		node := &prNode{
			Milestone: nil,
			URL:       githubv4.URI{URL: emptyURL},
		}
		info := fromPRNode(node)
		assert.Equal(t, "", info.Milestone)
	})

	t.Run("empty lists", func(t *testing.T) {
		emptyURL, _ := url.Parse("")
		node := &prNode{URL: githubv4.URI{URL: emptyURL}}
		info := fromPRNode(node)
		assert.Empty(t, info.Reviewers)
		assert.Empty(t, info.Assignees)
		assert.Empty(t, info.Labels)
		assert.Empty(t, info.Projects)
		assert.Empty(t, info.Issues)
	})

	t.Run("reviewer types — user and team", func(t *testing.T) {
		emptyURL, _ := url.Parse("")
		node := &prNode{URL: githubv4.URI{URL: emptyURL}}
		node.ReviewRequests.Nodes = []struct {
			RequestedReviewer prReviewerNode
		}{
			{RequestedReviewer: prReviewerNode{}},
			{RequestedReviewer: prReviewerNode{}},
		}
		// user reviewer
		node.ReviewRequests.Nodes[0].RequestedReviewer.User.Login = "bob"
		// team reviewer
		node.ReviewRequests.Nodes[1].RequestedReviewer.Team.Name = "core-team"

		info := fromPRNode(node)
		assert.Equal(t, []string{"@bob", "core-team"}, info.Reviewers)
	})

	t.Run("state normalized to lowercase", func(t *testing.T) {
		emptyURL, _ := url.Parse("")
		node := &prNode{State: "CLOSED", URL: githubv4.URI{URL: emptyURL}}
		info := fromPRNode(node)
		assert.Equal(t, "closed", info.State)
	})
}
