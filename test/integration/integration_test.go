//go:build integration

// Package integration contains integration tests for gh-contribute services.
//
// Suite (this file) runs against the real GitHub API (GH_CONTRIBUTE_TOKEN required).
// It targets the stable, locked PR #1 in ivanov-gv/gh-contribute and compares
// service Format() output against the shared expected files in
// test/ivanov-gv.gh-contribute.pr#1/ — the same files used by the E2E binary tests.
//
// EdgeCaseSuite (edge_cases_test.go) uses a local mock server and requires no token.
//
// Run real API tests:  go test -tags integration -count=1 -race ./test/integration/...
// Run edge case tests: go test -count=1 -race ./test/integration/...
package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	ghrest "github.com/google/go-github/v69/github"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	githubclient "github.com/ivanov-gv/gh-contribute/internal/client/github"
	"github.com/ivanov-gv/gh-contribute/internal/service/comment"
	"github.com/ivanov-gv/gh-contribute/internal/service/pr"
	"github.com/ivanov-gv/gh-contribute/internal/service/review"
	"github.com/ivanov-gv/gh-contribute/internal/service/thread"
)

const (
	realOwner   = "ivanov-gv"
	realRepo    = "gh-contribute"
	realPR      = 1
	testDataDir = "../ivanov-gv.gh-contribute.pr#1"
)

// Suite runs against the real GitHub API using GH_CONTRIBUTE_TOKEN.
// Target: ivanov-gv/gh-contribute PR #1 (stable, locked, known expected output).
// All test methods compare service Format() output to expected .md files.
type Suite struct {
	suite.Suite
	prService      *pr.Service
	commentService *comment.Service
	reviewService  *review.Service
	threadService  *thread.Service
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupSuite() {
	token := os.Getenv("GH_CONTRIBUTE_TOKEN")
	if token == "" {
		s.T().Skip("GH_CONTRIBUTE_TOKEN not set — skipping real API integration tests")
	}
	gql := githubclient.NewGraphQLClient(token)
	rest := ghrest.NewClient(nil).WithAuthToken(token)
	s.prService = pr.NewService(gql, realOwner, realRepo)
	s.commentService = comment.NewService(gql, rest, realOwner, realRepo)
	s.reviewService = review.NewService(gql, realOwner, realRepo)
	s.threadService = thread.NewService(gql, realOwner, realRepo)
}

// readExpected reads and normalizes an expected output file from the shared test data directory.
// Normalization matches E2E test behavior: strip trailing newlines, add exactly one.
func readExpected(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testDataDir, filename))
	require.NoError(t, err, "read expected file: %s", filename)
	return strings.TrimRight(string(data), "\n") + "\n"
}

// normalize strips extra trailing newlines so actual output matches readExpected normalization.
func normalize(s string) string {
	return strings.TrimRight(s, "\n") + "\n"
}
