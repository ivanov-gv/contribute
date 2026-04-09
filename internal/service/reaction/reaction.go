package reaction

import (
	"context"
	"fmt"

	ghrest "github.com/google/go-github/v69/github"
)

// Valid reaction types for GitHub REST API
var ValidReactions = []string{"+1", "-1", "laugh", "confused", "heart", "hooray", "rocket", "eyes"}

// reactionCreator creates reactions on GitHub comments
type reactionCreator interface {
	CreatePullRequestCommentReaction(ctx context.Context, owner, repo string, id int64, content string) (*ghrest.Reaction, *ghrest.Response, error)
	CreateIssueCommentReaction(ctx context.Context, owner, repo string, id int64, content string) (*ghrest.Reaction, *ghrest.Response, error)
}

// Service provides reaction operations via REST API
type Service struct {
	reactions reactionCreator
	owner     string
	repo      string
}

// NewService creates a new reaction service
func NewService(client *ghrest.Client, owner, repo string) *Service {
	return &Service{reactions: client.Reactions, owner: owner, repo: repo}
}

// AddToReviewComment adds a reaction to a PR review comment
func (s *Service) AddToReviewComment(commentID int64, reaction string) error {
	if !isValid(reaction) {
		return fmt.Errorf("invalid reaction '%s', valid: %v", reaction, ValidReactions)
	}
	_, _, err := s.reactions.CreatePullRequestCommentReaction(context.Background(), s.owner, s.repo, commentID, reaction)
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
	_, _, err := s.reactions.CreateIssueCommentReaction(context.Background(), s.owner, s.repo, commentID, reaction)
	if err != nil {
		return fmt.Errorf("Reactions.CreateIssueCommentReaction [comment=%d, reaction='%s']: %w", commentID, reaction, err)
	}
	return nil
}

func isValid(reaction string) bool {
	for _, r := range ValidReactions {
		if r == reaction {
			return true
		}
	}
	return false
}
