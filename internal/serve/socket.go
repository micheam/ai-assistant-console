//go:build unix

package serve

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"syscall"

	"micheam.com/aico/internal/config"
)

const (
	socketFileName = "serve.sock"
	lockFileName   = "serve.lock"
)

// ErrAlreadyRunning is returned by Listen when another daemon holds the
// lock for the socket's directory.
var ErrAlreadyRunning = errors.New("another aico serve is already running")

// DefaultSocketPath returns where the daemon listens unless told
// otherwise: under $XDG_RUNTIME_DIR, else $TMPDIR (macOS has a per-user
// one), else a per-user directory in /tmp.
func DefaultSocketPath() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, config.ApplicationFQN, socketFileName)
	}
	if dir := os.Getenv("TMPDIR"); dir != "" {
		return filepath.Join(dir, config.ApplicationFQN, socketFileName)
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s-%d", config.ApplicationFQN, os.Getuid()), socketFileName)
}

// Listener is a Unix socket listener that also owns the directory lock.
// Close releases the lock and removes the socket file.
type Listener struct {
	net.Listener
	path string
	lock *os.File
}

// Listen takes the exclusive lock next to path, removes any stale socket
// file, and listens on path. The socket's directory is created with mode
// 0700 if needed.
//
// errors:
//
//   - [ErrAlreadyRunning]: another process holds the lock.
func Listen(path string) (*Listener, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("create socket directory: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return nil, fmt.Errorf("restrict socket directory: %w", err)
	}

	lock, err := os.OpenFile(filepath.Join(dir, lockFileName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		lock.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrAlreadyRunning
		}
		return nil, fmt.Errorf("lock %s: %w", lock.Name(), err)
	}

	// Only the lock holder creates the socket, so a leftover file can only
	// be stale.
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		lock.Close()
		return nil, fmt.Errorf("remove stale socket: %w", err)
	}
	ln, err := net.Listen("unix", path)
	if err != nil {
		lock.Close()
		return nil, fmt.Errorf("listen: %w", err)
	}
	return &Listener{Listener: ln, path: path, lock: lock}, nil
}

// Close closes the listener, removes the socket file and releases the lock.
func (l *Listener) Close() error {
	err := l.Listener.Close()
	os.Remove(l.path)
	l.lock.Close() // closing the descriptor releases the flock
	return err
}
