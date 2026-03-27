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

### 1.1 Extract shared GraphQL types ✅
**Why**: `reactionNode`, `mapReactions`, thread/comment node types are duplicated in `comment/`, `review/`, `thread/`.
**What**:
- Create `internal/model/graphql/` with shared node types: `ReactionNode`, `ThreadCommentNode`, `ReviewThreadNode`
- Move `mapReactions` to a single location
- All services reference shared types

### 1.2 Add interfaces at consumer side ✅
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

### 1.3 Unit tests for all services ✅
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

### 1.4 Unit tests for CLI commands ✅
**What**: Test cobra command wiring — correct flags, argument parsing, error messages.
- Test `resolvePR` logic with mocked PR service
- Test `react` command validation of reaction types

### 1.5 CI pipeline ✅
**What**: Add `.github/workflows/ci.yml`:
- Trigger on push and PR to `main`
- Steps: checkout → Go setup → `make lint` → `make test` → `make build`
- Cache Go modules

### 1.6 E2E tests against PR #1 in ivanov-gv/gh-contribute ✅

**Why**: Real API calls catch serialization bugs, permission issues, and GraphQL schema changes.
**Test data**: PR #1 is a stable, locked PR with known expected outputs stored in `test/ivanov-gv.gh-contribute.pr#1/`.
**Auth**: Tests authenticate via `GH_CONTRIBUTE_TOKEN` env var.
**Guard**: `//go:build integration` so `make test` skips them. Run with: `go test -tags integration -count=1 -race ./test/...`

**File naming convention**:
- `pr-description.md` → expected output of `gh-contribute pr 1`
- `comments.md` / `comments-unhidden.md` → expected output of `gh-contribute comments --pr 1` (default / `--show-hidden`)
- `N-comments-ID.md` / `N-comments-ID-unhidden.md` → expected output of `gh-contribute comments ID --pr 1`
- `N-review-ID.md` / `N-review-ID-unhidden.md` → expected output of `gh-contribute review ID --pr 1`
- `thread-ID.md` / `thread-ID-unhidden.md` → expected output of `gh-contribute thread ID --pr 1`

**Test cases** (each compares CLI stdout to the corresponding `.md` file):

| # | Command | Expected output file | What it tests |
|---|---------|---------------------|---------------|
| 1 | `pr 1` | `pr-description.md` | PR metadata: title, state, reviewers, labels, linked issues, conversation count |
| 2 | `comments --pr 1` | `comments.md` | Full timeline with hidden items collapsed, reactions, "by you" tracking |
| 3 | `comments --pr 1 --show-hidden` | `comments-unhidden.md` | Full timeline with all hidden items expanded (dates, bodies, reactions) |
| 4 | `comments 4038597073 --pr 1` | `1-comments-4038597073.md` | Single hidden/resolved issue comment |
| 5 | `comments 4038597073 --pr 1 --show-hidden` | `1-comments-4038597073-unhidden.md` | Same comment with hidden content shown |
| 6 | `comments 4038819817 --pr 1` | `2-comments-4038819817.md` | Comment by viewer ("you"), resolved |
| 7 | `comments 4039142865 --pr 1` | `5-comments-4039142865.md` | Comment mid-timeline |
| 8 | `comments 4039221478 --pr 1` | `6-comments-4039221478.md` | Unresolved comment with markdown body |
| 9 | `comments 4039593663 --pr 1` | `8-comments-4039593663.md` | Comment with markdown list |
| 10 | `comments 4041153603 --pr 1` | `10-comments-4041153603.md` | Comment with markdown headings and lists |
| 11 | `comments 4042410800 --pr 1` | `11-comments-4042410800.md` | Comment with eyes emoji reaction |
| 12 | `comments 4067633036 --pr 1` | `12-comments-4067633036.md` | Comment by viewer, no reactions |
| 13 | `review 3929204495 --pr 1` | `3-review-3929204495.md` | Hidden/resolved review with 2 threads (1 unresolved, 1 resolved) |
| 14 | `review 3929204495 --pr 1 --show-hidden` | `3-review-3929204495-unhidden.md` | Same review with hidden thread content |
| 15 | `review 3929240428 --pr 1` | `3-3.2.1-review-3929240428.md` | Review with reply-only thread (no description), resolved thread |
| 16 | `review 3929240428 --pr 1 --show-hidden` | `3-3.2.1-review-3929240428-unhidden.md` | Same with hidden reply expanded |
| 17 | `review 3929353771 --pr 1` | `4-review-3929353771.md` | Resolved review with confused emoji |
| 18 | `review 3929353771 --pr 1 --show-hidden` | `4-review-3929353771-unhidden.md` | Same with hidden resolved comment |
| 19 | `review 3929758963 --pr 1` | `7-review-3929758963.md` | Large review with code blocks, reactions, long markdown |
| 20 | `review 3930039277 --pr 1` | `9-review-3930039277.md` | Review with 3 comments: 1 own thread + 2 cross-review replies |
| 21 | `review 3930039277 --pr 1 --show-hidden` | `9-review-3930039277-unhidden.md` | Same with all content |
| 22 | `thread 2918002761 --pr 1` | `thread-2918002761.md` | Single-comment unresolved thread |
| 23 | `thread 2918002761 --pr 1 --show-hidden` | `thread-2918002761-unhidden.md` | Same (no hidden content, should match) |
| 24 | `thread 2918006660 --pr 1` | `thread-2918006660.md` | Resolved thread with reply from different review |
| 25 | `thread 2918006660 --pr 1 --show-hidden` | `thread-2918006660-unhidden.md` | Same with hidden resolved reply expanded |

**Test structure** (`test/e2e_test.go`):
```go
//go:build integration

func TestE2E(t *testing.T) {
    // build binary once
    // for each test case:
    //   t.Run(name, func(t *testing.T) {
    //     run command, capture stdout
    //     read expected file
    //     assert.Equal(t, expected, actual)
    //   })
}
```

**Notes**:
- The `-unhidden` variant tests `--show-hidden` flag behavior
- PR #1 is locked so comments/reactions won't change
- Tests verify exact string match — any formatting change breaks a test, which is the point

### 1.7 Integration tests ← NEW

See **Phase 7** below for the full plan.

---

## Phase 2: Missing Write Operations

### 2.1 Reply to review comments ✅
### 2.2 Post inline review comments ✅
### 2.3 Submit a review ✅
### 2.4 Resolve/unresolve review threads ✅

---

## Phase 3: Issue-Driven Workflow

### 3.1 Read GitHub issues ✅
### 3.2 Issue comment operations ✅
### 3.3 Link issues to PRs ✅

---

## Phase 4: The AI Agent Workflow Loop

### 4.1 Poll for new activity ✅
### 4.2 Full workflow skill/script ✅
### 4.3 Claude Code integration ✅

---

## Phase 5: App-Level Authentication (Bot Account)

### 5.1 Current auth model ✅
### 5.2 Replace Device Flow with GitHub App Installation Token auth ✅
### 5.3 Token lifecycle management ✅

---

## Phase 6: Production Hardening

### 6.1 Pagination ✅
### 6.2 Rate limiting ✅
### 6.3 Structured error types ✅
### 6.4 Output formats ✅
### 6.5 Logging improvements ✅

---

## Phase 7: Integration Tests

### Overview

**Current test coverage**:
- **Unit tests** (`internal/**/*_test.go`): test individual functions with mocked interfaces — fast, isolated
- **E2E tests** (`test/e2e_test.go`): test the compiled binary against real GitHub API — requires token, binary compilation

**Gap**: nothing tests that real services wired with real HTTP clients produce correct results at the service layer. Unit tests mock at the interface level (skip HTTP serialization). E2E tests go through the compiled binary (can't call service methods directly, can't test error paths or edge cases).

**Integration tests fill this gap**: wire real service instances with real `githubv4.Client` / `ghrest.Client` pointed at the real GitHub API, call service methods directly, and compare `Format()` output against the same expected `.md` files used by E2E tests. The mock server (`testserver`) is only used for edge cases that cannot be tested against real data: HTTP error paths (4xx/5xx), auth failures, and write operations that would mutate state.

### Guidelines compliance

Per `~/.claude/shared/guidelines/code/golang/testing.md`:
- Use `github.com/stretchr/testify/suite` for test suites and dependency setup
- Each test must be independently executable via `go test -run "SuiteName/TestName"`
- Prefer real instances of dependencies over mocks
- Debugger with breakpoints must be available during test runs
- No garbage left after tests (no unattended processes, containers, etc.)

Per `~/.claude/shared/guidelines/code/general/testing.md`:
- Integration tests must ensure different units of the system work together with no issues
- Prefer real instances of dependencies over mocks — mocks hide real failure modes

### 7.1 Two test suites

**Suite** (`//go:build integration` — real GitHub API):
- Authenticates with `GH_CONTRIBUTE_TOKEN` env var; skips all tests if token is absent
- Points `githubv4.Client` and `ghrest.Client` at real `api.github.com`
- Target: `ivanov-gv/gh-contribute`, PR #1 (stable, locked, known data)
- Calls service methods directly (no binary compilation)
- Asserts `Format()` output matches expected `.md` files in `test/ivanov-gv.gh-contribute.pr#1/`

**EdgeCaseSuite** (no build tag — mock server, runs in CI without token):
- Points clients at `testserver.Server` (local `httptest.Server`)
- Tests only error paths and write operations — **no render/output assertions**
- Covers: HTTP error responses (4xx/5xx), GraphQL error payloads, write operations (POST/mutation) that must not hit real GitHub

### 7.2 Suite structure

**File**: `test/integration/integration_test.go`

```go
//go:build integration

package integration

const (
    realOwner = "ivanov-gv"
    realRepo  = "gh-contribute"
    realPR    = 1
)

// Suite runs against real GitHub API with GH_CONTRIBUTE_TOKEN.
type Suite struct {
    suite.Suite
    prService      *pr.Service
    commentService *comment.Service
    reviewService  *review.Service
    threadService  *thread.Service
}

func (s *Suite) SetupSuite() {
    token := os.Getenv("GH_CONTRIBUTE_TOKEN")
    if token == "" {
        s.T().Skip("GH_CONTRIBUTE_TOKEN not set")
    }
    gql := github.NewGraphQLClient(token) // uses rateLimitTransport
    rest := ghrest.NewClient(nil).WithAuthToken(token)
    s.prService = pr.NewService(gql, realOwner, realRepo)
    s.commentService = comment.NewService(gql, rest, realOwner, realRepo)
    s.reviewService = review.NewService(gql, realOwner, realRepo)
    s.threadService = thread.NewService(gql, realOwner, realRepo)
}

func TestSuite(t *testing.T) { suite.Run(t, new(Suite)) }
```

**File**: `test/integration/edge_cases_test.go` (or split per concern — no build tag)

```go
package integration

// EdgeCaseSuite uses a mock server — no GH_CONTRIBUTE_TOKEN needed.
type EdgeCaseSuite struct {
    suite.Suite
    server         *testserver.Server
    prService      *pr.Service
    commentService *comment.Service
    // ...
}

func (s *EdgeCaseSuite) SetupSuite()  { s.server = testserver.New() }
func (s *EdgeCaseSuite) TearDownSuite() { s.server.Close() }
func (s *EdgeCaseSuite) SetupTest()   { s.server.Reset(); s.rewireServices() }

func TestEdgeCaseSuite(t *testing.T) { suite.Run(t, new(EdgeCaseSuite)) }
```

### 7.3 Real API tests — read services against PR #1

Each test: call the service method, call `Format()`, compare to the expected file.
Helper `readExpected(t, filename)` reads from `../../ivanov-gv.gh-contribute.pr#1/<filename>`.

#### PR service

| Test | Service call | Expected file |
|------|-------------|---------------|
| `TestGetPR` | `prService.Get(1)` → `pr.Format(info)` | `pr-description.md` |

#### Comment service — reads

| Test | Service call | Expected file |
|------|-------------|---------------|
| `TestListComments` | `commentService.List(1)` → `result.Format(false)` | `comments.md` |
| `TestListComments_ShowHidden` | `commentService.List(1)` → `result.Format(true)` | `comments-unhidden.md` |
| `TestGetComment_4038597073` | `List(1).FilterByID(4038597073)` → `Format(false)` | `1-comments-4038597073.md` |
| `TestGetComment_4038597073_ShowHidden` | same → `Format(true)` | `1-comments-4038597073-unhidden.md` |
| `TestGetComment_4038819817` | `List(1).FilterByID(4038819817)` → `Format(false)` | `2-comments-4038819817.md` |
| `TestGetComment_4039142865` | `List(1).FilterByID(4039142865)` → `Format(false)` | `5-comments-4039142865.md` |
| `TestGetComment_4039221478` | `List(1).FilterByID(4039221478)` → `Format(false)` | `6-comments-4039221478.md` |
| `TestGetComment_4039593663` | `List(1).FilterByID(4039593663)` → `Format(false)` | `8-comments-4039593663.md` |
| `TestGetComment_4041153603` | `List(1).FilterByID(4041153603)` → `Format(false)` | `10-comments-4041153603.md` |
| `TestGetComment_4042410800` | `List(1).FilterByID(4042410800)` → `Format(false)` | `11-comments-4042410800.md` |
| `TestGetComment_4067633036` | `List(1).FilterByID(4067633036)` → `Format(false)` | `12-comments-4067633036.md` |

#### Review service

| Test | Service call | Expected file |
|------|-------------|---------------|
| `TestGetReview_3929204495` | `reviewService.Get(1, 3929204495, false)` → `Format()` | `3-review-3929204495.md` |
| `TestGetReview_3929204495_ShowHidden` | same with `showHidden=true` | `3-review-3929204495-unhidden.md` |
| `TestGetReview_3929240428` | `reviewService.Get(1, 3929240428, false)` → `Format()` | `3-3.2.1-review-3929240428.md` |
| `TestGetReview_3929240428_ShowHidden` | same with `showHidden=true` | `3-3.2.1-review-3929240428-unhidden.md` |
| `TestGetReview_3929353771` | `reviewService.Get(1, 3929353771, false)` → `Format()` | `4-review-3929353771.md` |
| `TestGetReview_3929353771_ShowHidden` | same with `showHidden=true` | `4-review-3929353771-unhidden.md` |
| `TestGetReview_3929758963` | `reviewService.Get(1, 3929758963, false)` → `Format()` | `7-review-3929758963.md` |
| `TestGetReview_3930039277` | `reviewService.Get(1, 3930039277, false)` → `Format()` | `9-review-3930039277.md` |
| `TestGetReview_3930039277_ShowHidden` | same with `showHidden=true` | `9-review-3930039277-unhidden.md` |

#### Thread service

| Test | Service call | Expected file |
|------|-------------|---------------|
| `TestGetThread_2918002761` | `threadService.Get(1, 2918002761)` → `Format(false)` | `thread-2918002761.md` |
| `TestGetThread_2918002761_ShowHidden` | same → `Format(true)` | `thread-2918002761-unhidden.md` |
| `TestGetThread_2918006660` | `threadService.Get(1, 2918006660)` → `Format(false)` | `thread-2918006660.md` |
| `TestGetThread_2918006660_ShowHidden` | same → `Format(true)` | `thread-2918006660-unhidden.md` |

### 7.4 Edge case tests (mock server only)

These use `EdgeCaseSuite` with `testserver.Server`. No `Format()` output assertions — only error propagation is verified.

#### HTTP error paths

| Test | Mock response | What it verifies |
|------|--------------|-----------------|
| `TestGraphQL_ServerError` | 500 from `/graphql` | Service returns error, no panic |
| `TestGraphQL_NotFound` | 200 + `{"errors":[{"type":"NOT_FOUND"}]}` | Service wraps GraphQL error |
| `TestGraphQL_EmptyResponse` | 200 + `{"data":null}` | No panic on null data |
| `TestREST_NotFound` | 404 from REST endpoint | Service wraps error with context |
| `TestREST_UnprocessableEntity` | 422 from REST endpoint | Service returns meaningful error |
| `TestREST_ServerError` | 500 from REST endpoint | Service returns error, no panic |

#### Write operations (cannot run against locked PR #1)

| Test | Mock response | What it verifies |
|------|--------------|-----------------|
| `TestPostComment_Success` | REST 201 with comment body | Correct endpoint called, response deserialized |
| `TestReplyToReviewComment_Success` | REST 201 with comment body | Correct `in_reply_to` payload sent |
| `TestPostInlineComment_Success` | REST 201 with comment body | Correct file/line/side/commit_id payload |
| `TestSubmitReview_Success` | REST 200 with review body | Correct event + body sent to `/reviews` |
| `TestAddReaction_IssueComment` | REST 201 | Correct endpoint (`/issues/comments/{id}/reactions`) |
| `TestAddReaction_ReviewComment` | REST 201 | Correct endpoint (`/pulls/comments/{id}/reactions`) |
| `TestAddReaction_InvalidType` | — (no HTTP call) | Service rejects invalid reaction before calling GitHub |
| `TestResolveThread_Success` | GraphQL mutation 200 | `resolveReviewThread` mutation sent with correct node ID |
| `TestUnresolveThread_Success` | GraphQL mutation 200 | `unresolveReviewThread` mutation sent with correct node ID |

### 7.5 GraphQL client transport tests

`internal/client/github/graphql_test.go` — already implemented as unit tests in the same package.
These use `httptest.Server` directly (no `testserver` wrapper needed) and have no build tag.

| Test | What it verifies |
|------|-----------------|
| `TestRateLimitTransport_AuthHeader` | Every request includes `Authorization: Bearer <token>` |
| `TestRateLimitTransport_RetryOn429_EventuallySucceeds` | 429 → retry → eventual 200 |
| `TestRateLimitTransport_RetryOn403` | 403 → retry → eventual 200 |
| `TestRateLimitTransport_MaxRetries_ReturnsLastResponse` | After `maxRetries` attempts returns last response |
| `TestRateLimitTransport_RespectsRetryAfterHeader` | `Retry-After: 0` header used as backoff |

### 7.6 File layout

```
test/
├── e2e_test.go                              # E2E: runs compiled binary (build tag: integration)
├── ivanov-gv.gh-contribute.pr#1/            # shared expected output files
└── integration/
    ├── integration_test.go                  # Suite (real API, build tag: integration)
    ├── pr_test.go                           # Real API: PR #1 read + format
    ├── comment_test.go                      # Real API: comments read + format
    ├── review_test.go                       # Real API: reviews read + format
    ├── thread_test.go                       # Real API: threads read + format
    ├── edge_cases_test.go                   # EdgeCaseSuite scaffolding (mock, no build tag)
    ├── writes_test.go                       # Write operations via mock (no build tag)
    ├── errors_test.go                       # HTTP error paths via mock (no build tag)
    └── testserver/
        └── testserver.go                    # Mock GitHub API server (used only by edge cases)
```

### 7.7 Makefile changes

```makefile
## test-integration: run integration tests against real GitHub API (requires GH_CONTRIBUTE_TOKEN)
test-integration:
	go test -tags integration -count=1 -race ./test/integration/...

## test-integration-local: run edge-case integration tests with mock server (no token needed)
test-integration-local:
	go test -count=1 -race ./test/integration/...

## test-e2e: run E2E tests against real GitHub API (requires GH_CONTRIBUTE_TOKEN)
test-e2e:
	go test -tags integration -count=1 -race ./test/...
```

### 7.8 CI changes

Edge case tests run on every PR (no token needed):

```yaml
- name: Integration edge case tests
  run: make test-integration-local
```

Real API integration tests run with a secret token (same job as E2E):

```yaml
- name: Integration + E2E tests
  env:
    GH_CONTRIBUTE_TOKEN: ${{ secrets.GH_CONTRIBUTE_TOKEN }}
  run: make test-e2e && make test-integration
```

### 7.9 Implementation order

1. **`edge_cases_test.go`** — `EdgeCaseSuite` scaffolding using existing `testserver`. This unblocks write + error tests.
2. **`writes_test.go`** — write operations (POST comment, reply, react, resolve) via mock.
3. **`errors_test.go`** — HTTP error paths via mock.
4. **`integration_test.go`** — `Suite` scaffolding with real GitHub clients.
5. **`pr_test.go`** — simplest real-API test: fetch PR #1, compare to `pr-description.md`.
6. **`comment_test.go`** — real-API comment reads + format comparison.
7. **`review_test.go`** and **`thread_test.go`** — remaining real-API reads.
8. **Makefile + CI** — wire up new targets.

### 7.10 Key design decisions

**Why real GitHub API for main tests?**
Per guidelines, prefer real instances over mocks. A mock server that returns hand-crafted JSON responses validates deserialization in isolation but cannot catch subtle mismatches between the actual GitHub API response shape and what the service expects. Real API calls against the stable, locked PR #1 provide a ground-truth check that the entire HTTP → deserialize → format stack produces the same output as the E2E binary tests.

**Why the same expected files as E2E tests?**
`test/ivanov-gv.gh-contribute.pr#1/*.md` files are the single source of truth for expected output. Using them in both E2E tests (binary stdout) and integration tests (service `Format()` output) guarantees that both layers agree. Any formatting change breaks both, which is intentional.

**Why keep the mock server for edge cases?**
PR #1 is locked — write operations cannot be tested against it without error. HTTP error paths (500, 401, 422) cannot be reproduced against real GitHub. The mock server is the right tool for these scenarios.

**Why `//go:build integration` on `Suite` but not `EdgeCaseSuite`?**
`Suite` requires `GH_CONTRIBUTE_TOKEN`. `EdgeCaseSuite` runs against a local mock with no external dependencies. Edge case tests belong in CI on every PR; real-API tests belong in the same CI job as E2E tests.

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
| **P1** | 7 Integration tests | M | Validates service stack without external deps |
| **P2** | 1.1 Extract shared types | S | Code quality |
| **P2** | 1.6 E2E tests | M | API contract validation |
| **P2** | 2.2-2.3 Inline comments, submit review | M | Advanced review features |
| **P2** | 5.2-5.3 App auth | L | Independent bot identity |
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
test/integration/integration_test.go               — Integration test suite
test/integration/testserver/testserver.go          — Mock GitHub API server
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
Makefile                                           — Add test-integration-local target
.github/workflows/ci.yml                           — Add integration test step
```
