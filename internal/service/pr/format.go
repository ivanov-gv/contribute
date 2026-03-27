package pr

import (
	"fmt"
	"strings"
)

// Format renders PR info as human-readable markdown
func (info *Info) Format() string {
	var b strings.Builder

	// header: title only (no # heading prefix)
	b.WriteString(fmt.Sprintf("%s #%d\n", info.Title, info.Number))

	// meta line: state, author, commits, branches, then either files-changed (merged) or merge status (open)
	state := info.State
	if info.IsDraft {
		state = "draft"
	}
	commitWord := "commits"
	if info.CommitCount == 1 {
		commitWord = "commit"
	}
	var statusSuffix string
	if info.State == "merged" {
		statusSuffix = fmt.Sprintf("%d files changed, lines +%d -%d", info.ChangedFiles, info.Additions, info.Deletions)
	} else if info.Mergeable == "CONFLICTING" {
		statusSuffix = "merge conflict"
	} else {
		statusSuffix = "no merge conflict"
	}
	b.WriteString(fmt.Sprintf("%s, by @%s, %d %s `%s` -> `%s`, %s\n",
		state, info.Author, info.CommitCount, commitWord, info.Head, info.Base, statusSuffix))

	// url
	b.WriteString(info.URL + "\n")
	b.WriteString("\n")

	// metadata fields
	b.WriteString(fmt.Sprintf("Reviewers: %s  \n", strings.Join(info.Reviewers, ", ")))
	b.WriteString(fmt.Sprintf("Assignees: %s  \n", strings.Join(info.Assignees, ", ")))
	b.WriteString(fmt.Sprintf("Labels: %s  \n", strings.Join(info.Labels, ", ")))
	b.WriteString(fmt.Sprintf("Projects: %s  \n", strings.Join(info.Projects, ", ")))
	b.WriteString(fmt.Sprintf("Milestone: %s  \n", info.Milestone))

	// linked issues
	var issueStrs []string
	for _, i := range info.Issues {
		issueStrs = append(issueStrs, fmt.Sprintf("#%d %s", i.Number, i.Title))
	}
	b.WriteString(fmt.Sprintf("Issues: %s  \n", strings.Join(issueStrs, ", ")))

	// conversation count, with locked indicator when applicable
	commentWord := "comments"
	if info.CommentCount == 1 {
		commentWord = "comment"
	}
	conversation := fmt.Sprintf("%d %s", info.CommentCount, commentWord)
	if info.IsLocked {
		conversation += ", locked"
	}
	b.WriteString(fmt.Sprintf("\nConversation: %s\n", conversation))

	// description — each line quoted with > so it reads as a blockquote
	b.WriteString("\n---\n\n")
	body := strings.TrimSpace(strings.ReplaceAll(info.Body, "\r\n", "\n"))
	if body == "" {
		body = "No description provided."
	}
	for _, line := range strings.Split(body, "\n") {
		b.WriteString(">" + line + "\n")
	}
	b.WriteString("\n---\n")

	return b.String()
}
