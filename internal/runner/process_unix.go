//go:build !windows

package runner

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

type unixProcessController struct {
	processGroupID int
}

func startControlledProcess(command *exec.Cmd) (processController, error) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := command.Start(); err != nil {
		return nil, err
	}
	return &unixProcessController{processGroupID: command.Process.Pid}, nil
}

func (controller *unixProcessController) Terminate(grace time.Duration) error {
	if controller.processGroupID <= 0 {
		return nil
	}
	if err := signalProcessGroup(controller.processGroupID, syscall.SIGTERM); err != nil {
		return err
	}
	if waitForProcessGroupExit(controller.processGroupID, grace) {
		return nil
	}
	if err := signalProcessGroup(controller.processGroupID, syscall.SIGKILL); err != nil {
		return err
	}
	if !waitForProcessGroupExit(controller.processGroupID, time.Second) {
		return fmt.Errorf("process group %d remained alive after SIGKILL", controller.processGroupID)
	}
	return nil
}

func (controller *unixProcessController) Close() error {
	return controller.Terminate(0)
}

func signalProcessGroup(processGroupID int, signal syscall.Signal) error {
	err := syscall.Kill(-processGroupID, signal)
	if err == nil || errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return fmt.Errorf("signal process group %d: %w", processGroupID, err)
}

func waitForProcessGroupExit(processGroupID int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		err := syscall.Kill(-processGroupID, 0)
		if errors.Is(err, syscall.ESRCH) {
			return true
		}
		if timeout <= 0 || time.Now().After(deadline) {
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func terminationDetail(command *exec.Cmd, _ error) string {
	if command.ProcessState == nil {
		return ""
	}
	status, ok := command.ProcessState.Sys().(syscall.WaitStatus)
	if ok && status.Signaled() {
		return status.Signal().String()
	}
	return ""
}
