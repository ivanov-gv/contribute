package issue

import (
	"fmt"
	"strings"
)

// Format renders issue info as human-readable markdown
func (info *Info) Format() string {
	var b strings.Builder

	// header
	fmt.Fprintf(&b, "# %s #%d\n", info.Title, info.Number)

	// meta line
	state := strings.ToLower(info.State)
	fmt.Fprintf(&b, "%s, by @%s\n", state, info.Author)
	b.WriteString(info.URL + "\n\n")

	// metadata fields
	fmt.Fprintf(&b, "Labels: %s  \n", strings.Join(info.Labels, ", "))
	fmt.Fprintf(&b, "Assignees: %s  \n", strings.Join(info.Assignees, ", "))

	// linked PRs
	if len(info.LinkedPRs) > 0 {
		var prStrs []string
		for _, pr := range info.LinkedPRs {
			prStrs = append(prStrs, fmt.Sprintf("#%d %s (%s)", pr.Number, pr.Title, strings.ToLower(pr.State)))
		}
		fmt.Fprintf(&b, "Linked PRs: %s  \n", strings.Join(prStrs, ", "))
	}

	// comment count
	commentWord := "comments"
	if info.CommentCount == 1 {
		commentWord = "comment"
	}
	fmt.Fprintf(&b, "\nConversation: %d %s\n", info.CommentCount, commentWord)

	// body
	b.WriteString("\n---\n\n")
	body := strings.TrimSpace(info.Body)
	if body == "" {
		body = "No description provided."
	}
	b.WriteString(body + "\n")
	b.WriteString("\n---\n")

	// comments
	if len(info.Comments) > 0 {
		b.WriteString("\n")
		for _, c := range info.Comments {
			fmt.Fprintf(&b, "comment #%d by @%s  \n", c.DatabaseID, c.Author)
			date := strings.TrimSuffix(c.CreatedAt, "Z")
			date = strings.Replace(date, "T", " ", 1)
			fmt.Fprintf(&b, "_%s_\n\n", date)
			b.WriteString(c.Body + "\n")
			b.WriteString("\n---\n")
		}
	}

	return b.String()
}

// FormatList renders a list of issues as human-readable markdown
func FormatList(items []ListItem) string {
	if len(items) == 0 {
		return "No open issues found.\n"
	}

	var b strings.Builder
	for _, item := range items {
		labels := ""
		if len(item.Labels) > 0 {
			labels = " [" + strings.Join(item.Labels, ", ") + "]"
		}
		comments := ""
		if item.Comments > 0 {
			comments = fmt.Sprintf(" (%d comments)", item.Comments)
		}
		fmt.Fprintf(&b, "#%d  %s%s  by @%s%s\n", item.Number, item.Title, labels, item.Author, comments)
	}
	return b.String()
}
