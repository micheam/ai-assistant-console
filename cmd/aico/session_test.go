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
