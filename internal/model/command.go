package model

import "time"

// MutationClass describes where a command may change state.
type MutationClass string

// Supported mutation classifications.
const (
	MutationNone    MutationClass = "none"
	MutationProject MutationClass = "project"
	MutationSystem  MutationClass = "system"
	MutationUnknown MutationClass = "unknown"
)

// NetworkClass describes a command's declared network access.
type NetworkClass string

// Supported network classifications.
const (
	NetworkNone    NetworkClass = "none"
	NetworkRead    NetworkClass = "read"
	NetworkWrite   NetworkClass = "write"
	NetworkUnknown NetworkClass = "unknown"
)

// CommandStatus is the terminal lifecycle state of an approved command.
type CommandStatus string

// Supported command lifecycle states.
const (
	CommandCompleted CommandStatus = "completed"
	CommandFailed    CommandStatus = "failed"
	CommandSkipped   CommandStatus = "skipped"
	CommandCancelled CommandStatus = "cancelled"
	CommandTimedOut  CommandStatus = "timed_out"
)

// CommandReason identifies why a command reached its terminal state.
type CommandReason string

// Stable command result reason codes.
const (
	ReasonApprovalDenied      CommandReason = "approval_denied"
	ReasonApprovalUnavailable CommandReason = "approval_unavailable"
	ReasonStartFailed         CommandReason = "start_failed"
	ReasonNonzeroExit         CommandReason = "nonzero_exit"
	ReasonCaptureFailed       CommandReason = "capture_failed"
	ReasonCleanupIncomplete   CommandReason = "cleanup_incomplete"
	ReasonCancelled           CommandReason = "cancelled"
	ReasonTimedOut            CommandReason = "timed_out"
)

// EnvironmentSpec declares the small environment surface supplied to a command.
type EnvironmentSpec struct {
	Pass []string          `json:"pass,omitempty"`
	Set  map[string]string `json:"-"`
}

// DataBundleDescriptor names approved data made available to a command.
type DataBundleDescriptor struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CommandSpec is an untrusted request that must be prepared and approved before execution.
type CommandSpec struct {
	OperationID      string                 `json:"operationId"`
	Purpose          string                 `json:"purpose"`
	Executable       string                 `json:"executable"`
	Arguments        []string               `json:"arguments"`
	WorkingDirectory string                 `json:"workingDirectory"`
	Timeout          time.Duration          `json:"timeout"`
	TerminationGrace time.Duration          `json:"terminationGrace"`
	OutputLimit      int64                  `json:"outputLimit"`
	Environment      EnvironmentSpec        `json:"environment"`
	Mutation         MutationClass          `json:"mutation"`
	Network          NetworkClass           `json:"network"`
	StartsService    bool                   `json:"startsService"`
	DataBundles      []DataBundleDescriptor `json:"dataBundles"`
}

// StreamCapture contains one independently bounded process stream.
type StreamCapture struct {
	Text           string `json:"text"`
	CapturedBytes  int64  `json:"capturedBytes"`
	DiscardedBytes int64  `json:"discardedBytes"`
	Truncated      bool   `json:"truncated"`
}

// CommandResult records a command's terminal state without changing the Phase 1 report schema.
type CommandResult struct {
	Status          CommandStatus `json:"status"`
	Reason          CommandReason `json:"reason,omitempty"`
	ExitCode        *int          `json:"exitCode,omitempty"`
	Termination     string        `json:"termination,omitempty"`
	Stdout          StreamCapture `json:"stdout"`
	Stderr          StreamCapture `json:"stderr"`
	Duration        time.Duration `json:"duration"`
	CleanupComplete bool          `json:"cleanupComplete"`
	CleanupError    string        `json:"cleanupError,omitempty"`
}
