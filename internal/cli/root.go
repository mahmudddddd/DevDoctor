package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mahmudddddd/DevDoctor/internal/app"
	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/report"
	"github.com/mahmudddddd/DevDoctor/internal/version"
)

// DiscoveryService supplies project reports to CLI commands.
type DiscoveryService interface {
	Diagnose(context.Context, string) (model.ProjectReport, error)
}

// Dependencies contains injectable terminal and discovery dependencies.
type Dependencies struct {
	Input          io.Reader
	Output         io.Writer
	ErrorOutput    io.Writer
	CurrentDir     func() (string, error)
	IsInteractive  func() bool
	UseColor       func() bool
	Discovery      DiscoveryService
	RunInteractive func(context.Context, io.Reader, io.Writer, string, bool, DiscoveryService) error
}

// Execute runs the DevDoctor command with process-standard dependencies.
func Execute() error {
	command := NewRootCommand(Dependencies{})
	return command.Execute()
}

// NewRootCommand constructs the DevDoctor command tree.
func NewRootCommand(dependencies Dependencies) *cobra.Command {
	dependencies = withDefaults(dependencies)

	root := &cobra.Command{
		Use:           "devdoctor",
		Short:         "Diagnose why software projects fail to build or start",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if !dependencies.IsInteractive() {
				return fmt.Errorf("interactive mode requires a terminal; use 'devdoctor diagnose --path <directory>'")
			}
			currentDirectory, err := dependencies.CurrentDir()
			if err != nil {
				return fmt.Errorf("determine current directory: %w", err)
			}
			err = dependencies.RunInteractive(
				command.Context(),
				command.InOrStdin(),
				command.OutOrStdout(),
				currentDirectory,
				dependencies.UseColor(),
				dependencies.Discovery,
			)
			return err
		},
	}
	root.SetIn(dependencies.Input)
	root.SetOut(dependencies.Output)
	root.SetErr(dependencies.ErrorOutput)
	root.AddCommand(newDiagnoseCommand(dependencies))
	root.AddCommand(newVersionCommand())
	return root
}

func newDiagnoseCommand(dependencies Dependencies) *cobra.Command {
	var projectPath string
	var outputFormat string

	command := &cobra.Command{
		Use:   "diagnose",
		Short: "Safely inspect project metadata",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			return runDiagnosis(command.Context(), dependencies, projectPath, outputFormat)
		},
	}
	command.Flags().StringVarP(&projectPath, "path", "p", ".", "project directory to inspect")
	command.Flags().StringVarP(&outputFormat, "format", "f", "text", "output format: text or json")
	return command
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show DevDoctor build information",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(command.OutOrStdout(), version.String())
			return err
		},
	}
}

func runDiagnosis(ctx context.Context, dependencies Dependencies, path, outputFormat string) error {
	outputFormat = strings.ToLower(strings.TrimSpace(outputFormat))
	if outputFormat != "text" && outputFormat != "json" {
		return fmt.Errorf("unsupported format %q; expected text or json", outputFormat)
	}

	discoveryReport, err := dependencies.Discovery.Diagnose(ctx, path)
	if err != nil {
		return fmt.Errorf("discover project: %w", err)
	}

	if outputFormat == "json" {
		return report.WriteJSON(dependencies.Output, discoveryReport)
	}
	return report.WriteText(dependencies.Output, discoveryReport, report.TextOptions{Color: dependencies.UseColor()})
}

func withDefaults(dependencies Dependencies) Dependencies {
	if dependencies.Input == nil {
		dependencies.Input = os.Stdin
	}
	if dependencies.Output == nil {
		dependencies.Output = os.Stdout
	}
	if dependencies.ErrorOutput == nil {
		dependencies.ErrorOutput = os.Stderr
	}
	if dependencies.CurrentDir == nil {
		dependencies.CurrentDir = os.Getwd
	}
	if dependencies.IsInteractive == nil {
		dependencies.IsInteractive = func() bool {
			return isInteractiveTerminal(dependencies.Input, dependencies.Output)
		}
	}
	if dependencies.UseColor == nil {
		dependencies.UseColor = func() bool {
			_, noColor := os.LookupEnv("NO_COLOR")
			return !noColor && isTerminal(dependencies.Output)
		}
	}
	if dependencies.Discovery == nil {
		service := app.NewDiscoveryService()
		dependencies.Discovery = service
	}
	if dependencies.RunInteractive == nil {
		dependencies.RunInteractive = runInteractive
	}
	return dependencies
}
