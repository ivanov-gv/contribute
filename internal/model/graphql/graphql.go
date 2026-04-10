// Package graphql contains shared GraphQL node types used across multiple services.
package graphql

import (
	"github.com/shurcooL/githubv4"

	"github.com/ivanov-gv/contribute/internal/utils/format"
)

// ReactionNode is a single reaction with content and author, used in GraphQL queries.
type ReactionNode struct {
	Content githubv4.String
	User    struct {
		Login githubv4.String
	}
}

// MapReactions converts GraphQL reaction nodes to format.Reaction slice.
func MapReactions(nodes []ReactionNode) []format.Reaction {
	reactions := make([]format.Reaction, len(nodes))
	for i, n := range nodes {
		reactions[i] = format.Reaction{Content: string(n.Content), Author: string(n.User.Login)}
	}
	return reactions
}

// ThreadCommentNode represents a comment within a review thread.
// Shared between thread and review services (no-diff variant).
type ThreadCommentNode struct {
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
		Nodes []ReactionNode
	} `graphql:"reactions(first: 20)"`
}

// GetID returns the database ID of the comment node.
func (c ThreadCommentNode) GetID() int64 { return c.DatabaseID }
