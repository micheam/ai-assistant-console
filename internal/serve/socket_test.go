//go:build unix

package serve

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/config"
)

func TestDefaultSocketPath(t *testing.T) {
	t.Run("XDG_RUNTIME_DIR first", func(t *testing.T) {
		t.Setenv("XDG_RUNTIME_DIR", "/run/user/501")
		t.Setenv("TMPDIR", "/var/tmp/x")
		assert.Equal(t, "/run/user/501/"+config.ApplicationFQN+"/serve.sock", DefaultSocketPath())
	})
	t.Run("then TMPDIR", func(t *testing.T) {
		t.Setenv("XDG_RUNTIME_DIR", "")
		t.Setenv("TMPDIR", "/var/tmp/x")
		assert.Equal(t, "/var/tmp/x/"+config.ApplicationFQN+"/serve.sock", DefaultSocketPath())
	})
	t.Run("then a per-user directory in the temp dir", func(t *testing.T) {
		t.Setenv("XDG_RUNTIME_DIR", "")
		t.Setenv("TMPDIR", "")
		got := DefaultSocketPath()
		assert.Contains(t, got, config.ApplicationFQN+"-")
		assert.Equal(t, "serve.sock", filepath.Base(got))
	})
}

func TestListen(t *testing.T) {
	// Socket paths are length-limited, so use a short temp dir.
	dir, err := os.MkdirTemp("", "aico")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	path := filepath.Join(dir, "sub", "serve.sock")

	t.Run("creates the directory with 0700, listens and holds the lock", func(t *testing.T) {
		ln, err := Listen(path)
		require.NoError(t, err)
		defer ln.Close()

		info, err := os.Stat(filepath.Dir(path))
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o700), info.Mode().Perm())

		conn, err := net.Dial("unix", path)
		require.NoError(t, err)
		conn.Close()

		_, err = Listen(path)
		assert.ErrorIs(t, err, ErrAlreadyRunning)
	})

	t.Run("Close removes the socket and releases the lock", func(t *testing.T) {
		ln, err := Listen(path)
		require.NoError(t, err)
		require.NoError(t, ln.Close())

		_, err = os.Stat(path)
		assert.ErrorIs(t, err, os.ErrNotExist)

		ln2, err := Listen(path)
		require.NoError(t, err)
		ln2.Close()
	})

	t.Run("replaces a stale socket file", func(t *testing.T) {
		stale, err := net.Listen("unix", path)
		require.NoError(t, err)
		// Drop the listener without unlinking, as a crashed daemon would.
		stale.(*net.UnixListener).SetUnlinkOnClose(false)
		stale.Close()
		_, err = os.Stat(path)
		require.NoError(t, err, "stale socket file should remain")

		ln, err := Listen(path)
		require.NoError(t, err)
		defer ln.Close()
		conn, err := net.Dial("unix", path)
		require.NoError(t, err)
		conn.Close()
	})
}
