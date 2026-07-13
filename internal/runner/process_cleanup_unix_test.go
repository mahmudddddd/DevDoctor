//go:build !windows

package runner

import (
	"context"
	"testing"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

func TestRunnerEscalatesWhenProcessIgnoresSIGTERM(t *testing.T) {
	t.Parallel()
	spec := helperSpec(t, "ignore-term")
	spec.Timeout = 100 * time.Millisecond
	spec.TerminationGrace = 50 * time.Millisecond
	result, err := runHelper(context.Background(), t, spec)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != model.CommandTimedOut || !result.CleanupComplete {
		t.Fatalf("result = %+v", result)
	}
}
