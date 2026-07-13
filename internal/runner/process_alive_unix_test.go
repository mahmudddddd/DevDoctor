//go:build !windows

package runner

import (
	"errors"
	"os/signal"
	"syscall"
	"time"
)

func ignoreTerminationSignal() {
	signal.Ignore(syscall.SIGTERM)
}

func waitForProcessExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return errors.Is(syscall.Kill(pid, 0), syscall.ESRCH)
}
