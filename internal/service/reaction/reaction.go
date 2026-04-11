package reaction

import (
	"context"
	"fmt"
	"slices"

	ghrest "github.com/google/go-github/v69/github"
	"github.com/shurcooL/githubv4"
)

// ValidReactions lists the CLI-accepted reaction names.
// thumbsup/thumbsdown replace the shell-unfriendly +1/-1.
var ValidReactions = []string{"thumbsup", "thumbsdown", "laugh", "confused", "heart", "hooray", "rocket", "eyes"}

// reactionCreator creates reactions on GitHub comments via REST API
type reactionCreator interface {
	CreatePullRequestCommentReaction(ctx context.Context, owner, repo string, id int64, content string) (*ghrest.Reaction, *ghrest.Response, error)
	CreateIssueCommentReaction(ctx context.Context, owner, repo string, id int64, content string) (*ghrest.Reaction, *ghrest.Response, error)
}

// graphQLMutator executes GraphQL mutations and queries
type graphQLMutator interface {
	Query(ctx context.Context, q any, variables map[string]any) error
	Mutate(ctx context.Context, m any, input githubv4.Input, variables map[string]any) error
}

// Service provides reaction operations via REST and GraphQL APIs
type Service struct {
	reactions reactionCreator
	gql       graphQLMutator
	owner     string
	repo      string
}

// NewService creates a new reaction service
func NewService(client *ghrest.Client, gql graphQLMutator, owner, repo string) *Service {
	return &Service{reactions: client.Reactions, gql: gql, owner: owner, repo: repo}
}

// AddToReviewComment adds a reaction to a PR review comment (inline line comment)
func (s *Service) AddToReviewComment(commentID int64, reaction string) error {
	if !isValid(reaction) {
		return fmt.Errorf("invalid reaction '%s', valid: %v", reaction, ValidReactions)
	}
	content := toRESTContent(reaction)
	_, _, err := s.reactions.CreatePullRequestCommentReaction(context.Background(), s.owner, s.repo, commentID, content)
	if err != nil {
		return fmt.Errorf("Reactions.CreatePullRequestCommentReaction [comment=%d, reaction='%s']: %w", commentID, reaction, err)
	}
	return nil
}

// AddToIssueComment adds a reaction to a top-level PR/issue comment
func (s *Service) AddToIssueComment(commentID int64, reaction string) error {
	if !isValid(reaction) {
		return fmt.Errorf("invalid reaction '%s', valid: %v", reaction, ValidReactions)
	}
	content := toRESTContent(reaction)
	_, _, err := s.reactions.CreateIssueCommentReaction(context.Background(), s.owner, s.repo, commentID, content)
	if err != nil {
		return fmt.Errorf("Reactions.CreateIssueCommentReaction [comment=%d, reaction='%s']: %w", commentID, reaction, err)
	}
	return nil
}

// reviewNodeQuery fetches the GraphQL node ID of a review by its database ID.
// GitHub REST API does not support reactions on review bodies, so we need GraphQL.
type reviewNodeQuery struct {
	Repository struct {
		PullRequest struct {
			Reviews struct {
				Nodes []struct {
					ID         githubv4.ID
					DatabaseID int64
				}
			} `graphql:"reviews(first: 100)"`
		} `graphql:"pullRequest(number: $number)"`
	} `graphql:"repository(owner: $owner, name: $repo)"`
}

// addReactionMutation adds a reaction to any reactable GitHub node
type addReactionMutation struct {
	AddReaction struct {
		Reaction struct {
			Content githubv4.ReactionContent
		}
	} `graphql:"addReaction(input: $input)"`
}

// AddReactionInput is the input for the addReaction GraphQL mutation.
// Must be PascalCase to match the GitHub GraphQL schema type name.
type AddReactionInput struct {
	SubjectID githubv4.ID              `json:"subjectId"`
	Content   githubv4.ReactionContent `json:"content"`
}

// AddToReview adds a reaction to a PR review body using GraphQL.
// GitHub REST API has no endpoint for reactions on review bodies.
func (s *Service) AddToReview(prNumber int, reviewID int64, reaction string) error {
	if !isValid(reaction) {
		return fmt.Errorf("invalid reaction '%s', valid: %v", reaction, ValidReactions)
	}

	// look up the review's GraphQL node ID from the database ID
	nodeID, err := s.findReviewNodeID(prNumber, reviewID)
	if err != nil {
		return fmt.Errorf("findReviewNodeID [pr=%d, review=%d]: %w", prNumber, reviewID, err)
	}

	input := AddReactionInput{
		SubjectID: nodeID,
		Content:   toGraphQLContent(reaction),
	}

	var mutation addReactionMutation
	if err := s.gql.Mutate(context.Background(), &mutation, input, nil); err != nil {
		return fmt.Errorf("gql.Mutate addReaction [pr=%d, review=%d, reaction='%s']: %w", prNumber, reviewID, reaction, err)
	}
	return nil
}

// findReviewNodeID returns the GraphQL node ID for the review with the given database ID.
func (s *Service) findReviewNodeID(prNumber int, reviewID int64) (githubv4.ID, error) {
	vars := map[string]any{
		"owner":  githubv4.String(s.owner),
		"repo":   githubv4.String(s.repo),
		"number": githubv4.Int(prNumber), //nolint:gosec // PR numbers fit in int32
	}

	var query reviewNodeQuery
	if err := s.gql.Query(context.Background(), &query, vars); err != nil {
		return nil, fmt.Errorf("gql.Query [pr=%d]: %w", prNumber, err)
	}

	for _, node := range query.Repository.PullRequest.Reviews.Nodes {
		if node.DatabaseID == reviewID {
			return node.ID, nil
		}
	}

	return nil, fmt.Errorf("review #%d not found in PR #%d", reviewID, prNumber)
}

func isValid(reaction string) bool {
	return slices.Contains(ValidReactions, reaction)
}

// toRESTContent maps CLI reaction names to GitHub REST API content values.
// thumbsup/thumbsdown replace +1/-1 which are shell-unfriendly.
func toRESTContent(reaction string) string {
	switch reaction {
	case "thumbsup":
		return "+1"
	case "thumbsdown":
		return "-1"
	default:
		return reaction
	}
}

// toGraphQLContent maps CLI reaction names to GitHub GraphQL ReactionContent enum values.
func toGraphQLContent(reaction string) githubv4.ReactionContent {
	switch reaction {
	case "thumbsup":
		return githubv4.ReactionContentThumbsUp
	case "thumbsdown":
		return githubv4.ReactionContentThumbsDown
	case "laugh":
		return githubv4.ReactionContentLaugh
	case "confused":
		return githubv4.ReactionContentConfused
	case "heart":
		return githubv4.ReactionContentHeart
	case "hooray":
		return githubv4.ReactionContentHooray
	case "rocket":
		return githubv4.ReactionContentRocket
	case "eyes":
		return githubv4.ReactionContentEyes
	default:
		return githubv4.ReactionContent(reaction)
	}
}
