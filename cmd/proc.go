package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jimtang2/nux/lib/collector"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const (
	COL_PID_FILE = ".nux/col_pid"
)

func pidFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".nux")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "col_pid"), nil
}

func writePidFile(pid int) error {
	path, err := pidFilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0o600)
}

func readPidFile() (int, error) {
	path, err := pidFilePath()
	if err != nil {
		return 0, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func removePidFile() error {
	path, err := pidFilePath()
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func processExists(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds; check if it's actually alive.
	err = p.Signal(unix.Signal(0))
	return err == nil
}

// startColProc starts the collector as a background daemon.
// It re-executes the current binary in daemon mode.
func startColProc(cmd *cobra.Command) (int, error) {
	// Check if already running
	pid, err := readPidFile()
	if err == nil && processExists(pid) {
		return pid, fmt.Errorf("collector already running (PID %d)", pid)
	}

	// Clean up stale PID file if present
	if err == nil {
		_ = removePidFile()
	}

	// Re-exec self in daemon mode
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build args: we assume the daemon invocation is:
	// nux col start --daemon
	args := []string{exe, "col", "start", "--daemon"}

	procAttr := &os.ProcAttr{
		Dir: "",
		Env: os.Environ(),
		Sys: nil,
	}

	proc, err := os.StartProcess(exe, args, procAttr)
	if err != nil {
		return 0, fmt.Errorf("failed to start daemon process: %w", err)
	}

	if err := proc.Release(); err != nil {
		return 0, fmt.Errorf("failed to release daemon process: %w", err)
	}

	return proc.Pid, nil
}

// runCollectorDaemon is intended to be called when running in daemon mode.
// It writes the PID file, runs the collector, and cleans up on exit.
func runCollectorDaemon(cmd *cobra.Command) error {
	cfg := getCollectorConfig(cmd)
	if cfg.IsEmpty() {
		return fmt.Errorf("collector config empty")
	}

	// Write PID file for this daemon process
	if err := writePidFile(os.Getpid()); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	defer removePidFile()

	// Set up signal handling
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c, err := collector.NewCollector(cfg)
	if err != nil {
		return fmt.Errorf("failed to create collector: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- c.Run(ctx)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		c.Shutdown()
		return nil
	}
}

// stopColProc stops the collector daemon by sending SIGTERM.
func stopColProc(cmd *cobra.Command) error {
	pid, err := readPidFile()
	if err != nil {
		return fmt.Errorf("collector not running (no PID file)")
	}

	if !processExists(pid) {
		_ = removePidFile()
		return fmt.Errorf("collector not running (stale PID file)")
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGTERM for graceful shutdown
	if err := p.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %w", err)
	}

	// Wait a bit for graceful shutdown
	for i := 0; i < 30; i++ {
		if !processExists(pid) {
			_ = removePidFile()

			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Force kill if still alive
	_ = p.Signal(unix.SIGKILL)
	_ = removePidFile()
	cmd.Println("collector force-killed")
	return nil
}

// findColProc returns the PID if the collector is running, or an error if not.
func findColProc(cmd *cobra.Command) (int, error) {
	pid, err := readPidFile()
	if err != nil {
		return 0, fmt.Errorf("collector not running (no PID file)")
	}

	if !processExists(pid) {
		_ = removePidFile()
		return 0, fmt.Errorf("collector not running (stale PID file)")
	}

	return pid, nil
}
