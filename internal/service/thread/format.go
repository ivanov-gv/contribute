package thread

import (
	"fmt"
	"strings"

	"github.com/ivanov-gv/gh-contribute/internal/utils/format"
)

// Format renders the thread as human-readable markdown.
// When showHidden is true, minimized comments show full content instead of header-only.
func (t *Thread) Format(showHidden bool) string {
	var b strings.Builder

	// thread header with resolved marker when applicable
	location := formatLocation(t)
	if t.IsResolved {
		if showHidden {
			fmt.Fprintf(&b, "thread #%d  %s | hidden: Resolved  \n", t.ThreadID, location)
		} else {
			fmt.Fprintf(&b, "thread #%d  %s | hidden: Resolved\n", t.ThreadID, location)
		}
	} else {
		fmt.Fprintf(&b, "thread #%d  %s  \n", t.ThreadID, location)
	}
	b.WriteString("\n")

	for i, c := range t.Comments {
		if i > 0 {
			b.WriteString("\n---\n")
		}
		b.WriteString(formatThreadComment(c, t.ViewerLogin, showHidden))
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func formatThreadComment(c ThreadComment, viewerLogin string, showHidden bool) string {
	var b strings.Builder
	authorDisplay := format.Author(c.Author, viewerLogin)

	if c.IsMinimized {
		reason := format.EnumLabel(c.MinimizedReason)
		if reason == "" {
			reason = "hidden"
		}
		if !showHidden {
			// compact header-only for minimized comments — trailing spaces for markdown line break
			if c.ReplyToID != 0 {
				fmt.Fprintf(&b, "reply #%d to #%d  by %s  review #%d | hidden: %s  \n", c.DatabaseID, c.ReplyToID, authorDisplay, c.ReviewDatabaseID, reason)
			} else {
				fmt.Fprintf(&b, "comment #%d by %s  review #%d | hidden: %s  \n", c.DatabaseID, authorDisplay, c.ReviewDatabaseID, reason)
			}
			return b.String()
		}
		// full content with hidden marker in header
		if c.ReplyToID != 0 {
			fmt.Fprintf(&b, "reply #%d to #%d  by %s  review #%d | hidden: %s  \n", c.DatabaseID, c.ReplyToID, authorDisplay, c.ReviewDatabaseID, reason)
		} else {
			fmt.Fprintf(&b, "comment #%d by %s  review #%d | hidden: %s  \n", c.DatabaseID, authorDisplay, c.ReviewDatabaseID, reason)
		}
	} else {
		if c.ReplyToID == 0 {
			fmt.Fprintf(&b, "comment #%d by %s  review #%d  \n", c.DatabaseID, authorDisplay, c.ReviewDatabaseID)
		} else {
			fmt.Fprintf(&b, "reply #%d to #%d  by %s  review #%d  \n", c.DatabaseID, c.ReplyToID, authorDisplay, c.ReviewDatabaseID)
		}
	}

	fmt.Fprintf(&b, "_%s_\n", format.Date(c.CreatedAt))
	b.WriteString("\n")

	commentBody := strings.TrimRight(strings.ReplaceAll(c.Body, "\r\n", "\n"), "\n")
	for _, line := range strings.Split(commentBody, "\n") {
		b.WriteString(">" + line + "\n")
	}

	if reactionsStr := format.Reactions(c.Reactions, viewerLogin); reactionsStr != "" {
		b.WriteString(reactionsStr)
	}

	return b.String()
}

// formatLocation builds the location string from thread-level fields.
func formatLocation(t *Thread) string {
	if t.Path == "" {
		return ""
	}
	if t.IsOutdated {
		return formatOutdatedLocation(t)
	}
	return formatCurrentLocation(t)
}

func formatCurrentLocation(t *Thread) string {
	if t.StartLine > 0 && t.Line > 0 && t.StartLine != t.Line {
		return fmt.Sprintf("%s on lines +%d to +%d", t.Path, t.StartLine, t.Line)
	} else if t.Line > 0 {
		return fmt.Sprintf("%s on line +%d", t.Path, t.Line)
	}
	return t.Path
}

func formatOutdatedLocation(t *Thread) string {
	if t.OriginalStartLine > 0 && t.OriginalLine > 0 && t.OriginalStartLine != t.OriginalLine {
		return fmt.Sprintf("%s on original lines %d to %d (outdated)", t.Path, t.OriginalStartLine, t.OriginalLine)
	} else if t.OriginalLine > 0 {
		return fmt.Sprintf("%s on original line %d (outdated)", t.Path, t.OriginalLine)
	}
	return fmt.Sprintf("%s (outdated)", t.Path)
}
