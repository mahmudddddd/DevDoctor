//go:build windows

package runner

import (
	"time"

	"golang.org/x/sys/windows"
)

const stillActiveExitCode = 259

func ignoreTerminationSignal() {}

func waitForProcessExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		process, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
		if err != nil {
			return true
		}
		var exitCode uint32
		err = windows.GetExitCodeProcess(process, &exitCode)
		_ = windows.CloseHandle(process)
		if err != nil || exitCode != stillActiveExitCode {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}
