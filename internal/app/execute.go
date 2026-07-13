package app

import (
	"context"
	"fmt"

	"github.com/mahmudddddd/DevDoctor/internal/consent"
	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
	"github.com/mahmudddddd/DevDoctor/internal/runner"
)

type commandPreparer interface {
	PrepareCommand(string, model.CommandSpec) (privacy.PreparedCommand, error)
}

type commandAuthorizer interface {
	Authorize(context.Context, privacy.PreparedCommand, []string) (consent.Decision, error)
}

type commandRunner interface {
	Run(context.Context, privacy.PreparedCommand) (model.CommandResult, error)
}

// ExecutionService orchestrates validation, exact consent, revalidation, and execution.
type ExecutionService struct {
	preparer   commandPreparer
	authorizer commandAuthorizer
	runner     commandRunner
}

// NewExecutionService creates the internal Phase 2 execution boundary.
func NewExecutionService(approver consent.Approver) ExecutionService {
	return ExecutionService{
		preparer:   privacy.NewPathPolicy(),
		authorizer: consent.NewManager(approver),
		runner:     runner.New(),
	}
}

// Execute validates and runs one structured command after exact approval.
func (service ExecutionService) Execute(ctx context.Context, projectRoot string, spec model.CommandSpec) (model.CommandResult, error) {
	if service.preparer == nil {
		return model.CommandResult{}, fmt.Errorf("command path policy is not configured")
	}
	if service.authorizer == nil {
		return model.CommandResult{}, fmt.Errorf("command approval policy is not configured")
	}
	if service.runner == nil {
		return model.CommandResult{}, fmt.Errorf("command runner is not configured")
	}

	prepared, err := service.preparer.PrepareCommand(projectRoot, spec)
	if err != nil {
		return model.CommandResult{}, fmt.Errorf("prepare command: %w", err)
	}
	environmentNames, err := runner.EnvironmentNames(prepared.Spec().Environment)
	if err != nil {
		return model.CommandResult{}, fmt.Errorf("validate command environment: %w", err)
	}
	decision, err := service.authorizer.Authorize(ctx, prepared, environmentNames)
	if err != nil {
		return model.CommandResult{}, fmt.Errorf("request command approval: %w", err)
	}
	switch decision.Outcome {
	case consent.Denied:
		return skippedResult(model.ReasonApprovalDenied), nil
	case consent.Unavailable:
		return skippedResult(model.ReasonApprovalUnavailable), nil
	case consent.Approved:
	default:
		return model.CommandResult{}, fmt.Errorf("unsupported approval outcome %q", decision.Outcome)
	}

	if err := prepared.Revalidate(); err != nil {
		return model.CommandResult{}, fmt.Errorf("revalidate approved command: %w", err)
	}
	return service.runner.Run(ctx, prepared)
}

func skippedResult(reason model.CommandReason) model.CommandResult {
	return model.CommandResult{
		Status:          model.CommandSkipped,
		Reason:          reason,
		CleanupComplete: true,
	}
}
