---
name: review-cycle
description: Full AI agent workflow — pick an issue, code, push, create PR, respond to reviews
user_invocable: true
---

# Review Cycle Skill

You are an AI agent working on a GitHub repository using `gh-contribute`.

## Usage

```
/review-cycle <issue-number>       — Start from an issue
/review-cycle --pr <pr-number>     — Resume on an existing PR
```

If no argument is given, auto-detect the PR from the current branch.

## Full Workflow (from issue)

### 1. Pick and read the issue

```bash
gh-contribute issue <number>
```

Understand the task from the issue body, comments, and labels.

### 2. Implement the fix

1. Create a branch: `git checkout -b fix/<issue-number>-<short-description>`
2. Write code, run tests (`make test`), lint (`make lint`)
3. Commit: `git commit -m "Fix #<number>: <description>"`
4. Push: `git push -u origin <branch>`

### 3. Create a PR

Use `gh pr create` with `Fixes #<number>` in the body, then notify:

```bash
gh-contribute comment "Ready for review" --pr <N>
```

### 4. Enter the review loop

## Review Loop (for new or existing PRs)

### 1. Read the current PR state

```bash
gh-contribute comments --pr <N>
```

Look for new reviews (CHANGES_REQUESTED or COMMENTED) that have not been addressed.

### 2. For each unaddressed review

```bash
gh-contribute review <review-id> --pr <N>
```

Read all inline threads. Understand the overall feedback before making changes.

### 3. For each thread in the review

1. **Acknowledge** — react with eyes to signal you've seen it:
   ```bash
   gh-contribute react <comment-id> EYES
   ```

2. **Understand** — read the comment body and the file/line context.

3. **Fix** — make the requested code change.

4. **Reply** — tell the reviewer what you did:
   ```bash
   gh-contribute reply <comment-id> "Fixed in <short-sha> — <what you changed>"
   ```

5. **React** — signal completion:
   ```bash
   gh-contribute react <comment-id> ROCKET
   ```

6. **Resolve** — mark the thread as resolved:
   ```bash
   gh-contribute resolve <thread-id> --pr <N>
   ```

### 4. Push and notify

```bash
git push
gh-contribute comment "All feedback addressed, PTAL" --pr <N>
```

### 5. Check for approval

```bash
gh-contribute comments --pr <N>
```

If the latest review is APPROVED, the cycle is complete. Otherwise, wait for the next review.

## Commands Reference

| Command | Description |
|---------|-------------|
| `gh-contribute issue <n>` | Read issue details |
| `gh-contribute issues [--label <l>]` | List open issues |
| `gh-contribute pr [n]` | Show PR details |
| `gh-contribute comments [id] --pr <n>` | List comments or show one by ID |
| `gh-contribute review <id> --pr <n>` | Show review with inline comments |
| `gh-contribute thread <id> --pr <n>` | Show a review thread |
| `gh-contribute comment <body> --pr <n>` | Post a top-level PR comment |
| `gh-contribute reply <id> <body> --pr <n>` | Reply to a review comment |
| `gh-contribute react <id> <emoji>` | Add reaction (EYES, ROCKET, +1, etc.) |
| `gh-contribute resolve <id> --pr <n>` | Resolve a review thread |
| `gh-contribute unresolve <id> --pr <n>` | Unresolve a review thread |
| `gh-contribute review-comment <body> --file <f> --line <l> --pr <n>` | Post inline comment |
| `gh-contribute submit-review --event <e> --pr <n>` | Submit review (APPROVE/REQUEST_CHANGES/COMMENT) |
| `gh-contribute issue-comment <n> <body>` | Comment on an issue |
| `gh-contribute watch --pr <n>` | Poll for new activity |

## Important Notes

- Always read the full review before making changes — understand the overall feedback first.
- Group related fixes into a single commit when possible.
- If a comment is unclear, reply asking for clarification instead of guessing.
- Never force-push during a review cycle — it breaks the review context.
- Use `gh-contribute thread <thread-id>` to see the full conversation history of a specific thread.
