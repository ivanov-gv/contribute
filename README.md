It's mostly a test project to try Claude Code. Completely vibecoded. Use with caution.

---

# contribute

A CLI tool that lets AI agents interact with pull requests as real contributors — reading reviews, posting comments, and leaving reactions.

## TL;DR

```bash
# install
go install github.com/ivanov-gv/contribute/cmd/contribute@latest

# see PR details (auto-detects from current branch)
contribute pr

# list all comments and reviews on a PR
contribute comments

# post a comment
contribute comment "Fixed the issue, please re-review"

# react to a comment
contribute react 123456789 eyes --type issue
contribute react 987654321 rocket --type review

# show inline comments for a specific review
contribute review 3929204495

# show all comments in a thread across reviews (use thread id from review output)
contribute thread 2935138407

# reply to a review comment in-thread
contribute reply 2935138407 "Fixed, thanks"

# resolve a thread
contribute resolve 2935138407

# post an inline review comment
contribute review-comment "Nit: rename this variable" --file internal/cmd/pr.go --line 42

# approve a PR
contribute submit-review --event APPROVE --body "LGTM"
```

All commands auto-detect the repository (from git remote) and PR number (from current branch). Authentication uses a GitHub App — set `GH_CONTRIBUTE_APP_ID` and `GH_CONTRIBUTE_PRIVATE_KEY_PATH` to authenticate automatically on startup.

---

## Why

AI coding agents (Claude Code, Copilot, Cursor, etc.) can write code, commit, and push — but they can't participate in the review process on GitHub. They have no way to:

- Read what reviewers said about their PR
- Acknowledge comments with reactions
- Reply to feedback
- Show progress on addressing review comments

**contribute** bridges this gap. It gives agents a simple CLI interface to the GitHub review workflow, turning them from "push and forget" tools into active PR participants.

## Use Cases

### Remote control through GitHub reviews

A typical agent workflow today:

1. Agent finishes work, commits, pushes, opens a PR
2. **Dead end** — the agent has no idea what happens next

With contribute:

1. Agent finishes work, commits, pushes, opens a PR
2. A reviewer leaves comments and suggestions on the PR
3. Something triggers the agent again (webhook, polling, slash command)
4. Agent runs `contribute comments` to read all review feedback
5. Agent addresses each comment, pushes fixes
6. Agent runs `contribute comment "Addressed all feedback, PTAL"`
7. Repeat until merged

The entire interaction happens through GitHub — no need to access the agent's terminal or UI.

### Live status through reactions

When an agent is processing review comments, nobody on GitHub knows what's happening. With reactions, the agent can broadcast its progress:

1. Agent receives notification about new review comments
2. Runs `contribute comments` to get the list
3. For each comment, the agent:
   - Adds 👀 (`eyes`) reaction — "I'm looking at this"
   - Works on the fix
   - Adds 🚀 (`rocket`) reaction — "Done"
4. When all comments are addressed, posts a summary comment

Everyone watching the PR sees real-time status without leaving GitHub.

### Automated triage and acknowledgment

An agent can periodically check for new comments across PRs and:

- React with 👍 to acknowledge simple suggestions
- React with 😕 (`confused`) to flag comments it doesn't understand
- Post clarifying questions as replies
- Prioritize comments based on reviewer authority

## Commands

### `contribute pr`

Show details about a pull request in human-readable markdown.

```bash
# auto-detect PR from current branch
contribute pr

# specify PR number explicitly
contribute pr 42
```

Output:
```
# test-pr: test pr #1
open, by @ivanov-gv, 1 commit `test-pr` -> `main`, no merge conflict
https://github.com/ivanov-gv/contribute/pull/1

Reviewers:
Assignees: @ivanov-gv
Labels:
Projects:
Milestone:
Issues:

Conversation: 1 comment

---

test description

---
```

### `contribute comments`

List issue comments and reviews on a pull request. Shows reactions with "by you" tracking, hides minimized comments and fully-resolved reviews.

```bash
# all comments and reviews
contribute comments

# specify PR
contribute comments --pr 42

# show a single comment by id
contribute comments 4038597073

# show content of hidden/minimized comments
contribute comments --show-hidden
```

Output:
```
issue #4038597073 by you (@ivanov-gv-ai-helper)
_2026-03-11 11:33:27_

test comment 🚀

(1 🚀)
reactions by you: (1 🚀)

---
issue #4038819817 by @ivanov-gv
_2026-03-11 12:15:54_

> test comment 🚀
test reply

(1 😕)
reactions by you:

---
review #3929204495 by @ivanov-gv
_2026-03-11 12:17:34_

submit review

comments: 3
(1 👀)
reactions by you:

---
review #3929353771 by @ivanov-gv | hidden: Resolved
```

Key features:
- **Issue comments** show id, author, date, body, and reactions
- **Reviews** show id, author, date, body, inline comment count, and reactions
- **"reactions by you"** tracks which reactions belong to the authenticated user (works with GitHub App `[bot]` accounts)
- **Hidden items**: minimized issue comments and reviews (`isMinimized: true`) show only the header line with the reason
- Review inline comments are not expanded — use the review id for detailed inspection

### `contribute comment`

Post a top-level comment on a pull request.

```bash
contribute comment "All review comments have been addressed. Ready for re-review."

contribute comment --pr 42 "Automated analysis complete. Found 3 potential issues."
```

### `contribute react`

Add a reaction to a comment. Use the comment id from the `comments` output.

```bash
# react to a review comment (default)
contribute react 123456789 rocket

# react to a top-level (issue) comment
contribute react 123456789 eyes --type issue
```

Valid reactions: `+1`, `-1`, `laugh`, `confused`, `heart`, `hooray`, `rocket`, `eyes`

### `contribute review`

Show a specific review's inline comments. Only comments belonging to the requested review are shown, grouped by thread. If a comment replies to one from a different review, it is flagged as `(not in this review)` — use `thread` to see the full context.

```bash
# show inline comments for review by id (use id from comments output)
contribute review 3929204495

# include the diff hunk for each thread
contribute review 3929204495 --diff

# specify PR explicitly
contribute review 3929204495 --pr 42
```

Output:
```
# review #3948671120 by you (@ivanov-gv)
_2026-03-14 11:13:03_

Needs fixes.


thread #2935132146  internal/auth/auth.go on original line 22 (outdated)
comment #2935132146 by you (@ivanov-gv)
_2026-03-14 11:13:03_

Name it as GH_CONTIBUTE_TOKEN

---
thread #2935132635  internal/auth/auth.go on original line 25 (outdated)
comment #2935132635 by you (@ivanov-gv)
_2026-03-14 11:13:32_

Use this path instead: .config/contribute/token
```

Cross-review reply example (reply belongs to this review, but replies to a comment from another):
```
thread #2935138407  internal/auth/auth.go on original line 88 (outdated)
reply #2935243067 to #2935138407 (not in this review)  by you (@ivanov-gv)
_2026-03-14 12:37:24_

Yes, you fixed this particular issue on this line in this file, ...
```

### `contribute thread`

Show all comments in a thread across all reviews. Use the thread id from the `review` output (the `thread #ID` header). Each comment is annotated with its review id.

```bash
# show the full thread
contribute thread 2935138407

# specify PR explicitly
contribute thread 2935138407 --pr 42
```

Output:
```
# thread #2935138407  internal/auth/auth.go on original line 88 (outdated)

comment #2935138407 by @ivanov-gv  review #3948671120
_2026-03-14 11:18:37_

No, stderr for errors, stdout for output. Use logging, not a simple printf

---
reply #2935243067 to #2935138407  by @ivanov-gv  review #3948810914
_2026-03-14 12:37:24_

Yes, you fixed this particular issue on this line in this file, ...
```

Typical workflow: `review` for focused reading of one review's feedback; `thread <id>` when you need the full cross-review conversation.

### `contribute reply`

Post a threaded reply to a specific review comment. Use the comment id from the `review` or `thread` output.

```bash
contribute reply 2935138407 "Fixed in the latest commit, thanks for the catch"

contribute reply 2935138407 "Can you clarify what you mean here?" --pr 42
```

### `contribute resolve` / `contribute unresolve`

Resolve or unresolve a review thread. Use the thread id from the `review` output.

```bash
contribute resolve 2935138407
contribute resolve 2935138407 --pr 42

contribute unresolve 2935138407
```

### `contribute review-comment`

Post an inline review comment on a specific file and line.

```bash
contribute review-comment "Nit: rename this to be more descriptive" --file internal/cmd/pr.go --line 42

# specify diff side (RIGHT is default)
contribute review-comment "This was removed" --file internal/cmd/pr.go --line 10 --side LEFT

contribute review-comment "Missing error check" --file internal/cmd/pr.go --line 42 --pr 42
```

### `contribute submit-review`

Submit a review with an event type.

```bash
# approve
contribute submit-review --event APPROVE

# request changes with a body
contribute submit-review --event REQUEST_CHANGES --body "Please address the comments above before merging"

# leave a comment-only review
contribute submit-review --event COMMENT --body "Some thoughts inline, no blocking issues"

# specify PR explicitly
contribute submit-review --event APPROVE --pr 42
```

Valid events: `APPROVE`, `REQUEST_CHANGES`, `COMMENT`

### `contribute issue`

Show issue details.

```bash
contribute issue 42
```

### `contribute issues`

List open issues in the repository.

```bash
# list open issues (default limit: 20)
contribute issues

# filter by label
contribute issues --label bug
contribute issues --label "bug,help wanted"

# increase limit
contribute issues --limit 50
```

### `contribute issue-comment`

Post a comment on an issue.

```bash
contribute issue-comment 42 "Looking into this now"
```

### `contribute watch`

Poll for new activity on a PR and print changes as they appear.

```bash
# poll every 30 seconds (default)
contribute watch

# custom interval
contribute watch --interval 1m

contribute watch --pr 42 --interval 15s
```

### `contribute token`

Print the active GitHub token to stdout. Follows the same priority chain as all other commands. Useful for passing the token to other tools.

```bash
GH_TOKEN=$(contribute token) gh pr view 123
GH_TOKEN=$(contribute token) gh api /user
```

## Installation

### Via go install

```bash
go install github.com/ivanov-gv/contribute/cmd/contribute@latest
```

### From source

```bash
git clone https://github.com/ivanov-gv/contribute.git
cd contribute
go install ./cmd/contribute
```

### Authentication

contribute authenticates as a **GitHub App**. API calls appear as `yourapp[bot]`, giving them proper attribution. The app must be installed on the target repository.

#### Automatic login via environment variables

Set these variables before running any command — contribute authenticates automatically on startup:

```bash
export GH_CONTRIBUTE_APP_ID=3063096
export GH_CONTRIBUTE_PRIVATE_KEY_PATH=/home/vscode/.config/gh-contribute/private-key.pem
# optional: export GH_CONTRIBUTE_INSTALLATION_ID=<id>  # auto-detected if unset
```

Alternatively, supply the PEM key inline as a base64-encoded string (useful in CI where writing a file is inconvenient):

```bash
export GH_CONTRIBUTE_APP_ID=3063096
export GH_CONTRIBUTE_PRIVATE_KEY=$(base64 < /path/to/private-key.pem)
```

`GH_CONTRIBUTE_PRIVATE_KEY` takes priority over `GH_CONTRIBUTE_PRIVATE_KEY_PATH` when both are set.

If neither env vars nor stored credentials are present, contribute exits with a non-zero code and prompts you to authenticate:

```
Error: not authenticated — set GH_CONTRIBUTE_APP_ID and GH_CONTRIBUTE_PRIVATE_KEY_PATH, or run 'contribute login'
```

#### Persisting credentials with contribute login

To store credentials in `~/.config/contribute/app.json` instead of setting env vars every time:

```bash
contribute login --app-id 3063096 --key-path /home/vscode/.config/gh-contribute/private-key.pem
# GH_CONTRIBUTE_APP_ID is read from env if --app-id is omitted
GH_CONTRIBUTE_APP_ID=3063096 contribute login --key-path /home/vscode/.config/gh-contribute/private-key.pem
```

Stored credentials are used when env vars are not set. Env vars always take priority.

In addition to storing credentials, `contribute login` automatically configures git for the app's bot identity:

```
INF login: git credential helper configured for github.com
INF login: git identity configured user.email=115546723+ai-contributor-helper[bot]@users.noreply.github.com user.name=ai-contributor-helper[bot]
INF login: authenticated successfully app=AI contributor helper app_id=3063096
```

This sets:
- `git config --global user.name` → `{app-slug}[bot]`
- `git config --global user.email` → `{installation_id}+{app-slug}[bot]@users.noreply.github.com`

So `git commit` and `git push` work immediately after login with no further setup.

#### Check status

```bash
contribute auth status
# logged in as app: AI contributor helper (app_id=3063096)
```

#### CI and non-interactive environments

In CI, set credentials via env vars (shown above) or use a pre-issued token directly:

```bash
export GH_CONTRIBUTE_TOKEN=github_pat_...
```

`GH_CONTRIBUTE_TOKEN` takes the highest priority — no app credentials needed when it is set.

#### Token lifecycle

Installation tokens expire after 1 hour. contribute automatically refreshes them via the `TokenProvider` — no manual intervention needed. If credentials become invalid, contribute exits with:

```
Error: token invalid or expired — run 'contribute login' to reauthenticate
```

## Auto-detection

When `--pr` is not specified, contribute automatically:

1. Reads the current git branch name
2. Searches for an open PR with that branch as the head
3. Uses the first match

When the repository is not specified (it never needs to be), contribute:

1. Reads the `origin` remote URL from git
2. Parses the owner and repo name from it (supports both SSH and HTTPS remotes)

This means in most cases you just run `contribute comments` with zero flags and it does the right thing.
