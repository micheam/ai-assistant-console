package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
)

// setupSessionTestEnv points AI_ASSISTANT_CONFIG_PATH at a nonexistent file
// (so config.Load falls back to config.DefaultConfig, as production code
// does) and XDG_DATA_HOME at a fresh temp dir, then returns the session
// directory config.DefaultSessionDir will resolve to.
func setupSessionTestEnv(t *testing.T) (sessionDir string) {
	t.Helper()
	t.Setenv(config.EnvKeyConfigPath, filepath.Join(t.TempDir(), "no-such-config.toml"))
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	return filepath.Join(dataHome, config.ApplicationFQN, "sessions")
}

// writeSessionFixture marshals sess via Session.MarshalJSON, writes it to
// <dir>/<sess.ID>.json, and sets its mtime.
func writeSessionFixture(t *testing.T, dir string, sess *assistant.Session, modTime time.Time) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	data, err := sess.MarshalJSON()
	require.NoError(t, err)
	path := filepath.Join(dir, sess.ID+".json")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	require.NoError(t, os.Chtimes(path, modTime, modTime))
}

// runSessionApp runs CmdSession with the global --json flag registered on
// the root, capturing stdout into buf. A fresh *cli.Command is built per
// call, since urfave/cli v3 resets flagJSON's value at the start of every
// Run.
func runSessionApp(t *testing.T, buf *bytes.Buffer, args ...string) error {
	t.Helper()
	app := &cli.Command{
		Name:     "aico",
		Writer:   buf,
		Flags:    []cli.Flag{flagJSON},
		Commands: []*cli.Command{CmdSession},
	}
	return app.Run(context.Background(), append([]string{"aico"}, args...))
}

func TestSessionList_JSON_NoSessionDir(t *testing.T) {
	setupSessionTestEnv(t) // session dir intentionally left uncreated

	var buf bytes.Buffer
	err := runSessionApp(t, &buf, "--json", "session", "list")
	require.NoError(t, err)
	require.Equal(t, "[]\n", buf.String())
}

func TestSessionList_JSON_Empty(t *testing.T) {
	dir := setupSessionTestEnv(t)
	require.NoError(t, os.MkdirAll(dir, 0o755))

	var buf bytes.Buffer
	err := runSessionApp(t, &buf, "--json", "session", "list")
	require.NoError(t, err)
	require.Equal(t, "[]\n", buf.String())
}

func TestSessionList_JSON_WithLimit(t *testing.T) {
	dir := setupSessionTestEnv(t)
	now := time.Now()

	older := &assistant.Session{ID: "older", Model: "m1",
		Messages: []assistant.Message{assistant.NewUserMessage(assistant.NewTextContent("hello"))}}
	writeSessionFixture(t, dir, older, now.Add(-time.Hour))

	newer := &assistant.Session{ID: "newer", Model: "m2",
		Messages: []assistant.Message{
			assistant.NewUserMessage(assistant.NewTextContent("hi\nthere")),
			assistant.NewAssistantMessage(assistant.NewTextContent("hello!")),
		}}
	writeSessionFixture(t, dir, newer, now)

	var buf bytes.Buffer
	// Post-positioned --json, to exercise the persistent-flag path too.
	err := runSessionApp(t, &buf, "session", "list", "--json", "-n", "1")
	require.NoError(t, err)

	var items []sessionListItemView
	require.NoError(t, json.Unmarshal(buf.Bytes(), &items))
	require.Len(t, items, 1)
	require.Equal(t, "newer", items[0].ID)
	require.Equal(t, 2, items[0].MessageCount)
	require.Equal(t, "hi there", items[0].Preview)
	require.Equal(t, now.Format(time.RFC3339), items[0].UpdatedAt)
}

func TestSessionList_Table_NoSessionDir(t *testing.T) {
	setupSessionTestEnv(t)

	var buf bytes.Buffer
	err := runSessionApp(t, &buf, "session", "list")
	require.NoError(t, err)
	require.Equal(t, "No sessions found.\n", buf.String())
}

func TestSessionShow_JSON(t *testing.T) {
	dir := setupSessionTestEnv(t)
	sess := &assistant.Session{
		ID:    "show-json",
		Model: "m1",
		Messages: []assistant.Message{
			assistant.NewUserMessage(assistant.NewTextContent("Hello, how are you?")),
			assistant.NewAssistantMessage(assistant.NewTextContent("I'm fine, thank you!")),
		},
	}
	writeSessionFixture(t, dir, sess, time.Now())

	var buf bytes.Buffer
	err := runSessionApp(t, &buf, "--json", "session", "show", "show-json")
	require.NoError(t, err)

	// sessionShowView.Messages is []assistant.Message (an interface), which
	// encoding/json cannot unmarshal directly, so decode into a generic
	// shape instead.
	var view struct {
		ID        string `json:"id"`
		Model     string `json:"model"`
		UpdatedAt string `json:"updated_at"`
		Messages  []struct {
			Author string `json:"author"`
		} `json:"messages"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &view))
	require.Equal(t, "show-json", view.ID)
	require.Equal(t, "m1", view.Model)
	require.NotEmpty(t, view.UpdatedAt)
	require.Len(t, view.Messages, 2)
	require.Equal(t, string(assistant.MessageAuthorUser), view.Messages[0].Author)
	require.Equal(t, string(assistant.MessageAuthorAssistant), view.Messages[1].Author)
}

func TestSessionShow_Text(t *testing.T) {
	dir := setupSessionTestEnv(t)
	sess := &assistant.Session{
		ID:    "show-text",
		Model: "m1",
		Messages: []assistant.Message{
			assistant.NewUserMessage(assistant.NewTextContent("Hello, how are you?")),
			assistant.NewAssistantMessage(assistant.NewTextContent("I'm fine, thank you!")),
		},
	}
	writeSessionFixture(t, dir, sess, time.Now())

	var buf bytes.Buffer
	err := runSessionApp(t, &buf, "session", "show", "show-text")
	require.NoError(t, err)
	require.Equal(t,
		"--- user ---\nHello, how are you?\n\n--- assistant ---\nI'm fine, thank you!\n",
		buf.String())
}

// TestContextFlag_CommaNotSplit_ThroughSessionResume exercises the actual
// production CmdSession command tree (root -> session -> resume) to confirm
// --context values aren't split on ',' at any level of subcommand dispatch,
// since urfave/cli v3 resets its (process-wide) slice separator setting
// from each command's own DisableSliceFlagSeparator field as it descends
// into subcommands. It stops short of invoking runSessionResume's real
// generation logic by intercepting via the session ID: the id doesn't
// resolve to a session file, but this test only cares that the flag
// parsing runs and StringSlice("context") comes back unsplit before that
// failure is reached.
func TestContextFlag_CommaNotSplit_ThroughSessionResume(t *testing.T) {
	setupSessionTestEnv(t)

	app := &cli.Command{
		Name:                      "aico",
		DisableSliceFlagSeparator: true,
		Flags:                     []cli.Flag{flagJSON, flagContext},
		Commands:                  []*cli.Command{CmdSession},
	}
	err := app.Run(context.Background(), []string{
		"aico", "session", "resume", "no-such-session", "prompt",
		"--context", "go doc output, with a comma",
	})
	// The session doesn't exist, so generation fails downstream; that's
	// expected and not what this test checks.
	require.Error(t, err)

	got := app.StringSlice(flagContext.Name)
	require.Equal(t, []string{"go doc output, with a comma"}, got)
}

// TestCmdSession_DisableSliceFlagSeparator guards the Step-0 fix: it is a
// regression test for the fact that urfave/cli v3 requires
// DisableSliceFlagSeparator to be set on every command in the dispatch
// chain that accepts --context, not just on the root command.
func TestCmdSession_DisableSliceFlagSeparator(t *testing.T) {
	require.True(t, CmdSession.DisableSliceFlagSeparator, "CmdSession must disable the slice flag separator")

	var resume *cli.Command
	for _, sub := range CmdSession.Commands {
		if sub.Name == "resume" {
			resume = sub
			break
		}
	}
	require.NotNil(t, resume, "expected a \"resume\" subcommand")
	require.True(t, resume.DisableSliceFlagSeparator, "\"resume\" must disable the slice flag separator")
}

func TestSessionShow_NoArgs(t *testing.T) {
	setupSessionTestEnv(t)

	var buf bytes.Buffer
	err := runSessionApp(t, &buf, "session", "show")
	require.Error(t, err)
}

func TestSessionShow_UnknownID(t *testing.T) {
	setupSessionTestEnv(t)

	var buf bytes.Buffer
	err := runSessionApp(t, &buf, "session", "show", "no-such-id")
	require.Error(t, err)
}
