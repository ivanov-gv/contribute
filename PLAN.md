# PLAN: From MVP to Production-Ready AI Contributor Tool

## Current State Assessment

gh-contribute is an MVP GitHub CLI extension that lets AI agents read and interact with PRs. It works, but:
- **Zero tests** — no unit tests, no integration tests, no e2e tests
- **Read-heavy** — can read reviews/comments but can't reply to inline review comments or post inline comments
- **No issue support** — can't read GitHub issues to start work autonomously
- **Auth is user-based** — Device Flow authenticates on behalf of a user, not as an independent app/bot
- **No CI** — only a release workflow, no test/lint pipeline
- **No reply-to-review-comment** — the most critical missing write operation for the review workflow
- **No polling/webhook** — agent has no way to know when a new review arrives
- **Duplicated code** — `reactionNode`, `mapReactions`, thread node types are copy-pasted across packages

---

## Phase 1: Foundation — Tests, CI, Refactoring

### 1.1 Extract shared GraphQL types
**Why**: `reactionNode`, `mapReactions`, thread/comment node types are duplicated in `comment/`, `review/`, `thread/`.
**What**:
- Create `internal/model/graphql/` with shared node types: `ReactionNode`, `ThreadCommentNode`, `ReviewThreadNode`
- Move `mapReactions` to a single location
- All services reference shared types

### 1.2 Add interfaces at consumer side
**Why**: Services directly depend on concrete `*githubv4.Client` and `*ghrest.Client`. This makes unit testing impossible without hitting GitHub.
**What**:
- Define interfaces in each service for the operations it needs:
  ```go
  // in service/pr/pr.go
  type graphQLQuerier interface {
      Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
  }
  ```
- Services accept interfaces, not concrete types
- Generate mocks with mockery for each interface

### 1.3 Unit tests for all services
**What** (per service):

| Package | What to test |
|---------|-------------|
| `config` | `parseRemoteURL` — SSH, HTTPS, proxy paths, edge cases. `LoadToken` — env var priority, file fallback, missing file |
| `utils/format` | `ReactionEmoji`, `IsViewer` (with/without `[bot]`), `Author`, `Date`, `EnumLabel`, `Reactions` |
| `service/pr` | `mapPR` — all fields populated, nil milestone, empty lists. `Format` — output matches expected markdown |
| `service/comment` | `Format` — timeline ordering, minimized items, reactions. `FilterByID` — found/not found |
| `service/review` | `collectGroups*` — grouping by thread, external reply detection, sorting. `Format` — output structure |
| `service/thread` | `buildThread` — all fields, nil handling. `Format` — output structure |
| `service/reaction` | `isValid` — all valid reactions, invalid input |

**How**: Use mockery-generated mocks for GraphQL/REST interfaces. Test the service methods with mock responses.

### 1.4 Unit tests for CLI commands
**What**: Test cobra command wiring — correct flags, argument parsing, error messages.
- Test `resolvePR` logic with mocked PR service
- Test `react` command validation of reaction types

### 1.5 CI pipeline
**What**: Add `.github/workflows/ci.yml`:
- Trigger on push and PR to `main`
- Steps: checkout → Go setup → `make lint` → `make test` → `make build`
- Cache Go modules

### 1.6 E2E tests against a private test repo
**Why**: Real API calls catch serialization bugs, permission issues, and GraphQL schema changes.
**What**:
- Create `test/` directory with integration tests
- Use a dedicated test repo (e.g., `ivanov-gv/gh-contribute-test`) with a stable test PR
- Tests authenticate via `GH_CONTRIBUTE_TOKEN` env var
- Test matrix:
  - `pr` — fetch PR details, verify fields
  - `comments` — list comments, verify structure
  - `comment` — post a comment, verify it appears, clean up
  - `react` — add a reaction, verify it appears
  - `review` — fetch review details with inline comments
  - `thread` — fetch thread across reviews
- Run in CI with a repository secret for the token
- Guard with build tag `//go:build integration` so `make test` skips them

---

## Phase 2: Missing Write Operations

### 2.1 Reply to review comments
**Why**: This is THE most important missing feature. When a reviewer leaves inline comments, the agent needs to reply to each one in-thread.
**What**:
- New command: `gh contribute reply <comment-id> <body>`
- Flags: `--pr` (auto-detected)
- Uses REST API: `POST /repos/{owner}/{repo}/pulls/{pull_number}/comments/{comment_id}/replies`
- Add `Reply` method to comment service (or new `reply` service)

### 2.2 Post inline review comments
**Why**: Agent should be able to leave its own review comments on specific lines.
**What**:
- New command: `gh contribute review-comment <body> --file <path> --line <n> [--side RIGHT|LEFT]`
- Uses REST API: `POST /repos/{owner}/{repo}/pulls/{pull_number}/comments`
- Requires `commit_id` (latest commit SHA from PR head)

### 2.3 Submit a review
**Why**: Agent should be able to approve, request changes, or leave a review.
**What**:
- New command: `gh contribute submit-review --event APPROVE|REQUEST_CHANGES|COMMENT [--body "..."]`
- Uses REST API: `POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews`

### 2.4 Resolve/unresolve review threads
**Why**: After addressing a comment, the agent should resolve the thread.
**What**:
- New command: `gh contribute resolve <thread-id>` / `gh contribute unresolve <thread-id>`
- Uses GraphQL mutations: `resolveReviewThread` / `unresolveReviewThread`

---

## Phase 3: Issue-Driven Workflow

### 3.1 Read GitHub issues
**Why**: The agent workflow should start from "here's an issue, go fix it" — not from a human typing a request.
**What**:
- New command: `gh contribute issue <number>` — shows issue details (title, body, labels, assignees, comments)
- New command: `gh contribute issues` — lists open issues, with filters:
  - `--label <name>` — filter by label (e.g., `agent-ready`, `good-first-issue`)
  - `--assignee <login>` — filter by assignee
  - `--limit <n>` — max results
- New service: `internal/service/issue/`
- GraphQL queries for rich issue data

### 3.2 Issue comment operations
**What**:
- `gh contribute issue-comment <issue-number> <body>` — post a comment on an issue
- Reuse existing comment service (issues and PRs share the same comment API)

### 3.3 Link issues to PRs
**What**:
- When creating a PR (via `gh pr create`), the agent can reference `Fixes #N` in the body
- `gh contribute issue` output shows linked PRs so the agent knows if work is already in progress

---

## Phase 4: The AI Agent Workflow Loop

### 4.1 Poll for new activity
**Why**: The agent needs to know when a review arrives. Options:

**Option A: Polling command (simplest, recommended first)**
- New command: `gh contribute watch --pr <number> --interval 30s`
- Polls `comments` endpoint, diffs against last known state
- Outputs new comments/reviews to stdout when detected
- Agent calls this periodically or the command runs in background

**Option B: Webhook server (more complex, better for production)**
- New command: `gh contribute serve --port 8080`
- Registers a webhook URL with the GitHub repo
- Receives `pull_request_review`, `pull_request_review_comment`, `issue_comment` events
- Outputs events to stdout or calls a configured script/command
- Requires public URL (ngrok, cloudflare tunnel, or deployment)

**Recommendation**: Start with Option A. It works anywhere, needs no infrastructure, and is sufficient for the agent workflow. The agent (Claude Code) can use a `/loop` skill or manual polling.

### 4.2 Full workflow skill/script
**Why**: Package the complete workflow as a reusable script the agent can execute.
**Flow**:
```
1. gh contribute issues --label agent-ready --limit 1
   → Pick an issue
2. Read issue details, understand the task
3. Create a branch, write code, commit, push
4. gh pr create (via gh CLI)
5. gh contribute comment "Ready for review"
6. LOOP:
   a. gh contribute comments --pr N
   b. If new review found:
      - gh contribute review <review-id>
      - For each thread:
        - gh contribute react <comment-id> eyes
        - Address the feedback in code
        - gh contribute reply <comment-id> "Fixed in <commit>"
        - gh contribute react <comment-id> rocket
        - gh contribute resolve <thread-id>
      - git commit, push
      - gh contribute comment "All feedback addressed, PTAL"
   c. If approved: done
   d. Sleep, repeat
```

### 4.3 Claude Code integration
**What**:
- A Claude Code skill (in `.claude/skills/`) that wraps the full workflow
- Skill name: `/contribute` or `/review-cycle`
- Takes an issue number or PR number as input
- Handles the full loop: read → code → push → respond to reviews
- Uses `gh contribute` commands internally

---

## Phase 5: App-Level Authentication (Bot Account)

### 5.1 Current auth model
Currently: Device Flow → user access token → all API calls are "app[bot] on behalf of user".
This is fine for human-triggered use, but for an autonomous agent with its own GitHub account, we need a different model.

### 5.2 GitHub App Installation Token auth
**Why**: The agent should authenticate as itself, not on behalf of a user.
**How**:
1. Create a GitHub App for the agent (e.g., "claude-contributor")
2. Install the App on target repositories
3. Auth flow:
   - App has a private key (PEM file)
   - Generate JWT from private key + app ID
   - Exchange JWT for installation token (scoped to specific repos)
   - Use installation token for API calls

**What changes**:
- New auth mode: `--auth-mode app` (default stays `device` for backward compat)
- New config: `GH_CONTRIBUTE_APP_ID`, `GH_CONTRIBUTE_PRIVATE_KEY` (path to PEM file or base64-encoded)
- New auth client: `internal/client/auth/app.go`
  - `GenerateJWT(appID int64, privateKey []byte) (string, error)`
  - `GetInstallationToken(jwt string, installationID int64) (string, error)`
- Token refresh: installation tokens expire after 1 hour — need automatic refresh
- **Secret storage**: The private key should NEVER be compiled into the binary. It should be:
  - Environment variable: `GH_CONTRIBUTE_PRIVATE_KEY` (base64-encoded PEM)
  - File path: `GH_CONTRIBUTE_PRIVATE_KEY_PATH`
  - In production: injected by the container runtime (Docker secret, K8s secret, cloud secret manager)

### 5.3 Token lifecycle management
**What**:
- Installation tokens expire in 1 hour
- Add a `tokenProvider` interface that handles refresh transparently
- Cache token, refresh 5 minutes before expiry
- All services use the token provider instead of a raw string

### 5.4 Dual auth support
**What**: Support both auth modes simultaneously:
```
# User mode (existing)
gh contribute auth login

# App mode (new)
export GH_CONTRIBUTE_APP_ID=12345
export GH_CONTRIBUTE_PRIVATE_KEY_PATH=/path/to/key.pem
gh contribute auth status
# → Authenticated as app: claude-contributor (installation on ivanov-gv/gh-contribute)
```

Config priority:
1. `GH_CONTRIBUTE_TOKEN` env var (explicit token, any source)
2. App credentials (`APP_ID` + `PRIVATE_KEY`) → generate installation token
3. `~/.config/gh-contribute/token` file (Device Flow token)

---

## Phase 6: Production Hardening

### 6.1 Pagination
**Why**: Current GraphQL queries use `first: 100` without pagination. PRs with 100+ comments will silently lose data.
**What**:
- Add cursor-based pagination to all GraphQL queries
- Helper function: `paginateQuery(client, query, variables, pageSize, appendFn)`

### 6.2 Rate limiting
**What**:
- Respect GitHub API rate limits (5000/hr for REST, point-based for GraphQL)
- Add rate limit headers to response handling
- Log warnings when approaching limits
- Add exponential backoff on 403 rate limit responses

### 6.3 Structured error types
**What**:
- Replace raw `fmt.Errorf` with typed errors for common cases:
  - `ErrNotFound` (PR, review, thread, comment not found)
  - `ErrRateLimited` (with retry-after)
  - `ErrPermissionDenied`
  - `ErrTokenExpired`
- Consumers can handle each case differently

### 6.4 Output formats
**What**:
- Add `--format json` flag to all read commands
- JSON output for machine consumption (by other tools, scripts, agents)
- Markdown remains default for human/agent readability

### 6.5 Logging improvements
**What**:
- Add `--verbose` / `-v` flag for debug logging
- Log all API calls at debug level (method, URL, response time)
- Log token source at startup (env var vs file vs app)

---

## Execution Order

| Priority | Phase | Effort | Impact |
|----------|-------|--------|--------|
| **P0** | 1.2 Interfaces | S | Unblocks all testing |
| **P0** | 1.3 Unit tests | M | Correctness guarantee |
| **P0** | 1.5 CI pipeline | S | Prevents regressions |
| **P1** | 2.1 Reply to review comments | S | Completes the review workflow |
| **P1** | 2.4 Resolve threads | S | Completes the review workflow |
| **P1** | 3.1 Read issues | M | Enables issue-driven workflow |
| **P1** | 4.1 Polling | M | Enables async workflow |
| **P2** | 1.1 Extract shared types | S | Code quality |
| **P2** | 1.6 E2E tests | M | API contract validation |
| **P2** | 2.2-2.3 Inline comments, submit review | M | Advanced review features |
| **P2** | 5.2-5.4 App auth | L | Independent bot identity |
| **P3** | 6.1-6.5 Production hardening | L | Scale and reliability |
| **P3** | 4.2-4.3 Workflow skill | M | Full autonomy |
| **P3** | 3.2-3.3 Issue comments, linking | S | Completes issue workflow |

S = Small (hours), M = Medium (1-2 days), L = Large (3+ days)

---

## Architecture Decisions

### Why not compile secrets into the binary?
The private key is a credential. Embedding it in the binary means:
- Anyone who gets the binary can extract it
- You can't rotate it without rebuilding and redeploying
- It violates 12-factor app principles
- It's a supply chain risk if the binary is distributed

**Instead**: Use environment variables or file paths. The binary reads the secret at runtime.

### Why polling over webhooks (initially)?
- Webhooks need a public URL, which means infrastructure
- Polling works in any environment (local, CI, containers)
- For a single-agent use case, polling every 30s is fine
- Webhooks can be added later as an optimization

### Why interfaces at consumer side?
- Go convention: define interfaces where they're used
- Each service declares exactly the methods it needs
- Mocks are trivially generated
- No coupling between services through shared interfaces

### Why separate issue service (not reuse comment service)?
- Issues have different fields (labels, assignees, milestone, state, linked PRs)
- Issue listing with filters is a distinct concern
- Comment posting can share the REST client, but the query layer is different

---

## Files to Create/Modify

### New Files
```
.github/workflows/ci.yml                          — CI pipeline
internal/service/pr/pr_test.go                     — PR service unit tests
internal/service/pr/format_test.go                 — PR format tests
internal/service/comment/comment_test.go           — Comment service unit tests
internal/service/comment/format_test.go            — Comment format tests
internal/service/review/review_test.go             — Review service unit tests
internal/service/review/format_test.go             — Review format tests
internal/service/thread/thread_test.go             — Thread service unit tests
internal/service/thread/format_test.go             — Thread format tests
internal/service/reaction/reaction_test.go         — Reaction service unit tests
internal/config/config_test.go                     — Config unit tests
internal/config/token_test.go                      — Token unit tests
internal/utils/format/format_test.go               — Format utils unit tests
internal/cmd/reply.go                              — Reply to review comment command
internal/service/issue/issue.go                    — Issue service
internal/service/issue/format.go                   — Issue formatting
internal/cmd/issue.go                              — Issue commands
test/e2e_test.go                                   — E2E tests
.mockery.yaml                                      — Mockery configuration
```

### Modified Files
```
internal/service/pr/pr.go                          — Add interface for GraphQL client
internal/service/comment/comment.go                — Add interface, add Reply method
internal/service/review/review.go                  — Add interface
internal/service/thread/thread.go                  — Add interface
internal/service/reaction/reaction.go              — Add interface
internal/cmd/root.go                               — Wire new commands
internal/config/config.go                          — Support app auth mode
internal/config/token.go                           — Support app token generation
internal/client/auth/auth.go                       — Add app auth flow
```
