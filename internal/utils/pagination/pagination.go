// Package pagination provides cursor-based pagination helpers for GitHub GraphQL API.
package pagination

import "github.com/shurcooL/githubv4"

const defaultPageSize = 100

// PageInfo holds cursor-based pagination state from GraphQL responses
type PageInfo struct {
	HasNextPage githubv4.Boolean
	EndCursor   githubv4.String
}

// HasMore returns true if there are additional pages to fetch
func (p PageInfo) HasMore() bool {
	return bool(p.HasNextPage)
}

// Cursor returns the end cursor for the next page query, or nil for the first page
func (p PageInfo) Cursor() *githubv4.String {
	if string(p.EndCursor) == "" {
		return nil
	}
	return &p.EndCursor
}
