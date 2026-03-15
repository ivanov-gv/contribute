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

// ReviewComment holds a single comment within a review thread
type ReviewComment struct {
	DatabaseID      int64
	Author          string
	Body            string
	CreatedAt       string
	ReplyToID       int64 // 0 if thread root
	IsMinimized     bool
	MinimizedReason string
	Reactions       []format.Reaction
}

// ReviewThread holds a thread with its location info and all comments (across all reviews)
type ReviewThread struct {
	IsOutdated        bool
	Path              string
	Line              int
	StartLine         int
	OriginalLine      int
	OriginalStartLine int
	DiffHunk          string // populated only when showDiff is true
	Comments          []ReviewComment
}

// ReviewDetail holds the full review with its threads
type ReviewDetail struct {
	DatabaseID  int64
	Author      string
	Body        string
	State       string
	CreatedAt   string
	ViewerLogin string
	Threads     []ReviewThread
	Reactions   []format.Reaction
}

// reactionNode is a single reaction with content and author
type reactionNode struct {
	Content githubv4.String
	User    struct {
		Login githubv4.String
	}
}

// reviewMetaNode holds review-level metadata (author, body, state, reactions)
// Thread contents are fetched separately via reviewThreads.
type reviewMetaNode struct {
	DatabaseID int64
	Author     struct {
		Login githubv4.String
	}
	Body      githubv4.String
	State     githubv4.String
	CreatedAt githubv4.DateTime
	Reactions struct {
		Nodes []reactionNode
	} `graphql:"reactions(first: 20)"`
}

// threadCommentNodeNoDiff - a comment within a review thread, without diffHunk
type threadCommentNodeNoDiff struct {
	DatabaseID int64
	Author     struct {
		Login githubv4.String
	}
	Body            githubv4.String
	CreatedAt       githubv4.DateTime
	IsMinimized     githubv4.Boolean
	MinimizedReason githubv4.String
	ReplyTo         *struct {
		DatabaseID int64
	}
	PullRequestReview *struct {
		DatabaseID int64
	}
	Reactions struct {
		Nodes []reactionNode
	} `graphql:"reactions(first: 20)"`
}

// threadCommentNodeWithDiff - a comment within a review thread, with diffHunk
type threadCommentNodeWithDiff struct {
	DatabaseID int64
	Author     struct {
		Login githubv4.String
	}
	Body            githubv4.String
	CreatedAt       githubv4.DateTime
	IsMinimized     githubv4.Boolean
	MinimizedReason githubv4.String
	DiffHunk        githubv4.String
	ReplyTo         *struct {
		DatabaseID int64
	}
	PullRequestReview *struct {
		DatabaseID int64
	}
	Reactions struct {
		Nodes []reactionNode
	} `graphql:"reactions(first: 20)"`
}

// reviewThreadNodeNoDiff - a review thread node without diffHunk
type reviewThreadNodeNoDiff struct {
	IsOutdated        githubv4.Boolean
	Path              githubv4.String
	Line              *githubv4.Int
	StartLine         *githubv4.Int
	OriginalLine      *githubv4.Int
	OriginalStartLine *githubv4.Int
	Comments          struct {
		Nodes []threadCommentNodeNoDiff
	} `graphql:"comments(first: 50)"`
}

// reviewThreadNodeWithDiff - a review thread node with diffHunk on each comment
type reviewThreadNodeWithDiff struct {
	IsOutdated        githubv4.Boolean
	Path              githubv4.String
	Line              *githubv4.Int
	StartLine         *githubv4.Int
	OriginalLine      *githubv4.Int
	OriginalStartLine *githubv4.Int
	Comments          struct {
		Nodes []threadCommentNodeWithDiff
	} `graphql:"comments(first: 50)"`
}

// We need to find the review by databaseId, but GraphQL doesn't support filtering by databaseId directly.
// Instead, fetch all reviews and threads and filter client-side.

// allReviewsQueryNoDiff fetches review metadata + all threads without diffHunk
type allReviewsQueryNoDiff struct {
	Viewer struct {
		Login githubv4.String
	}
	Repository struct {
		PullRequest struct {
			Reviews struct {
				Nodes []reviewMetaNode
			} `graphql:"reviews(first: 100)"`
			ReviewThreads struct {
				Nodes []reviewThreadNodeNoDiff
			} `graphql:"reviewThreads(first: 100)"`
		} `graphql:"pullRequest(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// allReviewsQueryWithDiff fetches review metadata + all threads with diffHunk
type allReviewsQueryWithDiff struct {
	Viewer struct {
		Login githubv4.String
	}
	Repository struct {
		PullRequest struct {
			Reviews struct {
				Nodes []reviewMetaNode
			} `graphql:"reviews(first: 100)"`
			ReviewThreads struct {
				Nodes []reviewThreadNodeWithDiff
			} `graphql:"reviewThreads(first: 100)"`
		} `graphql:"pullRequest(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// Get returns the review detail with all threads that contain at least one comment from the review.
// Threads are fetched via reviewThreads to include cross-review replies.
// When showDiff is true, diffHunk is fetched and included in the first thread comment.
func (s *Service) Get(prNumber int, reviewDatabaseID int64, showDiff bool) (*ReviewDetail, error) {
	variables := map[string]interface{}{
		"owner":  githubv4.String(s.owner),
		"repo":   githubv4.String(s.repo),
		"number": githubv4.Int(prNumber),
	}

	if showDiff {
		var query allReviewsQueryWithDiff
		if err := s.gql.Query(context.Background(), &query, variables); err != nil {
			return nil, fmt.Errorf("gql.Query [pr=%d, review=%d]: %w", prNumber, reviewDatabaseID, err)
		}
		review := findReviewMeta(query.Repository.PullRequest.Reviews.Nodes, reviewDatabaseID)
		if review == nil {
			return nil, fmt.Errorf("review #%d not found in PR #%d", reviewDatabaseID, prNumber)
		}
		threads := collectThreadsWithDiff(query.Repository.PullRequest.ReviewThreads.Nodes, reviewDatabaseID)
		return buildReviewDetail(review, string(query.Viewer.Login), threads), nil
	}

	var query allReviewsQueryNoDiff
	if err := s.gql.Query(context.Background(), &query, variables); err != nil {
		return nil, fmt.Errorf("gql.Query [pr=%d, review=%d]: %w", prNumber, reviewDatabaseID, err)
	}
	review := findReviewMeta(query.Repository.PullRequest.Reviews.Nodes, reviewDatabaseID)
	if review == nil {
		return nil, fmt.Errorf("review #%d not found in PR #%d", reviewDatabaseID, prNumber)
	}
	threads := collectThreadsNoDiff(query.Repository.PullRequest.ReviewThreads.Nodes, reviewDatabaseID)
	return buildReviewDetail(review, string(query.Viewer.Login), threads), nil
}

// findReviewMeta returns the review metadata node with the given database ID, or nil.
func findReviewMeta(nodes []reviewMetaNode, reviewDatabaseID int64) *reviewMetaNode {
	for i := range nodes {
		if nodes[i].DatabaseID == reviewDatabaseID {
			return &nodes[i]
		}
	}
	return nil
}

// collectThreadsNoDiff returns threads (without diffHunk) that have at least one comment from the given review.
func collectThreadsNoDiff(nodes []reviewThreadNodeNoDiff, reviewDatabaseID int64) []ReviewThread {
	var threads []ReviewThread
	for _, n := range nodes {
		if !threadBelongsToReview(n.Comments.Nodes, reviewDatabaseID) {
			continue
		}
		thread := ReviewThread{
			IsOutdated: bool(n.IsOutdated),
			Path:       string(n.Path),
		}
		setThreadLines(&thread, n.Line, n.StartLine, n.OriginalLine, n.OriginalStartLine)
		for _, c := range n.Comments.Nodes {
			thread.Comments = append(thread.Comments, mapThreadComment(
				c.DatabaseID, c.Author.Login, c.Body, c.CreatedAt,
				c.IsMinimized, c.MinimizedReason, c.ReplyTo, c.Reactions.Nodes,
			))
		}
		threads = append(threads, thread)
	}
	return sortedThreads(threads)
}

// collectThreadsWithDiff returns threads (with diffHunk from first comment) that belong to the review.
func collectThreadsWithDiff(nodes []reviewThreadNodeWithDiff, reviewDatabaseID int64) []ReviewThread {
	var threads []ReviewThread
	for _, n := range nodes {
		if !threadBelongsToReview(n.Comments.Nodes, reviewDatabaseID) {
			continue
		}
		thread := ReviewThread{
			IsOutdated: bool(n.IsOutdated),
			Path:       string(n.Path),
		}
		setThreadLines(&thread, n.Line, n.StartLine, n.OriginalLine, n.OriginalStartLine)
		// diffHunk is the same for all comments in a thread — take it from the first
		if len(n.Comments.Nodes) > 0 {
			thread.DiffHunk = string(n.Comments.Nodes[0].DiffHunk)
		}
		for _, c := range n.Comments.Nodes {
			thread.Comments = append(thread.Comments, mapThreadComment(
				c.DatabaseID, c.Author.Login, c.Body, c.CreatedAt,
				c.IsMinimized, c.MinimizedReason, c.ReplyTo, c.Reactions.Nodes,
			))
		}
		threads = append(threads, thread)
	}
	return sortedThreads(threads)
}

// threadBelongsToReview checks whether any comment in the thread was posted in the given review.
// Uses a generic constraint to work with both comment node types.
func threadBelongsToReview[C interface {
	getReviewID() int64
}](nodes []C, reviewDatabaseID int64) bool {
	for _, c := range nodes {
		if c.getReviewID() == reviewDatabaseID {
			return true
		}
	}
	return false
}

func (c threadCommentNodeNoDiff) getReviewID() int64 {
	if c.PullRequestReview == nil {
		return 0
	}
	return c.PullRequestReview.DatabaseID
}

func (c threadCommentNodeWithDiff) getReviewID() int64 {
	if c.PullRequestReview == nil {
		return 0
	}
	return c.PullRequestReview.DatabaseID
}

func setThreadLines(t *ReviewThread, line, startLine, originalLine, originalStartLine *githubv4.Int) {
	if line != nil {
		t.Line = int(*line)
	}
	if startLine != nil {
		t.StartLine = int(*startLine)
	}
	if originalLine != nil {
		t.OriginalLine = int(*originalLine)
	}
	if originalStartLine != nil {
		t.OriginalStartLine = int(*originalStartLine)
	}
}

func mapThreadComment(
	databaseID int64,
	authorLogin githubv4.String,
	body githubv4.String,
	createdAt githubv4.DateTime,
	isMinimized githubv4.Boolean,
	minimizedReason githubv4.String,
	replyTo *struct{ DatabaseID int64 },
	reactions []reactionNode,
) ReviewComment {
	rc := ReviewComment{
		DatabaseID:      databaseID,
		Author:          string(authorLogin),
		Body:            string(body),
		CreatedAt:       createdAt.UTC().Format("2006-01-02T15:04:05Z"),
		IsMinimized:     bool(isMinimized),
		MinimizedReason: string(minimizedReason),
		Reactions:       mapReactions(reactions),
	}
	if replyTo != nil {
		rc.ReplyToID = replyTo.DatabaseID
	}
	return rc
}

// buildReviewDetail assembles a ReviewDetail from review metadata and pre-grouped threads.
func buildReviewDetail(n *reviewMetaNode, viewerLogin string, threads []ReviewThread) *ReviewDetail {
	return &ReviewDetail{
		DatabaseID:  n.DatabaseID,
		Author:      string(n.Author.Login),
		Body:        string(n.Body),
		State:       string(n.State),
		CreatedAt:   n.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		ViewerLogin: viewerLogin,
		Reactions:   mapReactions(n.Reactions.Nodes),
		Threads:     threads,
	}
}

// sortedThreads sorts threads by the creation time of their first comment.
func sortedThreads(threads []ReviewThread) []ReviewThread {
	sort.Slice(threads, func(i, j int) bool {
		if len(threads[i].Comments) == 0 || len(threads[j].Comments) == 0 {
			return false
		}
		return threads[i].Comments[0].CreatedAt < threads[j].Comments[0].CreatedAt
	})
	return threads
}

func mapReactions(nodes []reactionNode) []format.Reaction {
	reactions := make([]format.Reaction, len(nodes))
	for i, n := range nodes {
		reactions[i] = format.Reaction{Content: string(n.Content), Author: string(n.User.Login)}
	}
	return reactions
}
