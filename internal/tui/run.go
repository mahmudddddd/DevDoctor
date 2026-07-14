package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// RunOptions contains terminal streams for the production Bubble Tea program.
type RunOptions struct {
	Config Config
	Input  io.Reader
	Output io.Writer
}

// Run launches the full-screen alternate-screen DevDoctor shell.
func Run(options RunOptions) error {
	config := configForEnvironment(options.Config)
	ctx := config.Context
	if ctx == nil {
		ctx = context.Background()
	}
	program := tea.NewProgram(
		NewModel(config),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
		tea.WithInput(options.Input),
		tea.WithOutput(options.Output),
	)
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run terminal UI: %w", err)
	}
	return nil
}

func configForEnvironment(config Config) Config {
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		config.Color = false
	}
	if strings.EqualFold(os.Getenv("TERM"), "dumb") {
		config.ASCII = true
		config.Flat = true
		config.Color = false
	}
	return config
}
