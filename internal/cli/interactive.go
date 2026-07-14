package cli

import (
	"context"
	"io"

	"github.com/mahmudddddd/DevDoctor/internal/tui"
)

func runInteractive(ctx context.Context, input io.Reader, output io.Writer, currentDirectory string, color bool, discovery DiscoveryService) error {
	return tui.Run(tui.RunOptions{
		Config: tui.Config{
			Context:     ctx,
			ProjectPath: currentDirectory,
			Discover:    discovery.Diagnose,
			Color:       color,
		},
		Input:  input,
		Output: output,
	})
}
