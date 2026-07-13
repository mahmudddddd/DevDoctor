//go:build windows

package runner

import (
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type windowsProcessController struct {
	job windows.Handle
}

func startControlledProcess(command *exec.Cmd) (processController, error) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
	if err := command.Start(); err != nil {
		return nil, err
	}

	job, err := createKillOnCloseJob()
	if err != nil {
		return nil, abortSuspendedProcess(command, err)
	}
	process, err := windows.OpenProcess(
		windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE|windows.PROCESS_QUERY_INFORMATION|windows.SYNCHRONIZE,
		false,
		uint32(command.Process.Pid),
	)
	if err != nil {
		_ = windows.CloseHandle(job)
		return nil, abortSuspendedProcess(command, fmt.Errorf("open child process: %w", err))
	}
	defer func() { _ = windows.CloseHandle(process) }()

	if err := windows.AssignProcessToJobObject(job, process); err != nil {
		_ = windows.CloseHandle(job)
		return nil, abortSuspendedProcess(command, fmt.Errorf("assign child to job object: %w", err))
	}
	if err := resumeProcessThreads(uint32(command.Process.Pid)); err != nil {
		_ = windows.TerminateJobObject(job, 1)
		_ = windows.CloseHandle(job)
		return nil, abortSuspendedProcess(command, fmt.Errorf("resume child process: %w", err))
	}
	return &windowsProcessController{job: job}, nil
}

func createKillOnCloseJob() (windows.Handle, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return 0, fmt.Errorf("create job object: %w", err)
	}
	information := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	information.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&information)),
		uint32(unsafe.Sizeof(information)),
	)
	if err != nil {
		_ = windows.CloseHandle(job)
		return 0, fmt.Errorf("configure job object: %w", err)
	}
	return job, nil
}

func resumeProcessThreads(processID uint32) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return err
	}
	defer func() { _ = windows.CloseHandle(snapshot) }()

	entry := windows.ThreadEntry32{Size: uint32(unsafe.Sizeof(windows.ThreadEntry32{}))}
	if err := windows.Thread32First(snapshot, &entry); err != nil {
		return err
	}
	resumed := false
	for {
		if entry.OwnerProcessID == processID {
			thread, openErr := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, entry.ThreadID)
			if openErr != nil {
				return openErr
			}
			_, resumeErr := windows.ResumeThread(thread)
			closeErr := windows.CloseHandle(thread)
			if resumeErr != nil {
				return resumeErr
			}
			if closeErr != nil {
				return closeErr
			}
			resumed = true
		}
		if err := windows.Thread32Next(snapshot, &entry); err != nil {
			if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
				break
			}
			return err
		}
	}
	if !resumed {
		return fmt.Errorf("no suspended thread found for process %d", processID)
	}
	return nil
}

func abortSuspendedProcess(command *exec.Cmd, cause error) error {
	if command.Process != nil {
		_ = command.Process.Kill()
		_ = command.Wait()
	}
	return cause
}

func (controller *windowsProcessController) Terminate(_ time.Duration) error {
	if controller.job == 0 {
		return nil
	}
	if err := windows.TerminateJobObject(controller.job, 1); err != nil && !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return fmt.Errorf("terminate job object: %w", err)
	}
	return nil
}

func (controller *windowsProcessController) Close() error {
	if controller.job == 0 {
		return nil
	}
	err := windows.CloseHandle(controller.job)
	controller.job = 0
	if err != nil {
		return fmt.Errorf("close job object: %w", err)
	}
	return nil
}

func terminationDetail(command *exec.Cmd, _ error) string {
	if command.ProcessState == nil || command.ProcessState.Success() {
		return ""
	}
	return fmt.Sprintf("exit code %d", command.ProcessState.ExitCode())
}
