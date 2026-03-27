package thread

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"

	graphql_model "github.com/ivanov-gv/gh-contribute/internal/model/graphql"
	"github.com/ivanov-gv/gh-contribute/internal/utils/format"
)

// graphQLClient executes GraphQL queries and mutations
type graphQLClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

// Service provides thread lookup operations via GraphQL
type Service struct {
	gql   graphQLClient
	owner string
	repo  string
}

// NewService creates a new thread service
func NewService(gql graphQLClient, owner, repo string) *Service {
	return &Service{gql: gql, owner: owner, repo: repo}
}

// ThreadComment holds a single comment in a thread across all reviews
type ThreadComment struct {
	DatabaseID       int64
	Author           string
	Body             string
	CreatedAt        string
	ReviewDatabaseID int64 // which review this comment belongs to
	ReplyToID        int64 // 0 if thread root
	IsMinimized      bool
	MinimizedReason  string
	Reactions        []format.Reaction
}

// Thread holds all comments in a thread and location info
type Thread struct {
	ThreadID          int64 // databaseId of the first comment
	IsOutdated        bool
	IsResolved        bool
	Path              string
	Line              int
	StartLine         int
	OriginalLine      int
	OriginalStartLine int
	ViewerLogin       string
	Comments          []ThreadComment
}

// reviewThreadNode represents a single review thread with its comments
type reviewThreadNode struct {
	ID                githubv4.ID
	IsOutdated        githubv4.Boolean
	IsResolved        githubv4.Boolean
	Path              githubv4.String
	Line              *githubv4.Int
	StartLine         *githubv4.Int
	OriginalLine      *githubv4.Int
	OriginalStartLine *githubv4.Int
	Comments          struct {
		Nodes []graphql_model.ThreadCommentNode
	} `graphql:"comments(first: 50)"`
}

// resolveThreadMutation is the GraphQL mutation for resolving a review thread
type resolveThreadMutation struct {
	ResolveReviewThread struct {
		Thread struct {
			IsResolved githubv4.Boolean
		}
	} `graphql:"resolveReviewThread(input: $input)"`
}

// unresolveThreadMutation is the GraphQL mutation for unresolving a review thread
type unresolveThreadMutation struct {
	UnresolveReviewThread struct {
		Thread struct {
			IsResolved githubv4.Boolean
		}
	} `graphql:"unresolveReviewThread(input: $input)"`
}

// threadsQuery fetches all review threads for a PR
type threadsQuery struct {
	Viewer struct {
		Login githubv4.String
	}
	Repository struct {
		PullRequest struct {
			ReviewThreads struct {
				Nodes []reviewThreadNode
			} `graphql:"reviewThreads(first: 100)"`
		} `graphql:"pullRequest(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// Get returns the full thread identified by threadID (the databaseId of the first comment).
func (s *Service) Get(prNumber int, threadID int64) (*Thread, error) {
	variables := map[string]interface{}{
		"owner":  githubv4.String(s.owner),
		"repo":   githubv4.String(s.repo),
		"number": githubv4.Int(prNumber), //nolint:gosec // PR numbers fit in int32
	}

	var query threadsQuery
	if err := s.gql.Query(context.Background(), &query, variables); err != nil {
		return nil, fmt.Errorf("gql.Query [pr=%d, thread=%d]: %w", prNumber, threadID, err)
	}

	viewerLogin := string(query.Viewer.Login)
	for _, n := range query.Repository.PullRequest.ReviewThreads.Nodes {
		if len(n.Comments.Nodes) == 0 || n.Comments.Nodes[0].DatabaseID != threadID {
			continue
		}
		return buildThread(n, viewerLogin, threadID), nil
	}

	return nil, fmt.Errorf("thread #%d not found in PR #%d", threadID, prNumber)
}

// Resolve marks a review thread as resolved.
// threadID is the database ID of the first comment in the thread.
func (s *Service) Resolve(prNumber int, threadID int64) error {
	nodeID, err := s.findThreadNodeID(prNumber, threadID)
	if err != nil {
		return fmt.Errorf("findThreadNodeID [pr=%d, thread=%d]: %w", prNumber, threadID, err)
	}

	var mutation resolveThreadMutation
	input := githubv4.ResolveReviewThreadInput{
		ThreadID: nodeID,
	}
	if err := s.gql.Mutate(context.Background(), &mutation, input, nil); err != nil {
		return fmt.Errorf("gql.Mutate resolveReviewThread [pr=%d, thread=%d]: %w", prNumber, threadID, err)
	}
	return nil
}

// Unresolve marks a review thread as unresolved.
// threadID is the database ID of the first comment in the thread.
func (s *Service) Unresolve(prNumber int, threadID int64) error {
	nodeID, err := s.findThreadNodeID(prNumber, threadID)
	if err != nil {
		return fmt.Errorf("findThreadNodeID [pr=%d, thread=%d]: %w", prNumber, threadID, err)
	}

	var mutation unresolveThreadMutation
	input := githubv4.UnresolveReviewThreadInput{
		ThreadID: nodeID,
	}
	if err := s.gql.Mutate(context.Background(), &mutation, input, nil); err != nil {
		return fmt.Errorf("gql.Mutate unresolveReviewThread [pr=%d, thread=%d]: %w", prNumber, threadID, err)
	}
	return nil
}

// findThreadNodeID fetches all threads for a PR and returns the GraphQL node ID
// of the thread whose first comment has the given database ID.
func (s *Service) findThreadNodeID(prNumber int, threadID int64) (githubv4.ID, error) {
	variables := map[string]interface{}{
		"owner":  githubv4.String(s.owner),
		"repo":   githubv4.String(s.repo),
		"number": githubv4.Int(prNumber), //nolint:gosec // PR numbers fit in int32
	}

	var query threadsQuery
	if err := s.gql.Query(context.Background(), &query, variables); err != nil {
		return nil, fmt.Errorf("gql.Query [pr=%d, thread=%d]: %w", prNumber, threadID, err)
	}

	for _, n := range query.Repository.PullRequest.ReviewThreads.Nodes {
		if len(n.Comments.Nodes) > 0 && n.Comments.Nodes[0].DatabaseID == threadID {
			return n.ID, nil
		}
	}

	return nil, fmt.Errorf("thread #%d not found in PR #%d", threadID, prNumber)
}

func buildThread(n reviewThreadNode, viewerLogin string, threadID int64) *Thread {
	t := &Thread{
		ThreadID:    threadID,
		IsOutdated:  bool(n.IsOutdated),
		IsResolved:  bool(n.IsResolved),
		Path:        string(n.Path),
		ViewerLogin: viewerLogin,
	}
	if n.Line != nil {
		t.Line = int(*n.Line)
	}
	if n.StartLine != nil {
		t.StartLine = int(*n.StartLine)
	}
	if n.OriginalLine != nil {
		t.OriginalLine = int(*n.OriginalLine)
	}
	if n.OriginalStartLine != nil {
		t.OriginalStartLine = int(*n.OriginalStartLine)
	}

	for _, c := range n.Comments.Nodes {
		tc := ThreadComment{
			DatabaseID:      c.DatabaseID,
			Author:          string(c.Author.Login),
			Body:            string(c.Body),
			CreatedAt:       c.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			IsMinimized:     bool(c.IsMinimized),
			MinimizedReason: string(c.MinimizedReason),
			Reactions:       graphql_model.MapReactions(c.Reactions.Nodes),
		}
		if c.ReplyTo != nil {
			tc.ReplyToID = c.ReplyTo.DatabaseID
		}
		if c.PullRequestReview != nil {
			tc.ReviewDatabaseID = c.PullRequestReview.DatabaseID
		}
		t.Comments = append(t.Comments, tc)
	}
	return t
}
