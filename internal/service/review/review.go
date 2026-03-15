package review

import (
	"context"
	"fmt"
	"sort"

	"github.com/shurcooL/githubv4"

	"github.com/ivanov-gv/gh-contribute/internal/utils/format"
)

// Service provides review detail operations via GraphQL
type Service struct {
	gql   *githubv4.Client
	owner string
	repo  string
}

// NewService creates a new review service
func NewService(gql *githubv4.Client, owner, repo string) *Service {
	return &Service{gql: gql, owner: owner, repo: repo}
}

// ReviewComment holds a single inline review comment
type ReviewComment struct {
	DatabaseID      int64
	Author          string
	Body            string
	CreatedAt       string
	Path            string
	Line            int
	StartLine       int
	DiffHunk        string
	ReplyToID       int64 // 0 if top-level
	IsMinimized     bool
	MinimizedReason string
	Outdated        bool
	SubjectType     string // LINE or FILE
	Reactions       []format.Reaction
}

// ReviewDetail holds the full review with its inline comments
type ReviewDetail struct {
	DatabaseID  int64
	Author      string
	Body        string
	State       string
	CreatedAt   string
	ViewerLogin string
	Comments    []ReviewComment
	Reactions   []format.Reaction
}

// reactionNode is a single reaction with content and author
type reactionNode struct {
	Content githubv4.String
	User    struct {
		Login githubv4.String
	}
}

// reviewCommentNode is a single inline review comment node
type reviewCommentNode struct {
	DatabaseID githubv4.Int
	Author     struct {
		Login githubv4.String
	}
	Body            githubv4.String
	CreatedAt       githubv4.DateTime
	Path            githubv4.String
	Line            *githubv4.Int
	StartLine       *githubv4.Int
	DiffHunk        githubv4.String
	ReplyTo         *struct {
		DatabaseID githubv4.Int
	}
	IsMinimized     githubv4.Boolean
	MinimizedReason githubv4.String
	Outdated        githubv4.Boolean
	SubjectType     githubv4.String
	Reactions       struct {
		Nodes []reactionNode
	} `graphql:"reactions(first: 20)"`
}

// reviewDetailNode is a single review node with inline comments
type reviewDetailNode struct {
	DatabaseID githubv4.Int
	Author     struct {
		Login githubv4.String
	}
	Body      githubv4.String
	State     githubv4.String
	CreatedAt githubv4.DateTime
	Reactions struct {
		Nodes []reactionNode
	} `graphql:"reactions(first: 20)"`
	Comments struct {
		Nodes []reviewCommentNode
	} `graphql:"comments(first: 100)"`
}

// We need to find the review by databaseId, but GraphQL doesn't support filtering by databaseId directly.
// Instead, fetch all reviews and filter client-side.

// allReviewsQuery defines the GraphQL query shape for fetching all reviews with inline comments
type allReviewsQuery struct {
	Viewer struct {
		Login githubv4.String
	}
	Repository struct {
		PullRequest struct {
			Reviews struct {
				Nodes []reviewDetailNode
			} `graphql:"reviews(first: 100)"`
		} `graphql:"pullRequest(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// Get returns the review detail with all inline comments
func (s *Service) Get(prNumber int, reviewDatabaseID int64) (*ReviewDetail, error) {
	var query allReviewsQuery
	variables := map[string]interface{}{
		"owner":  githubv4.String(s.owner),
		"repo":   githubv4.String(s.repo),
		"number": githubv4.Int(prNumber),
	}
	if err := s.gql.Query(context.Background(), &query, variables); err != nil {
		return nil, fmt.Errorf("gql.Query [pr=%d, review=%d]: %w", prNumber, reviewDatabaseID, err)
	}

	for _, n := range query.Repository.PullRequest.Reviews.Nodes {
		if int64(n.DatabaseID) == reviewDatabaseID {
			return mapReviewDetail(&n, string(query.Viewer.Login)), nil
		}
	}

	return nil, fmt.Errorf("review #%d not found in PR #%d", reviewDatabaseID, prNumber)
}

func mapReviewDetail(n *reviewDetailNode, viewerLogin string) *ReviewDetail {
	detail := &ReviewDetail{
		DatabaseID:  int64(n.DatabaseID),
		Author:      string(n.Author.Login),
		Body:        string(n.Body),
		State:       string(n.State),
		CreatedAt:   n.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		ViewerLogin: viewerLogin,
		Reactions:   mapReactions(n.Reactions.Nodes),
	}

	for _, c := range n.Comments.Nodes {
		rc := ReviewComment{
			DatabaseID:      int64(c.DatabaseID),
			Author:          string(c.Author.Login),
			Body:            string(c.Body),
			CreatedAt:       c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			Path:            string(c.Path),
			DiffHunk:        string(c.DiffHunk),
			IsMinimized:     bool(c.IsMinimized),
			MinimizedReason: string(c.MinimizedReason),
			Outdated:        bool(c.Outdated),
			SubjectType:     string(c.SubjectType),
			Reactions:       mapReactions(c.Reactions.Nodes),
		}
		if c.Line != nil {
			rc.Line = int(*c.Line)
		}
		if c.StartLine != nil {
			rc.StartLine = int(*c.StartLine)
		}
		if c.ReplyTo != nil {
			rc.ReplyToID = int64(c.ReplyTo.DatabaseID)
		}
		detail.Comments = append(detail.Comments, rc)
	}

	sort.Slice(detail.Comments, func(i, j int) bool {
		return detail.Comments[i].CreatedAt < detail.Comments[j].CreatedAt
	})

	return detail
}

func mapReactions(nodes []reactionNode) []format.Reaction {
	reactions := make([]format.Reaction, len(nodes))
	for i, n := range nodes {
		reactions[i] = format.Reaction{Content: string(n.Content), Author: string(n.User.Login)}
	}
	return reactions
}
