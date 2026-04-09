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

func TestReactCmd_RequiresExactlyTwoArgs(t *testing.T) {
	a := &app{}
	cmd := a.newReactCmd()

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
		cmd.SetArgs([]string{"123", "+1", "extra"})
		err := cmd.Execute()
		assert.Error(t, err)
	})
}

func TestReactCmd_InvalidCommentID(t *testing.T) {
	a := &app{}
	cmd := a.newReactCmd()
	cmd.SetArgs([]string{"not-a-number", "+1"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid comment ID")
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

	// verify command names and Use strings
	commands := map[string]string{
		"pr":        "pr [number]",
		"react":     "react <comment-id> <reaction>",
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

func TestReactCmd_HasTypeFlag(t *testing.T) {
	a := &app{}
	cmd := a.newReactCmd()
	flag := cmd.Flags().Lookup("type")
	require.NotNil(t, flag)
	assert.Equal(t, "review", flag.DefValue)
}

func TestResolveCmd_HasPRFlag(t *testing.T) {
	a := &app{}
	cmd := a.newResolveCmd()
	flag := cmd.Flags().Lookup("pr")
	require.NotNil(t, flag)
	assert.Equal(t, "0", flag.DefValue)
}
