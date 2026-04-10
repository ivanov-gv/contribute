package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolvePR_ExplicitNumber(t *testing.T) {
	// when prNumber > 0, resolvePR returns it directly without hitting any service
	a := &app{}
	number, err := a.resolvePR(42)
	require.NoError(t, err)
	assert.Equal(t, 42, number)
}

func TestReactCommentCmd_RequiresExactlyTwoArgs(t *testing.T) {
	a := &app{}
	cmd := a.newReactCommentCmd()

	t.Run("no args", func(t *testing.T) {
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("one arg", func(t *testing.T) {
		cmd.SetArgs([]string{"123"})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("three args", func(t *testing.T) {
		cmd.SetArgs([]string{"123", "thumbsup", "extra"})
		err := cmd.Execute()
		assert.Error(t, err)
	})
}

func TestReactCommentCmd_InvalidCommentID(t *testing.T) {
	a := &app{}
	cmd := a.newReactCommentCmd()
	cmd.SetArgs([]string{"not-a-number", "thumbsup"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid comment ID")
}

func TestReactReviewCmd_RequiresExactlyTwoArgs(t *testing.T) {
	a := &app{}
	cmd := a.newReactReviewCmd()

	t.Run("no args", func(t *testing.T) {
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		assert.Error(t, err)
	})

	t.Run("one arg", func(t *testing.T) {
		cmd.SetArgs([]string{"456"})
		err := cmd.Execute()
		assert.Error(t, err)
	})
}

func TestReactReviewCmd_InvalidReviewID(t *testing.T) {
	a := &app{}
	cmd := a.newReactReviewCmd()
	cmd.SetArgs([]string{"not-a-number", "eyes"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid review ID")
}

func TestReactReviewCmd_HasPRFlag(t *testing.T) {
	a := &app{}
	cmd := a.newReactReviewCmd()
	flag := cmd.Flags().Lookup("pr")
	require.NotNil(t, flag)
	assert.Equal(t, "0", flag.DefValue)
}

func TestReactIssueCommentCmd_InvalidCommentID(t *testing.T) {
	a := &app{}
	cmd := a.newReactIssueCommentCmd()
	cmd.SetArgs([]string{"not-a-number", "rocket"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid comment ID")
}

func TestReactCmd_HasSubcommands(t *testing.T) {
	a := &app{}
	cmd := a.newReactCmd()

	subNames := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subNames[sub.Name()] = true
	}

	assert.True(t, subNames["comment"], "react should have 'comment' subcommand")
	assert.True(t, subNames["review"], "react should have 'review' subcommand")
	assert.True(t, subNames["issue-comment"], "react should have 'issue-comment' subcommand")
}

func TestResolveCmd_RequiresExactlyOneArg(t *testing.T) {
	a := &app{}
	cmd := a.newResolveCmd()

	t.Run("no args", func(t *testing.T) {
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		assert.Error(t, err)
	})
}

func TestResolveCmd_InvalidThreadID(t *testing.T) {
	a := &app{}
	cmd := a.newResolveCmd()
	cmd.SetArgs([]string{"not-a-number"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid thread ID")
}

func TestPRCmd_InvalidPRNumber(t *testing.T) {
	a := &app{}
	cmd := a.newPRCmd()
	cmd.SetArgs([]string{"not-a-number"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PR number")
}

func TestCommandWiring(t *testing.T) {
	a := &app{}

	commands := map[string]string{
		"pr":        "pr [number]",
		"react":     "react",
		"resolve":   "resolve <thread-id>",
		"unresolve": "unresolve <thread-id>",
	}

	for name, expectedUse := range commands {
		t.Run(name, func(t *testing.T) {
			switch name {
			case "pr":
				c := a.newPRCmd()
				assert.Equal(t, expectedUse, c.Use)
			case "react":
				c := a.newReactCmd()
				assert.Equal(t, expectedUse, c.Use)
			case "resolve":
				c := a.newResolveCmd()
				assert.Equal(t, expectedUse, c.Use)
			case "unresolve":
				c := a.newUnresolveCmd()
				assert.Equal(t, expectedUse, c.Use)
			}
		})
	}
}

func TestGitCredentialsCmd_IsHidden(t *testing.T) {
	cmd := newGitCredentialsCmd()
	assert.True(t, cmd.Hidden)
}

func TestGitCredentialsCmd_NonGetOperationsAreNoop(t *testing.T) {
	for _, op := range []string{"store", "erase"} {
		t.Run(op, func(t *testing.T) {
			cmd := newGitCredentialsCmd()
			cmd.SetArgs([]string{op})
			err := cmd.Execute()
			assert.NoError(t, err)
		})
	}
}

func TestGitCredentialsCmd_NoArgsIsNoop(t *testing.T) {
	cmd := newGitCredentialsCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestTokenCmd_RequiresAuth(t *testing.T) {
	// clear all auth env vars and point HOME to empty temp dir
	t.Setenv("GH_CONTRIBUTE_TOKEN", "")
	t.Setenv("GH_CONTRIBUTE_APP_ID", "")
	t.Setenv("GH_CONTRIBUTE_PRIVATE_KEY", "")
	t.Setenv("GH_CONTRIBUTE_PRIVATE_KEY_PATH", "")
	t.Setenv("HOME", t.TempDir())

	cmd := newTokenCmd()
	err := cmd.Execute()
	assert.Error(t, err)
}

func TestResolveCmd_HasPRFlag(t *testing.T) {
	a := &app{}
	cmd := a.newResolveCmd()
	flag := cmd.Flags().Lookup("pr")
	require.NotNil(t, flag)
	assert.Equal(t, "0", flag.DefValue)
}
