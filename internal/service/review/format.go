package review

import (
	"fmt"
	"strings"

	"github.com/ivanov-gv/gh-contribute/internal/utils/format"
)

// Format renders the review detail as human-readable markdown.
// When showDiff is true, diffHunk is included for each thread.
func (d *ReviewDetail) Format(showDiff bool) string {
	var b strings.Builder

	authorDisplay := format.Author(d.Author, d.ViewerLogin)
	b.WriteString(fmt.Sprintf("# review #%d by %s  \n", d.DatabaseID, authorDisplay))
	b.WriteString(fmt.Sprintf("_%s_\n", format.Date(d.CreatedAt)))
	b.WriteString("\n")

	if d.Body != "" {
		b.WriteString(d.Body + "\n")
		b.WriteString("\n")
	}

	b.WriteString(format.Reactions(d.Reactions, d.ViewerLogin))

	for i, thread := range d.Threads {
		b.WriteString("\n---\n")
		b.WriteString(formatThread(thread, i+1, d.ViewerLogin, showDiff))
		b.WriteString("\n---")
	}

	if len(d.Threads) > 0 {
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func formatThread(thread ReviewThread, threadNum int, viewerLogin string, showDiff bool) string {
	var b strings.Builder

	if len(thread.Comments) == 0 {
		return ""
	}

	// root comment — first in the thread
	root := thread.Comments[0]
	b.WriteString(formatThreadRoot(root, thread, threadNum, viewerLogin))

	if showDiff && thread.DiffHunk != "" {
		b.WriteString("\n```diff\n")
		b.WriteString(thread.DiffHunk + "\n")
		b.WriteString("```\n")
	}

	// replies — all subsequent comments
	for _, reply := range thread.Comments[1:] {
		b.WriteString(formatReply(reply, viewerLogin))
	}

	return b.String()
}

// formatThreadRoot formats the first comment in a thread.
// Format: `thread #N comment #ID by author  location  `
func formatThreadRoot(c ReviewComment, thread ReviewThread, threadNum int, viewerLogin string) string {
	var b strings.Builder
	authorDisplay := format.Author(c.Author, viewerLogin)

	if c.IsMinimized {
		reason := format.EnumLabel(c.MinimizedReason)
		if reason == "" {
			reason = "hidden"
		}
		b.WriteString(fmt.Sprintf("thread #%d comment #%d by %s | hidden: %s\n", threadNum, c.DatabaseID, authorDisplay, reason))
		return b.String()
	}

	location := formatLocation(thread)
	b.WriteString(fmt.Sprintf("thread #%d comment #%d by %s  %s  \n", threadNum, c.DatabaseID, authorDisplay, location))
	b.WriteString(fmt.Sprintf("_%s_\n", format.Date(c.CreatedAt)))
	b.WriteString("\n")

	for _, line := range strings.Split(c.Body, "\n") {
		b.WriteString(line + "\n")
	}

	if reactionsStr := format.Reactions(c.Reactions, viewerLogin); reactionsStr != "" {
		b.WriteString(reactionsStr)
	}

	return b.String()
}

// formatReply formats a reply comment in a thread.
// Format: `reply #ID to #parentID  by author`
func formatReply(c ReviewComment, viewerLogin string) string {
	var b strings.Builder
	authorDisplay := format.Author(c.Author, viewerLogin)

	if c.IsMinimized {
		reason := format.EnumLabel(c.MinimizedReason)
		if reason == "" {
			reason = "hidden"
		}
		b.WriteString(fmt.Sprintf("reply #%d to #%d  by %s | hidden: %s\n", c.DatabaseID, c.ReplyToID, authorDisplay, reason))
		return b.String()
	}

	b.WriteString(fmt.Sprintf("reply #%d to #%d  by %s\n", c.DatabaseID, c.ReplyToID, authorDisplay))
	b.WriteString(fmt.Sprintf("_%s_\n", format.Date(c.CreatedAt)))
	b.WriteString("\n")

	for _, line := range strings.Split(c.Body, "\n") {
		b.WriteString(line + "\n")
	}

	if reactionsStr := format.Reactions(c.Reactions, viewerLogin); reactionsStr != "" {
		b.WriteString(reactionsStr)
	}

	return b.String()
}

// formatLocation builds the file/line location string from thread-level fields.
// For up-to-date: `path on lines +startLine to +line`
// For outdated:   `path on original lines startLine to line (outdated)`
func formatLocation(t ReviewThread) string {
	if t.Path == "" {
		return ""
	}
	if t.IsOutdated {
		return formatOutdatedLocation(t)
	}
	return formatCurrentLocation(t)
}

func formatCurrentLocation(t ReviewThread) string {
	if t.StartLine > 0 && t.Line > 0 && t.StartLine != t.Line {
		return fmt.Sprintf("%s on lines +%d to +%d", t.Path, t.StartLine, t.Line)
	} else if t.Line > 0 {
		return fmt.Sprintf("%s on line +%d", t.Path, t.Line)
	}
	return t.Path
}

func formatOutdatedLocation(t ReviewThread) string {
	if t.OriginalStartLine > 0 && t.OriginalLine > 0 && t.OriginalStartLine != t.OriginalLine {
		return fmt.Sprintf("%s on original lines %d to %d (outdated)", t.Path, t.OriginalStartLine, t.OriginalLine)
	} else if t.OriginalLine > 0 {
		return fmt.Sprintf("%s on original line %d (outdated)", t.Path, t.OriginalLine)
	}
	return fmt.Sprintf("%s (outdated)", t.Path)
}
