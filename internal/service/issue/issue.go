// Package issue provides GitHub issue read operations via GraphQL.
package issue

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	"github.com/shurcooL/githubv4"

	"github.com/ivanov-gv/contribute/internal/utils/pagination"
)

// graphQLQuerier executes GraphQL queries
type graphQLQuerier interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
}

// Service provides issue operations via GraphQL
type Service struct {
	gql   graphQLQuerier
	owner string
	repo  string
}

// NewService creates a new issue service
func NewService(gql graphQLQuerier, owner, repo string) *Service {
	return &Service{gql: gql, owner: owner, repo: repo}
}

// Info holds issue details from GraphQL
type Info struct {
	Number       int
	Title        string
	State        string
	Body         string
	URL          string
	Author       string
	Labels       []string
	Assignees    []string
	CommentCount int
	Comments     []Comment
	LinkedPRs    []LinkedPR
}

// Comment holds an issue comment
type Comment struct {
	DatabaseID int64
	Author     string
	Body       string
	CreatedAt  string
}

// LinkedPR is a pull request that references this issue
type LinkedPR struct {
	Number int
	Title  string
	State  string
}

// issueNode is the issue shape returned by the query
type issueNode struct {
	Number githubv4.Int
	Title  githubv4.String
	State  githubv4.String
	Body   githubv4.String
	URL    githubv4.URI
	Author struct {
		Login githubv4.String
	}
	Labels struct {
		Nodes []struct {
			Name githubv4.String
		}
	} `graphql:"labels(first: 20)"`
	Assignees struct {
		Nodes []struct {
			Login githubv4.String
		}
	} `graphql:"assignees(first: 20)"`
	Comments struct {
		TotalCount githubv4.Int
		Nodes      []struct {
			DatabaseID int64
			Author     struct {
				Login githubv4.String
			}
			Body      githubv4.String
			CreatedAt githubv4.DateTime
		}
	} `graphql:"comments(first: 50)"`
	TimelineItems struct {
		Nodes []struct {
			CrossReferencedEvent struct {
				Source struct {
					PullRequest struct {
						Number githubv4.Int
						Title  githubv4.String
						State  githubv4.String
					} `graphql:"... on PullRequest"`
				}
			} `graphql:"... on CrossReferencedEvent"`
		}
	} `graphql:"timelineItems(first: 50, itemTypes: CROSS_REFERENCED_EVENT)"`
}

// issueQuery defines the GraphQL query for a single issue
type issueQuery struct {
	Repository struct {
		Issue issueNode `graphql:"issue(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// Get returns issue info by number
func (s *Service) Get(number int) (*Info, error) {
	var query issueQuery
	variables := map[string]interface{}{
		"owner":  githubv4.String(s.owner),
		"repo":   githubv4.String(s.repo),
		"number": githubv4.Int(number), //nolint:gosec // issue numbers fit in int32
	}
	if err := s.gql.Query(context.Background(), &query, variables); err != nil {
		return nil, fmt.Errorf("gql.Query [number=%d]: %w", number, err)
	}
	return fromIssueNode(&query.Repository.Issue), nil
}

// ListItem holds summary info for issue listing
type ListItem struct {
	Number   int
	Title    string
	State    string
	Author   string
	Labels   []string
	Comments int
}

// issueListNode is the issue shape for listing
type issueListNode struct {
	Number githubv4.Int
	Title  githubv4.String
	State  githubv4.String
	Author struct {
		Login githubv4.String
	}
	Labels struct {
		Nodes []struct {
			Name githubv4.String
		}
	} `graphql:"labels(first: 10)"`
	Comments struct {
		TotalCount githubv4.Int
	}
}

// issueListQuery fetches open issues with label filter, cursor-paginated
type issueListQuery struct {
	Repository struct {
		Issues struct {
			Nodes    []issueListNode
			PageInfo pagination.PageInfo
		} `graphql:"issues(first: 100, after: $cursor, states: OPEN, orderBy: {field: CREATED_AT, direction: DESC}, labels: $labels)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// issueListQueryNoLabel fetches open issues without label filter, cursor-paginated
type issueListQueryNoLabel struct {
	Repository struct {
		Issues struct {
			Nodes    []issueListNode
			PageInfo pagination.PageInfo
		} `graphql:"issues(first: 100, after: $cursor, states: OPEN, orderBy: {field: CREATED_AT, direction: DESC})"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// List returns open issues, optionally filtered by label, up to limit total items.
// Fetches all pages automatically until limit is reached or no more pages exist.
func (s *Service) List(limit int, labels []string) ([]ListItem, error) {
	variables := map[string]interface{}{
		"owner":  githubv4.String(s.owner),
		"repo":   githubv4.String(s.repo),
		"cursor": (*githubv4.String)(nil),
	}

	var nodes []issueListNode

	if len(labels) > 0 {
		gqlLabels := lo.Map(labels, func(l string, _ int) *githubv4.String {
			s := githubv4.String(l)
			return &s
		})
		variables["labels"] = gqlLabels

		for {
			var query issueListQuery
			if err := s.gql.Query(context.Background(), &query, variables); err != nil {
				return nil, fmt.Errorf("gql.Query [labels='%v']: %w", labels, err)
			}
			nodes = append(nodes, query.Repository.Issues.Nodes...)
			if len(nodes) >= limit || !query.Repository.Issues.PageInfo.HasMore() {
				break
			}
			variables["cursor"] = query.Repository.Issues.PageInfo.Cursor()
		}
	} else {
		for {
			var query issueListQueryNoLabel
			if err := s.gql.Query(context.Background(), &query, variables); err != nil {
				return nil, fmt.Errorf("gql.Query [no labels]: %w", err)
			}
			nodes = append(nodes, query.Repository.Issues.Nodes...)
			if len(nodes) >= limit || !query.Repository.Issues.PageInfo.HasMore() {
				break
			}
			variables["cursor"] = query.Repository.Issues.PageInfo.Cursor()
		}
	}

	if len(nodes) > limit {
		nodes = nodes[:limit]
	}

	return lo.Map(nodes, func(n issueListNode, _ int) ListItem {
		return fromIssueListNode(&n)
	}), nil
}

func fromIssueNode(n *issueNode) *Info {
	info := &Info{
		Number:       int(n.Number),
		Title:        string(n.Title),
		State:        string(n.State),
		Body:         string(n.Body),
		URL:          n.URL.String(),
		Author:       string(n.Author.Login),
		CommentCount: int(n.Comments.TotalCount),
		Labels: lo.Map(n.Labels.Nodes, func(l struct{ Name githubv4.String }, _ int) string {
			return string(l.Name)
		}),
		Assignees: lo.Map(n.Assignees.Nodes, func(a struct{ Login githubv4.String }, _ int) string {
			return "@" + string(a.Login)
		}),
	}

	info.Comments = lo.Map(n.Comments.Nodes, func(c struct {
		DatabaseID int64
		Author     struct{ Login githubv4.String }
		Body       githubv4.String
		CreatedAt  githubv4.DateTime
	}, _ int) Comment {
		return Comment{
			DatabaseID: c.DatabaseID,
			Author:     string(c.Author.Login),
			Body:       string(c.Body),
			CreatedAt:  c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
	})

	// linked PRs from cross-reference events — deduplicate by PR number
	seen := make(map[int]bool)
	for _, item := range n.TimelineItems.Nodes {
		pr := item.CrossReferencedEvent.Source.PullRequest
		prNumber := int(pr.Number)
		if prNumber == 0 || seen[prNumber] {
			continue
		}
		seen[prNumber] = true
		info.LinkedPRs = append(info.LinkedPRs, LinkedPR{
			Number: prNumber,
			Title:  string(pr.Title),
			State:  string(pr.State),
		})
	}

	return info
}

func fromIssueListNode(n *issueListNode) ListItem {
	return ListItem{
		Number:   int(n.Number),
		Title:    string(n.Title),
		State:    string(n.State),
		Author:   string(n.Author.Login),
		Comments: int(n.Comments.TotalCount),
		Labels: lo.Map(n.Labels.Nodes, func(l struct{ Name githubv4.String }, _ int) string {
			return string(l.Name)
		}),
	}
}
