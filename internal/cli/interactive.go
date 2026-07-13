package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/huh"
)

var errInteractiveExit = errors.New("interactive menu exited")

func runInteractive(ctx context.Context, input io.Reader, output io.Writer, currentDirectory string, diagnose func(context.Context, string, string) error) error {
	action := "diagnose"
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What would you like to do?").
				Options(
					huh.NewOption("Diagnose a project", "diagnose"),
					huh.NewOption("Exit", "exit"),
				).
				Value(&action),
		),
	).WithInput(input).WithOutput(output).RunWithContext(ctx); err != nil {
		return fmt.Errorf("run interactive menu: %w", err)
	}
	if action == "exit" {
		return errInteractiveExit
	}

	projectPath := currentDirectory
	if err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project directory").
				Description("DevDoctor will only inspect safe project metadata in Phase 1.").
				Value(&projectPath),
		),
	).WithInput(input).WithOutput(output).RunWithContext(ctx); err != nil {
		return fmt.Errorf("choose project directory: %w", err)
	}

	return diagnose(ctx, projectPath, "text")
}
