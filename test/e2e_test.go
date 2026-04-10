//go:build integration

package test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDataDir = "ivanov-gv.contribute.pr#1"
	prNumber    = "1"
)

// testCase maps a CLI command to its expected output file
type testCase struct {
	name         string
	args         []string
	expectedFile string
}

func testCases() []testCase {
	return []testCase{
		// pr description
		{"pr", []string{"pr", prNumber}, "pr-description.md"},

		// comments list
		{"comments", []string{"comments", "--pr", prNumber}, "comments.md"},
		{"comments --show-hidden", []string{"comments", "--pr", prNumber, "--show-hidden"}, "comments-unhidden.md"},

		// single comments
		{"comments 4038597073", []string{"comments", "4038597073", "--pr", prNumber}, "1-comments-4038597073.md"},
		{"comments 4038597073 --show-hidden", []string{"comments", "4038597073", "--pr", prNumber, "--show-hidden"}, "1-comments-4038597073-unhidden.md"},
		{"comments 4038819817", []string{"comments", "4038819817", "--pr", prNumber}, "2-comments-4038819817.md"},
		{"comments 4039142865", []string{"comments", "4039142865", "--pr", prNumber}, "5-comments-4039142865.md"},
		{"comments 4039221478", []string{"comments", "4039221478", "--pr", prNumber}, "6-comments-4039221478.md"},
		{"comments 4039593663", []string{"comments", "4039593663", "--pr", prNumber}, "8-comments-4039593663.md"},
		{"comments 4041153603", []string{"comments", "4041153603", "--pr", prNumber}, "10-comments-4041153603.md"},
		{"comments 4042410800", []string{"comments", "4042410800", "--pr", prNumber}, "11-comments-4042410800.md"},
		{"comments 4067633036", []string{"comments", "4067633036", "--pr", prNumber}, "12-comments-4067633036.md"},

		// reviews
		{"review 3929204495", []string{"review", "3929204495", "--pr", prNumber}, "3-review-3929204495.md"},
		{"review 3929204495 --show-hidden", []string{"review", "3929204495", "--pr", prNumber, "--show-hidden"}, "3-review-3929204495-unhidden.md"},
		{"review 3929240428", []string{"review", "3929240428", "--pr", prNumber}, "3-3.2.1-review-3929240428.md"},
		{"review 3929240428 --show-hidden", []string{"review", "3929240428", "--pr", prNumber, "--show-hidden"}, "3-3.2.1-review-3929240428-unhidden.md"},
		{"review 3929353771", []string{"review", "3929353771", "--pr", prNumber}, "4-review-3929353771.md"},
		{"review 3929353771 --show-hidden", []string{"review", "3929353771", "--pr", prNumber, "--show-hidden"}, "4-review-3929353771-unhidden.md"},
		{"review 3929758963", []string{"review", "3929758963", "--pr", prNumber}, "7-review-3929758963.md"},
		{"review 3930039277", []string{"review", "3930039277", "--pr", prNumber}, "9-review-3930039277.md"},
		{"review 3930039277 --show-hidden", []string{"review", "3930039277", "--pr", prNumber, "--show-hidden"}, "9-review-3930039277-unhidden.md"},

		// threads
		{"thread 2918002761", []string{"thread", "2918002761", "--pr", prNumber}, "thread-2918002761.md"},
		{"thread 2918002761 --show-hidden", []string{"thread", "2918002761", "--pr", prNumber, "--show-hidden"}, "thread-2918002761-unhidden.md"},
		{"thread 2918006660", []string{"thread", "2918006660", "--pr", prNumber}, "thread-2918006660.md"},
		{"thread 2918006660 --show-hidden", []string{"thread", "2918006660", "--pr", prNumber, "--show-hidden"}, "thread-2918006660-unhidden.md"},
	}
}

func TestE2E(t *testing.T) {
	// require GH_CONTRIBUTE_TOKEN
	token := os.Getenv("GH_CONTRIBUTE_TOKEN")
	if token == "" {
		t.Skip("GH_CONTRIBUTE_TOKEN not set — skipping e2e tests")
	}

	// build binary
	binaryPath := filepath.Join(t.TempDir(), "contribute")
	buildCmd := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, "./cmd/contribute") //nolint:gosec // test builds known binary
	buildCmd.Dir = ".."
	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(out))

	for _, tc := range testCases() {
		t.Run(tc.name, func(t *testing.T) {
			// read expected output
			expectedBytes, err := os.ReadFile(filepath.Join(testDataDir, tc.expectedFile))
			require.NoError(t, err, "read expected file %s", tc.expectedFile)
			expected := string(expectedBytes)

			// run binary
			cmd := exec.CommandContext(context.Background(), binaryPath, tc.args...) //nolint:gosec // test runs known binary with test args
			cmd.Dir = ".."
			cmd.Env = append(os.Environ(), "GH_CONTRIBUTE_TOKEN="+token)
			stdout, err := cmd.Output()
			require.NoError(t, err, "command failed: %s\nstderr: %s", tc.name, getStderr(cmd, err))

			actual := string(stdout)

			// normalize trailing whitespace for comparison
			expected = strings.TrimRight(expected, "\n") + "\n"
			actual = strings.TrimRight(actual, "\n") + "\n"

			assert.Equal(t, expected, actual, "output mismatch for %s", tc.name)
		})
	}
}

// getStderr extracts stderr from an exec.ExitError if available
func getStderr(_ *exec.Cmd, err error) string {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(exitErr.Stderr)
	}
	return ""
}
