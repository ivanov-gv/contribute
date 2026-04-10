//go:build !integration

// Package integration — EdgeCaseSuite tests error paths and write operations
// using a local mock HTTP server. Runs in CI without GH_CONTRIBUTE_TOKEN.
//
// Run: go test -count=1 -race ./test/integration/...
package integration

import (
	"net/http"
	"net/url"
	"testing"

	ghrest "github.com/google/go-github/v69/github"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/suite"

	"github.com/ivanov-gv/contribute/internal/service/comment"
	"github.com/ivanov-gv/contribute/internal/service/issue"
	"github.com/ivanov-gv/contribute/internal/service/pr"
	"github.com/ivanov-gv/contribute/internal/service/reaction"
	"github.com/ivanov-gv/contribute/internal/service/review"
	"github.com/ivanov-gv/contribute/internal/service/thread"
	"github.com/ivanov-gv/contribute/test/integration/testserver"
)

const (
	testOwner = "testowner"
	testRepo  = "testrepo"
	testToken = "test-token"
)

// Query patterns used to match GraphQL requests in the mock server.
const (
	prQueryPattern         = "closingIssuesReferences"
	commentsQueryPattern   = "reactions(first: 100)"
	allReviewsQueryPattern = "reviews(first: 100){nodes{databaseId"
	threadsQueryPattern    = "nodes{id,isOutdated"
	issueGetPattern        = "issue(number: $number)"
	issueListPattern       = "states: OPEN, orderBy"
)

// EdgeCaseSuite tests error paths and write operations using a mock HTTP server.
type EdgeCaseSuite struct {
	suite.Suite
	server          *testserver.Server
	prService       *pr.Service
	commentService  *comment.Service
	reactionService *reaction.Service
	reviewService   *review.Service
	threadService   *thread.Service
	issueService    *issue.Service
}

func TestEdgeCaseSuite(t *testing.T) {
	suite.Run(t, new(EdgeCaseSuite))
}

func (s *EdgeCaseSuite) SetupSuite() {
	s.server = testserver.New()
}

func (s *EdgeCaseSuite) TearDownSuite() {
	s.server.Close()
}

func (s *EdgeCaseSuite) SetupTest() {
	s.server.Reset()
	gql := githubv4.NewEnterpriseClient(s.server.GraphQLURL(), &http.Client{})
	rest := newRESTClient(s.server.URL)
	s.prService = pr.NewService(gql, testOwner, testRepo)
	s.commentService = comment.NewService(gql, rest, testOwner, testRepo)
	s.reactionService = reaction.NewService(rest, gql, testOwner, testRepo)
	s.reviewService = review.NewService(gql, testOwner, testRepo)
	s.threadService = thread.NewService(gql, testOwner, testRepo)
	s.issueService = issue.NewService(gql, testOwner, testRepo)
}

// newRESTClient creates a go-github client pointed at the mock server.
func newRESTClient(serverURL string) *ghrest.Client {
	base, _ := url.Parse(serverURL + "/") //nolint:errcheck // test helper with well-formed URL from httptest.Server
	c := ghrest.NewClient(&http.Client{}).WithAuthToken(testToken)
	c.BaseURL = base
	return c
}

// ── Mock data builders ────────────────────────────────────────────────────────
// These build minimal but structurally correct GraphQL/REST responses for the mock server.

func prData(number int, title, state string) map[string]interface{} {
	return map[string]interface{}{
		"repository": map[string]interface{}{
			"pullRequest": map[string]interface{}{
				"number":                  number,
				"title":                   title,
				"state":                   state,
				"isDraft":                 false,
				"mergeable":               "MERGEABLE",
				"body":                    "PR body",
				"url":                     "https://github.com/testowner/testrepo/pull/42",
				"headRefName":             "feature-branch",
				"headRefOid":              "abc123def456",
				"baseRefName":             "main",
				"locked":                  false,
				"changedFiles":            3,
				"additions":               10,
				"deletions":               5,
				"totalCommentsCount":      2,
				"author":                  map[string]interface{}{"login": "testuser"},
				"commits":                 map[string]interface{}{"totalCount": 1},
				"comments":                map[string]interface{}{"totalCount": 1},
				"reviews":                 map[string]interface{}{"totalCount": 1},
				"assignees":               map[string]interface{}{"nodes": []interface{}{}},
				"labels":                  map[string]interface{}{"nodes": []interface{}{}},
				"reviewRequests":          map[string]interface{}{"nodes": []interface{}{}},
				"milestone":               nil,
				"projectsV2":              map[string]interface{}{"nodes": []interface{}{}},
				"closingIssuesReferences": map[string]interface{}{"nodes": []interface{}{}},
			},
		},
	}
}

func commentsQueryData(issueComments, reviews []interface{}, threads []interface{}) map[string]interface{} {
	if issueComments == nil {
		issueComments = []interface{}{}
	}
	if reviews == nil {
		reviews = []interface{}{}
	}
	if threads == nil {
		threads = []interface{}{}
	}
	return map[string]interface{}{
		"viewer": map[string]interface{}{"login": "testviewer"},
		"repository": map[string]interface{}{
			"pullRequest": map[string]interface{}{
				"comments":      map[string]interface{}{"nodes": issueComments},
				"reviews":       map[string]interface{}{"nodes": reviews},
				"reviewThreads": map[string]interface{}{"nodes": threads},
			},
		},
	}
}

func issueCommentNode(id int, author, body string, minimized bool) map[string]interface{} {
	return map[string]interface{}{
		"databaseId":      id,
		"author":          map[string]interface{}{"login": author},
		"body":            body,
		"createdAt":       "2024-01-15T10:00:00Z",
		"isMinimized":     minimized,
		"minimizedReason": "",
		"reactions":       map[string]interface{}{"nodes": []interface{}{}},
	}
}

func reviewNode(id int, author, body, state string, commentCount int) map[string]interface{} {
	return map[string]interface{}{
		"databaseId":      id,
		"author":          map[string]interface{}{"login": author},
		"body":            body,
		"state":           state,
		"createdAt":       "2024-01-15T11:00:00Z",
		"isMinimized":     false,
		"minimizedReason": "",
		"comments":        map[string]interface{}{"totalCount": commentCount},
		"reactions":       map[string]interface{}{"nodes": []interface{}{}},
	}
}

func allReviewsData(reviewID int64, threads []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"viewer": map[string]interface{}{"login": "testviewer"},
		"repository": map[string]interface{}{
			"pullRequest": map[string]interface{}{
				"reviews": map[string]interface{}{
					"nodes": []interface{}{
						map[string]interface{}{
							"databaseId":      reviewID,
							"author":          map[string]interface{}{"login": "reviewer1"},
							"body":            "Review body",
							"state":           "CHANGES_REQUESTED",
							"createdAt":       "2024-01-15T10:00:00Z",
							"isMinimized":     false,
							"minimizedReason": "",
							"reactions":       map[string]interface{}{"nodes": []interface{}{}},
						},
					},
				},
				"reviewThreads": map[string]interface{}{"nodes": threads},
			},
		},
	}
}

func buildReviewThreadNode(reviewID int64, resolved bool, commentIDs []int64) map[string]interface{} {
	comments := make([]interface{}, len(commentIDs))
	for i, id := range commentIDs {
		comments[i] = map[string]interface{}{
			"databaseId":      id,
			"author":          map[string]interface{}{"login": "reviewer1"},
			"body":            "Thread comment",
			"createdAt":       "2024-01-15T10:30:00Z",
			"isMinimized":     false,
			"minimizedReason": "",
			"replyTo":         nil,
			"pullRequestReview": map[string]interface{}{
				"databaseId": reviewID,
			},
			"reactions": map[string]interface{}{"nodes": []interface{}{}},
		}
	}
	return map[string]interface{}{
		"isOutdated":        false,
		"isResolved":        resolved,
		"path":              "main.go",
		"line":              42,
		"startLine":         nil,
		"originalLine":      42,
		"originalStartLine": nil,
		"comments":          map[string]interface{}{"nodes": comments},
	}
}

func threadsQueryData(viewerLogin string, threads []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"viewer": map[string]interface{}{"login": viewerLogin},
		"repository": map[string]interface{}{
			"pullRequest": map[string]interface{}{
				"reviewThreads": map[string]interface{}{"nodes": threads},
			},
		},
	}
}

func threadNode(nodeID string, resolved bool, comments []interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":                nodeID,
		"isOutdated":        false,
		"isResolved":        resolved,
		"path":              "pkg/service.go",
		"line":              15,
		"startLine":         nil,
		"originalLine":      15,
		"originalStartLine": nil,
		"comments":          map[string]interface{}{"nodes": comments},
	}
}

func threadCommentNode(id int64, reviewID int64, replyToID *int64, author, body string) map[string]interface{} {
	node := map[string]interface{}{
		"databaseId":      id,
		"author":          map[string]interface{}{"login": author},
		"body":            body,
		"createdAt":       "2024-01-15T10:00:00Z",
		"isMinimized":     false,
		"minimizedReason": "",
		"replyTo":         nil,
		"pullRequestReview": map[string]interface{}{
			"databaseId": reviewID,
		},
		"reactions": map[string]interface{}{"nodes": []interface{}{}},
	}
	if replyToID != nil {
		node["replyTo"] = map[string]interface{}{"databaseId": *replyToID}
	}
	return node
}

func issueData(number int, title, state string) map[string]interface{} {
	return map[string]interface{}{
		"repository": map[string]interface{}{
			"issue": map[string]interface{}{
				"number":    number,
				"title":     title,
				"state":     state,
				"body":      "Issue body",
				"url":       "https://github.com/testowner/testrepo/issues/10",
				"author":    map[string]interface{}{"login": "issueauthor"},
				"labels":    map[string]interface{}{"nodes": []interface{}{}},
				"assignees": map[string]interface{}{"nodes": []interface{}{}},
				"comments": map[string]interface{}{
					"totalCount": 0,
					"nodes":      []interface{}{},
				},
				"timelineItems": map[string]interface{}{"nodes": []interface{}{}},
			},
		},
	}
}

// reactionResponse is the go-github REST response for a created reaction.
var reactionResponse = map[string]interface{}{
	"id":      1,
	"content": "+1",
	"user":    map[string]interface{}{"login": "testuser"},
}
