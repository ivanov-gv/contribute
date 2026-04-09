It's mostly a test project to try Claude Code. Completely vibecoded. Use with caution.

---

# gh-contribute

A GitHub CLI extension that lets AI agents interact with pull requests as real contributors — reading reviews, posting comments, and leaving reactions.

## TL;DR

```bash
# install
gh extension install ivanov-gv/gh-contribute

# see PR details (auto-detects from current branch)
gh contribute pr

# list all comments and reviews on a PR
gh contribute comments

# post a comment
gh contribute comment "Fixed the issue, please re-review"

# react to a comment
gh contribute react 123456789 eyes --type issue
gh contribute react 987654321 rocket --type review

# show inline comments for a specific review
gh contribute review 3929204495

# show all comments in a thread across reviews (use thread id from review output)
gh contribute thread 2935138407
```

All commands auto-detect the repository (from git remote) and PR number (from current branch). Authentication uses a GitHub App — set `GH_CONTRIBUTE_APP_ID` and `GH_CONTRIBUTE_PRIVATE_KEY_PATH` to authenticate automatically on startup.

---

## Why

AI coding agents (Claude Code, Copilot, Cursor, etc.) can write code, commit, and push — but they can't participate in the review process on GitHub. They have no way to:

- Read what reviewers said about their PR
- Acknowledge comments with reactions
- Reply to feedback
- Show progress on addressing review comments

**gh-contribute** bridges this gap. It gives agents a simple CLI interface to the GitHub review workflow, turning them from "push and forget" tools into active PR participants.

## Use Cases

### Remote control through GitHub reviews

A typical agent workflow today:

1. Agent finishes work, commits, pushes, opens a PR
2. **Dead end** — the agent has no idea what happens next

With gh-contribute:

1. Agent finishes work, commits, pushes, opens a PR
2. A reviewer leaves comments and suggestions on the PR
3. Something triggers the agent again (webhook, polling, slash command)
4. Agent runs `gh contribute comments` to read all review feedback
5. Agent addresses each comment, pushes fixes
6. Agent runs `gh contribute comment "Addressed all feedback, PTAL"`
7. Repeat until merged

The entire interaction happens through GitHub — no need to access the agent's terminal or UI.

### Live status through reactions

When an agent is processing review comments, nobody on GitHub knows what's happening. With reactions, the agent can broadcast its progress:

1. Agent receives notification about new review comments
2. Runs `gh contribute comments` to get the list
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

### `gh contribute pr`

Show details about a pull request in human-readable markdown.

```bash
# auto-detect PR from current branch
gh contribute pr

# specify PR number explicitly
gh contribute pr 42
```

Output:
```
# test-pr: test gh extension #1
open, by @ivanov-gv, 1 commit `test-pr` -> `main`, no merge conflict
https://github.com/ivanov-gv/gh-contribute/pull/1

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

### `gh contribute comments`

List issue comments and reviews on a pull request. Shows reactions with "by you" tracking, hides minimized comments and fully-resolved reviews.

```bash
# all comments and reviews
gh contribute comments

# specify PR
gh contribute comments --pr 42
```

Output:
```
issue #4038597073 by you (@ivanov-gv-ai-helper)
_2026-03-11 11:33:27_

test comment from gh-contribute 🚀

(1 🚀)
reactions by you: (1 🚀)

---
issue #4038819817 by @ivanov-gv
_2026-03-11 12:15:54_

> test comment from gh-contribute 🚀
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

### `gh contribute comment`

Post a top-level comment on a pull request.

```bash
gh contribute comment "All review comments have been addressed. Ready for re-review."

gh contribute comment --pr 42 "Automated analysis complete. Found 3 potential issues."
```

### `gh contribute react`

Add a reaction to a comment. Use the comment id from the `comments` output.

```bash
# react to a review comment (default)
gh contribute react 123456789 rocket

# react to a top-level (issue) comment
gh contribute react 123456789 eyes --type issue
```

Valid reactions: `+1`, `-1`, `laugh`, `confused`, `heart`, `hooray`, `rocket`, `eyes`

### `gh contribute review`

Show a specific review's inline comments. Only comments belonging to the requested review are shown, grouped by thread. If a comment replies to one from a different review, it is flagged as `(not in this review)` — use `thread` to see the full context.

```bash
# show inline comments for review by id (use id from comments output)
gh contribute review 3929204495

# include the diff hunk for each thread
gh contribute review 3929204495 --diff

# specify PR explicitly
gh contribute review 3929204495 --pr 42
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

Use this path instead: .config/gh-contribute/token
```

Cross-review reply example (reply belongs to this review, but replies to a comment from another):
```
thread #2935138407  internal/auth/auth.go on original line 88 (outdated)
reply #2935243067 to #2935138407 (not in this review)  by you (@ivanov-gv)
_2026-03-14 12:37:24_

Yes, you fixed this particular issue on this line in this file, ...
```

### `gh contribute thread`

Show all comments in a thread across all reviews. Use the thread id from the `review` output (the `thread #ID` header). Each comment is annotated with its review id.

```bash
# show the full thread
gh contribute thread 2935138407

# specify PR explicitly
gh contribute thread 2935138407 --pr 42
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

## Installation

### From GitHub releases

```bash
gh extension install ivanov-gv/gh-contribute
```

### From source

```bash
git clone https://github.com/ivanov-gv/gh-contribute.git
cd gh-contribute
go build -o gh-contribute ./cmd/gh-contribute
```

Then either:
- Add the binary to your `PATH`, or
- Symlink it into `~/.local/share/gh/extensions/gh-contribute/`

### Authentication

gh-contribute authenticates as a **GitHub App**. API calls appear as `yourapp[bot]`, giving them proper attribution. The app must be installed on the target repository.

#### Automatic login via environment variables

Set these variables before running any command — gh-contribute authenticates automatically on startup:

```bash
export GH_CONTRIBUTE_APP_ID=123456
export GH_CONTRIBUTE_PRIVATE_KEY_PATH=/path/to/private-key.pem
# optional: export GH_CONTRIBUTE_INSTALLATION_ID=<id>  # auto-detected if unset
```

If neither env vars nor stored credentials are present, gh-contribute exits with a non-zero code and prompts you to authenticate:

```
Error: not authenticated — set GH_CONTRIBUTE_APP_ID and GH_CONTRIBUTE_PRIVATE_KEY_PATH, or run 'gh contribute auth login-app'
```

#### Persisting credentials with login-app

To store credentials in `~/.config/gh-contribute/app.json` instead of setting env vars every time:

```bash
gh contribute auth login-app --app-id 123456 --key-path /path/to/private-key.pem
# GH_CONTRIBUTE_APP_ID is read from env if --app-id is omitted
GH_CONTRIBUTE_APP_ID=123456 gh contribute auth login-app --key-path /path/to/private-key.pem
```

Stored credentials are used when env vars are not set. Env vars always take priority.

#### Check status

```bash
gh contribute auth status
# logged in as app: MyApp (app_id=123456)
```

#### CI and non-interactive environments

In CI, set credentials via env vars (shown above) or use a pre-issued token directly:

```bash
export GH_CONTRIBUTE_TOKEN=github_pat_...
```

`GH_CONTRIBUTE_TOKEN` takes the highest priority — no app credentials needed when it is set.

#### Token lifecycle

Installation tokens expire after 1 hour. gh-contribute automatically refreshes them via the `TokenProvider` — no manual intervention needed. If credentials become invalid, gh-contribute exits with:

```
Error: token invalid or expired — run 'gh contribute auth login-app' to reauthenticate
```

## Auto-detection

When `--pr` is not specified, gh-contribute automatically:

1. Reads the current git branch name
2. Searches for an open PR with that branch as the head
3. Uses the first match

When the repository is not specified (it never needs to be), gh-contribute:

1. Reads the `origin` remote URL from git
2. Parses the owner and repo name from it (supports both SSH and HTTPS remotes)

This means in most cases you just run `gh contribute comments` with zero flags and it does the right thing.

## Project Structure

```
gh-contribute/
├── cmd/gh-contribute/main.go           # entry point
├── internal/
│   ├── client/
│   │   ├── auth/                       # GitHub App authentication
│   │   │   ├── app.go                  # JWT generation, installation token exchange, LoadAppConfig
│   │   │   ├── provider.go             # TokenProvider — thread-safe automatic token refresh
│   │   │   └── errors.go               # ErrTokenInvalid sentinel
│   │   ├── git/git.go                  # git helpers (current branch, remote URL)
│   │   └── github/graphql.go           # GraphQL client (queries)
│   ├── cmd/                            # cobra command definitions
│   │   ├── root.go                     # root command, dependency wiring, PersistentPreRunE
│   │   ├── auth.go                     # auth login-app / auth status commands
│   │   ├── pr.go                       # pr command + PR auto-detection
│   │   ├── comments.go                 # comments command
│   │   ├── comment.go                  # comment command (post)
│   │   ├── react.go                    # react command
│   │   ├── review.go                   # review command (inline comment detail)
│   │   └── thread.go                   # thread command (full thread across reviews)
│   ├── config/
│   │   ├── config.go                   # Config struct, Load(), repo detection from git remote
│   │   ├── app.go                      # LoadAppConfig(), SaveAppCredentials(), stored app.json
│   │   ├── token.go                    # loadTokenWithProvider(), GH_CONTRIBUTE_TOKEN env priority
│   │   └── errors.go                   # ErrNotAuthenticated sentinel
│   └── service/
│       ├── pr/                         # PR info and formatting
│       ├── comment/                    # list via GraphQL, post via REST
│       ├── reaction/                   # add reactions via REST
│       ├── review/                     # review detail — inline comments grouped by thread
│       ├── thread/                     # full thread across all reviews via GraphQL
│       └── issue/                      # issue info and formatting
├── .claude/
│   ├── hooks/session-start.sh          # SessionStart hook: build + auth check
│   └── settings.json                   # Claude Code hook registration
├── go.mod
└── go.sum
```

Built with:
- [google/go-github](https://github.com/google/go-github) — GitHub REST API client (mutations)
- GitHub GraphQL API v4 — for rich read queries (reactions, review threads, metadata)
- [spf13/cobra](https://github.com/spf13/cobra) — CLI framework
- [joho/godotenv](https://github.com/joho/godotenv) — `.env` file loading
- [rs/zerolog](https://github.com/rs/zerolog) — structured logging

### Claude Code on the web

The `.claude/hooks/session-start.sh` hook runs automatically at the start of every remote Claude Code session. It:

1. Runs `go mod download` to warm the module cache
2. Builds the extension binary
3. Checks `auth status` — if `GH_CONTRIBUTE_APP_ID` and `GH_CONTRIBUTE_PRIVATE_KEY_PATH` are set, authentication is already active; otherwise the session exits with a clear error

This ensures the agent always has valid GitHub credentials before it needs them, preventing mid-task authentication interruptions.

## Ways to Improve

- **Reply to review comments** — post threaded replies to specific inline comments
- **Diff-aware new comments** — post inline review comments on specific files and lines
- **Webhook listener** — built-in server that watches for review events and triggers agent actions
- **Multi-PR support** — list and manage comments across all open PRs in a repo
- **Token refresh** — currently tokens are non-expiring; if GitHub App expiration is enabled, add a refresh flow
