package comment

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ivanov-gv/gh-contribute/internal/utils/format"
)

// timelineEntry is a union type for sorting issue comments and reviews together
type timelineEntry struct {
	createdAt    string
	issueComment *IssueComment
	review       *Review
}

// Format renders the full comments result as human-readable markdown.
// When showHidden is true, hidden/minimized items show full content instead of header-only.
func (r *CommentsResult) Format(showHidden bool) string {
	// merge issue comments and reviews into a single timeline
	var entries []timelineEntry
	for i := range r.IssueComments {
		entries = append(entries, timelineEntry{
			createdAt:    r.IssueComments[i].CreatedAt,
			issueComment: &r.IssueComments[i],
		})
	}
	for i := range r.Reviews {
		entries = append(entries, timelineEntry{
			createdAt: r.Reviews[i].CreatedAt,
			review:    &r.Reviews[i],
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].createdAt < entries[j].createdAt
	})

	if len(entries) == 0 {
		return ""
	}

	var b strings.Builder
	for i, e := range entries {
		if i > 0 {
			b.WriteString("\n---\n")
		}
		if e.issueComment != nil {
			b.WriteString(formatIssueComment(e.issueComment, r.ViewerLogin, showHidden))
		} else {
			b.WriteString(formatReview(e.review, r.ViewerLogin, showHidden))
		}
	}

	return b.String() + "\n"
}

func formatIssueComment(c *IssueComment, viewerLogin string, showHidden bool) string {
	var b strings.Builder

	authorDisplay := format.Author(c.Author, viewerLogin)
	if c.IsMinimized {
		reason := format.EnumLabel(c.MinimizedReason)
		if reason == "" {
			reason = "hidden"
		}
		if !showHidden {
			// compact header-only for hidden comments
			b.WriteString(fmt.Sprintf("issue #%d by %s | hidden: %s\n", c.DatabaseID, authorDisplay, reason))
			return b.String()
		}
		// full content with hidden marker in header
		b.WriteString(fmt.Sprintf("issue #%d by %s | hidden: %s  \n", c.DatabaseID, authorDisplay, reason))
	} else {
		b.WriteString(fmt.Sprintf("issue #%d by %s  \n", c.DatabaseID, authorDisplay))
	}
	b.WriteString(fmt.Sprintf("_%s_  \n", format.Date(c.CreatedAt)))
	b.WriteString("\n")
	body := strings.TrimRight(strings.ReplaceAll(c.Body, "\r\n", "\n"), "\n")
	for _, line := range strings.Split(body, "\n") {
		b.WriteString(">" + line + "\n")
	}
	if reactionsStr := format.Reactions(c.Reactions, viewerLogin); reactionsStr != "" {
		b.WriteString("\n")
		b.WriteString(reactionsStr)
	}

	return b.String()
}

func formatReview(r *Review, viewerLogin string, showHidden bool) string {
	var b strings.Builder

	authorDisplay := format.Author(r.Author, viewerLogin)

	if r.IsHidden {
		reason := r.HiddenReason
		if reason == "" {
			reason = "hidden"
		}
		if !showHidden {
			// compact header-only for hidden reviews
			b.WriteString(fmt.Sprintf("review #%d by %s | hidden: %s\n", r.DatabaseID, authorDisplay, reason))
			return b.String()
		}
		// full content with hidden marker in header
		b.WriteString(fmt.Sprintf("review #%d by %s | hidden: %s  \n", r.DatabaseID, authorDisplay, reason))
	} else if r.State == "DISMISSED" {
		if !showHidden {
			b.WriteString(fmt.Sprintf("review #%d by %s | hidden: Dismissed\n", r.DatabaseID, authorDisplay))
			return b.String()
		}
		b.WriteString(fmt.Sprintf("review #%d by %s | hidden: Dismissed  \n", r.DatabaseID, authorDisplay))
	} else {
		b.WriteString(fmt.Sprintf("review #%d by %s  \n", r.DatabaseID, authorDisplay))
	}
	b.WriteString(fmt.Sprintf("_%s_  \n", format.Date(r.CreatedAt)))
	b.WriteString("\n")

	if r.Body != "" {
		reviewBody := strings.TrimRight(strings.ReplaceAll(r.Body, "\r\n", "\n"), "\n")
		for _, line := range strings.Split(reviewBody, "\n") {
			b.WriteString(">" + line + "\n")
		}
		b.WriteString("\n")
	}

	if r.CommentCount > 0 {
		b.WriteString(fmt.Sprintf("comments: %d  \n", r.CommentCount))
	}

	b.WriteString(format.Reactions(r.Reactions, viewerLogin))

	return b.String()
}
