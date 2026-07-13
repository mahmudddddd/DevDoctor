package runner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

// Runner executes prepared commands without a shell and owns their process trees.
type Runner struct{}

// New creates a structured command runner.
func New() Runner {
	return Runner{}
}

// Run revalidates and executes an already approved command.
func (Runner) Run(ctx context.Context, prepared privacy.PreparedCommand) (model.CommandResult, error) {
	startedAt := time.Now()
	if err := prepared.Revalidate(); err != nil {
		return model.CommandResult{}, err
	}

	spec := prepared.Spec()
	environment, err := buildEnvironment(spec.Environment)
	if err != nil {
		return model.CommandResult{}, fmt.Errorf("build command environment: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return contextResult(err, time.Since(startedAt)), nil
	}

	stdout := newBoundedCapture(spec.OutputLimit)
	stderr := newBoundedCapture(spec.OutputLimit)
	command := exec.Command(spec.Executable, spec.Arguments...)
	command.Dir = spec.WorkingDirectory
	command.Env = environment
	command.Stdout = stdout
	command.Stderr = stderr

	controller, err := startControlledProcess(command)
	if err != nil {
		return model.CommandResult{
			Status:          model.CommandFailed,
			Reason:          model.ReasonStartFailed,
			Duration:        time.Since(startedAt),
			CleanupComplete: true,
		}, nil
	}

	waitResult := make(chan error, 1)
	go func() {
		waitResult <- command.Wait()
	}()

	timer := time.NewTimer(spec.Timeout)
	defer timer.Stop()

	var status model.CommandStatus
	var reason model.CommandReason
	var waitErr error
	var cleanupErr error

	select {
	case waitErr = <-waitResult:
		status, reason = classifyExit(waitErr)
	default:
		select {
		case waitErr = <-waitResult:
			status, reason = classifyExit(waitErr)
		case <-ctx.Done():
			status, reason = classifyContext(ctx.Err())
			cleanupErr = controller.Terminate(spec.TerminationGrace)
			waitErr = <-waitResult
		case <-timer.C:
			status = model.CommandTimedOut
			reason = model.ReasonTimedOut
			cleanupErr = controller.Terminate(spec.TerminationGrace)
			waitErr = <-waitResult
		}
	}

	closeErr := controller.Close()
	cleanupErr = errors.Join(cleanupErr, closeErr)

	result := model.CommandResult{
		Status:          status,
		Reason:          reason,
		ExitCode:        commandExitCode(command),
		Termination:     terminationDetail(command, waitErr),
		Stdout:          stdout.Result(),
		Stderr:          stderr.Result(),
		Duration:        time.Since(startedAt),
		CleanupComplete: cleanupErr == nil,
	}
	if cleanupErr != nil {
		result.CleanupError = cleanupErr.Error()
		return result, fmt.Errorf("command process-tree cleanup: %w", cleanupErr)
	}
	if status == model.CommandFailed && waitErr != nil {
		var exitError *exec.ExitError
		if !errors.As(waitErr, &exitError) {
			result.Reason = model.ReasonCaptureFailed
			return result, fmt.Errorf("wait for command: %w", waitErr)
		}
	}
	return result, nil
}

func classifyExit(err error) (model.CommandStatus, model.CommandReason) {
	if err == nil {
		return model.CommandCompleted, ""
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return model.CommandFailed, model.ReasonNonzeroExit
	}
	return model.CommandFailed, model.ReasonCaptureFailed
}

func classifyContext(err error) (model.CommandStatus, model.CommandReason) {
	if errors.Is(err, context.DeadlineExceeded) {
		return model.CommandTimedOut, model.ReasonTimedOut
	}
	return model.CommandCancelled, model.ReasonCancelled
}

func contextResult(err error, duration time.Duration) model.CommandResult {
	status, reason := classifyContext(err)
	return model.CommandResult{
		Status:          status,
		Reason:          reason,
		Duration:        duration,
		CleanupComplete: true,
	}
}

func commandExitCode(command *exec.Cmd) *int {
	if command.ProcessState == nil {
		return nil
	}
	exitCode := command.ProcessState.ExitCode()
	return &exitCode
}
