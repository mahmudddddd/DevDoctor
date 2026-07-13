package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/mahmudddddd/DevDoctor/internal/consent"
	"github.com/mahmudddddd/DevDoctor/internal/report"
)

// InteractiveApprover displays and approves one exact command in a terminal.
type InteractiveApprover struct {
	input  io.Reader
	output io.Writer
}

// NewInteractiveApprover creates a Huh-based exact consent adapter.
func NewInteractiveApprover(input io.Reader, output io.Writer) InteractiveApprover {
	return InteractiveApprover{input: input, output: output}
}

// Approve renders all material command properties before asking for a scope.
func (approver InteractiveApprover) Approve(ctx context.Context, request consent.Request) (consent.Decision, error) {
	if approver.input == nil || approver.output == nil {
		return consent.Decision{Outcome: consent.Unavailable}, nil
	}
	if _, err := fmt.Fprint(approver.output, formatApprovalRequest(request)); err != nil {
		return consent.Decision{}, fmt.Errorf("display command approval: %w", err)
	}

	choice := "deny"
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Allow this exact command?").
				Options(
					huh.NewOption("Approve once", "once"),
					huh.NewOption("Approve this exact check", "check"),
					huh.NewOption("Approve this exact request for this run", "run"),
					huh.NewOption("Deny", "deny"),
				).
				Value(&choice),
		),
	).WithInput(approver.input).WithOutput(approver.output).RunWithContext(ctx)
	if err != nil {
		return consent.Decision{}, fmt.Errorf("request command approval: %w", err)
	}

	switch choice {
	case "once":
		return consent.Decision{Outcome: consent.Approved, Scope: consent.ScopeOnce}, nil
	case "check":
		return consent.Decision{Outcome: consent.Approved, Scope: consent.ScopeThisCheck}, nil
	case "run":
		return consent.Decision{Outcome: consent.Approved, Scope: consent.ScopeThisRun}, nil
	default:
		return consent.Decision{Outcome: consent.Denied}, nil
	}
}

// NonInteractiveApprover always fails closed without reading input.
type NonInteractiveApprover struct{}

// Approve reports that approval is unavailable in non-interactive execution.
func (NonInteractiveApprover) Approve(context.Context, consent.Request) (consent.Decision, error) {
	return consent.Decision{Outcome: consent.Unavailable}, nil
}

func formatApprovalRequest(request consent.Request) string {
	var output strings.Builder
	fmt.Fprintln(&output, "\nCommand approval required")
	fmt.Fprintf(&output, "Operation: %s\n", report.SafeText(request.OperationID()))
	fmt.Fprintf(&output, "Purpose: %s\n", report.SafeText(request.Purpose()))
	fmt.Fprintf(&output, "Executable: %s\n", report.SafeText(request.Executable()))
	fmt.Fprintln(&output, "Arguments:")
	arguments := request.Arguments()
	if len(arguments) == 0 {
		fmt.Fprintln(&output, "  (none)")
	}
	for index, argument := range arguments {
		fmt.Fprintf(&output, "  %d: %s\n", index+1, report.SafeText(strconv.Quote(argument)))
	}
	fmt.Fprintf(&output, "Working directory: %s\n", report.SafeText(request.WorkingDirectory()))
	fmt.Fprintf(&output, "Mutation: %s\n", request.Mutation())
	fmt.Fprintf(&output, "Network: %s\n", request.Network())
	fmt.Fprintf(&output, "Starts service: %t\n", request.StartsService())
	fmt.Fprintf(&output, "Timeout: %s\n", request.Timeout())
	fmt.Fprintf(&output, "Termination grace: %s\n", request.TerminationGrace())
	fmt.Fprintf(&output, "Output limit: %d bytes per stream\n", request.OutputLimit())
	fmt.Fprintln(&output, "Environment names:")
	for _, name := range request.EnvironmentNames() {
		fmt.Fprintf(&output, "  - %s\n", report.SafeText(name))
	}
	fmt.Fprintln(&output, "Data bundles:")
	bundles := request.DataBundles()
	if len(bundles) == 0 {
		fmt.Fprintln(&output, "  (none)")
	}
	for _, bundle := range bundles {
		fmt.Fprintf(&output, "  - %s", report.SafeText(bundle.Name))
		if bundle.Description != "" {
			fmt.Fprintf(&output, ": %s", report.SafeText(bundle.Description))
		}
		fmt.Fprintln(&output)
	}
	return output.String()
}
