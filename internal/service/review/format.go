package review

import (
	"fmt"
	"strings"

	"github.com/ivanov-gv/contribute/internal/utils/format"
)

// Format renders the review detail as human-readable markdown.
// When showDiff is true, the diff hunk is included below each thread header.
// When showHidden is false, resolved threads are collapsed to a header-only line.
func (d *ReviewDetail) Format(showDiff bool, showHidden bool) string {
	var b strings.Builder

	// review header line — include hidden marker if the review is minimized
	authorDisplay := format.Author(d.Author, d.ViewerLogin)
	if d.IsMinimized {
		reason := format.EnumLabel(d.MinimizedReason)
		if reason == "" {
			reason = "hidden"
		}
		fmt.Fprintf(&b, "review #%d by %s  | hidden: %s  \n", d.DatabaseID, authorDisplay, reason)
	} else {
		fmt.Fprintf(&b, "review #%d by %s  \n", d.DatabaseID, authorDisplay)
	}
	fmt.Fprintf(&b, "_%s_\n", format.Date(d.CreatedAt))
	b.WriteString("\n")

	hasBodyReactions := len(d.Reactions) > 0

	if d.Body != "" && !hasBodyReactions && len(d.ThreadGroups) > 0 {
		// body + first thread group merged into one > block
		formatBodyWithFirstThread(&b, d, showDiff, showHidden)
	} else if d.Body != "" {
		// body in > block, then reactions, then threads outside >
		formatBodyWithReactions(&b, d, showDiff, showHidden)
	} else if len(d.ThreadGroups) > 0 {
		// no body — threads after --- separator
		formatThreadsOnly(&b, d, showDiff, showHidden)
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// formatBodyWithFirstThread renders body and first thread group inside a single > block.
// Used when the review body has no reactions.
func formatBodyWithFirstThread(b *strings.Builder, d *ReviewDetail, showDiff bool, showHidden bool) {
	// body lines inside >
	bodyLines := splitBody(d.Body)
	for _, line := range bodyLines {
		b.WriteString(">" + line + "\n")
	}
	// visual separator: two blank > lines between body and thread content
	b.WriteString(">\n>\n")

	firstGroup := d.ThreadGroups[0]
	// first thread group content inside > block
	threadContent := formatThreadGroupPlain(firstGroup, d.ViewerLogin, showDiff, showHidden)
	for _, line := range strings.Split(strings.TrimRight(threadContent, "\n"), "\n") {
		b.WriteString(">" + line + "\n")
	}

	// end > block — first thread group reactions outside >
	writeThreadGroupReactions(b, firstGroup, d.ViewerLogin)

	// remaining thread groups outside >
	for i := 1; i < len(d.ThreadGroups); i++ {
		b.WriteString("\n---\n")
		b.WriteString(formatThreadGroup(d.ThreadGroups[i], d.ViewerLogin, showDiff, showHidden))
	}
}

// formatBodyWithReactions renders body in >, then reactions, then threads outside >.
func formatBodyWithReactions(b *strings.Builder, d *ReviewDetail, showDiff bool, showHidden bool) {
	bodyLines := splitBody(d.Body)
	for _, line := range bodyLines {
		b.WriteString(">" + line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(format.Reactions(d.Reactions, d.ViewerLogin))

	// thread groups outside > with --- separators
	for i, group := range d.ThreadGroups {
		if i > 0 || len(d.ThreadGroups) > 1 {
			b.WriteString("\n---\n")
		} else {
			b.WriteString("\n")
		}
		b.WriteString(formatThreadGroup(group, d.ViewerLogin, showDiff, showHidden))
	}
}

// formatThreadsOnly renders threads when there is no body.
func formatThreadsOnly(b *strings.Builder, d *ReviewDetail, showDiff bool, showHidden bool) {
	for i, group := range d.ThreadGroups {
		if i == 0 {
			b.WriteString("---\n")
		} else {
			b.WriteString("\n---\n")
		}
		b.WriteString(formatThreadGroup(group, d.ViewerLogin, showDiff, showHidden))
	}
}

// formatThreadGroupPlain renders thread group content WITHOUT > prefix and WITHOUT reactions.
// Used for content that will be placed inside a > block.
func formatThreadGroupPlain(g ReviewThreadGroup, viewerLogin string, showDiff bool, showHidden bool) string {
	var b strings.Builder

	// thread header
	location := formatLocation(g)
	if g.IsResolved {
		fmt.Fprintf(&b, "thread #%d  %s | hidden: Resolved  \n", g.ThreadID, location)
	} else {
		fmt.Fprintf(&b, "thread #%d  %s  \n", g.ThreadID, location)
	}

	if g.IsResolved && !showHidden {
		return b.String()
	}

	if showDiff && g.DiffHunk != "" {
		b.WriteString("\n```diff\n")
		b.WriteString(g.DiffHunk + "\n")
		b.WriteString("```\n")
	}

	// comments without reactions (reactions go outside > block)
	for _, c := range g.Comments {
		b.WriteString(formatReviewCommentPlain(c, viewerLogin, showHidden))
	}

	return b.String()
}

// formatReviewCommentPlain renders a review comment WITHOUT > prefix on body and WITHOUT reactions.
func formatReviewCommentPlain(c ReviewComment, viewerLogin string, showHidden bool) string {
	var b strings.Builder
	authorDisplay := format.Author(c.Author, viewerLogin)

	if c.IsMinimized {
		reason := format.EnumLabel(c.MinimizedReason)
		if reason == "" {
			reason = "hidden"
		}
		if !showHidden {
			fmt.Fprintf(&b, "comment #%d by %s | hidden: %s\n", c.DatabaseID, authorDisplay, reason)
			return b.String()
		}
		// full content with hidden marker
		header := formatCommentHeader(c, authorDisplay)
		header = strings.TrimRight(header, " \n")
		fmt.Fprintf(&b, "%s | hidden: %s  \n", header, reason)
	} else {
		b.WriteString(formatCommentHeader(c, authorDisplay))
	}

	fmt.Fprintf(&b, "_%s_\n", format.Date(c.CreatedAt))
	b.WriteString("\n")

	// body text without > prefix (the caller adds > for the whole block)
	commentBody := strings.TrimRight(strings.ReplaceAll(c.Body, "\r\n", "\n"), "\n")
	for _, line := range strings.Split(commentBody, "\n") {
		b.WriteString(line + "\n")
	}

	return b.String()
}

// writeThreadGroupReactions writes all comment reactions for a thread group outside the > block.
func writeThreadGroupReactions(b *strings.Builder, g ReviewThreadGroup, viewerLogin string) {
	for _, c := range g.Comments {
		if c.IsMinimized {
			continue
		}
		if reactionsStr := format.Reactions(c.Reactions, viewerLogin); reactionsStr != "" {
			b.WriteString("\n")
			b.WriteString(reactionsStr)
		}
	}
}

func formatThreadGroup(g ReviewThreadGroup, viewerLogin string, showDiff bool, showHidden bool) string {
	var b strings.Builder

	// thread header
	location := formatLocation(g)
	if g.IsResolved {
		fmt.Fprintf(&b, "thread #%d  %s | hidden: Resolved  \n", g.ThreadID, location)
	} else {
		fmt.Fprintf(&b, "thread #%d  %s  \n", g.ThreadID, location)
	}

	// for resolved threads, hide comment content unless showHidden is set
	if g.IsResolved && !showHidden {
		return b.String()
	}

	if showDiff && g.DiffHunk != "" {
		b.WriteString("\n```diff\n")
		b.WriteString(g.DiffHunk + "\n")
		b.WriteString("```\n")
	}

	for i, c := range g.Comments {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(formatReviewComment(c, viewerLogin, showHidden))
	}

	return b.String()
}

func formatReviewComment(c ReviewComment, viewerLogin string, showHidden bool) string {
	var b strings.Builder
	authorDisplay := format.Author(c.Author, viewerLogin)

	if c.IsMinimized {
		reason := format.EnumLabel(c.MinimizedReason)
		if reason == "" {
			reason = "hidden"
		}
		if !showHidden {
			// compact header for hidden comments
			fmt.Fprintf(&b, "comment #%d by %s | hidden: %s\n", c.DatabaseID, authorDisplay, reason)
			return b.String()
		}
		// full content with hidden marker — use reply format if applicable
		header := formatCommentHeader(c, authorDisplay)
		// trim trailing "  \n" and append hidden marker
		header = strings.TrimRight(header, " \n")
		fmt.Fprintf(&b, "%s | hidden: %s  \n", header, reason)
	} else {
		b.WriteString(formatCommentHeader(c, authorDisplay))
	}

	fmt.Fprintf(&b, "_%s_\n", format.Date(c.CreatedAt))
	b.WriteString("\n")

	commentBody := strings.TrimRight(strings.ReplaceAll(c.Body, "\r\n", "\n"), "\n")
	for _, line := range strings.Split(commentBody, "\n") {
		b.WriteString(">" + line + "\n")
	}

	if reactionsStr := format.Reactions(c.Reactions, viewerLogin); reactionsStr != "" {
		b.WriteString("\n")
		b.WriteString(reactionsStr)
	}

	return b.String()
}

// formatCommentHeader renders the comment/reply header line
func formatCommentHeader(c ReviewComment, authorDisplay string) string {
	if c.ReplyToID == 0 {
		return fmt.Sprintf("comment #%d by %s  \n", c.DatabaseID, authorDisplay)
	}
	if c.ReplyToIsExternal {
		return fmt.Sprintf("reply #%d to #%d (not in this review)  by %s  \n", c.DatabaseID, c.ReplyToID, authorDisplay)
	}
	return fmt.Sprintf("reply #%d to #%d  by %s  \n", c.DatabaseID, c.ReplyToID, authorDisplay)
}

// splitBody normalizes and splits the body text into lines for > quoting
func splitBody(body string) []string {
	normalized := strings.TrimRight(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	return strings.Split(normalized, "\n")
}

// formatLocation builds the location string from thread-level fields.
// For up-to-date: `path on lines +startLine to +line`
// For outdated:   `path on original lines startLine to line (outdated)`
func formatLocation(g ReviewThreadGroup) string {
	if g.Path == "" {
		return ""
	}
	if g.IsOutdated {
		return formatOutdatedLocation(g)
	}
	return formatCurrentLocation(g)
}

func formatCurrentLocation(g ReviewThreadGroup) string {
	if g.StartLine > 0 && g.Line > 0 && g.StartLine != g.Line {
		return fmt.Sprintf("%s on lines +%d to +%d", g.Path, g.StartLine, g.Line)
	} else if g.Line > 0 {
		return fmt.Sprintf("%s on line +%d", g.Path, g.Line)
	}
	return g.Path
}

func formatOutdatedLocation(g ReviewThreadGroup) string {
	if g.OriginalStartLine > 0 && g.OriginalLine > 0 && g.OriginalStartLine != g.OriginalLine {
		return fmt.Sprintf("%s on original lines %d to %d (outdated)", g.Path, g.OriginalStartLine, g.OriginalLine)
	} else if g.OriginalLine > 0 {
		return fmt.Sprintf("%s on original line %d (outdated)", g.Path, g.OriginalLine)
	}
	return fmt.Sprintf("%s (outdated)", g.Path)
}
