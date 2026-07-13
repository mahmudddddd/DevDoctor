package consent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

// Outcome is the fail-closed result of an approval request.
type Outcome string

// Supported approval outcomes.
const (
	Approved    Outcome = "approved"
	Denied      Outcome = "denied"
	Unavailable Outcome = "unavailable"
)

// Scope controls the in-memory lifetime of an exact approval.
type Scope string

// Supported approval scopes.
const (
	ScopeOnce      Scope = "once"
	ScopeThisCheck Scope = "this_check"
	ScopeThisRun   Scope = "this_run"
)

// Decision is returned by an approval adapter.
type Decision struct {
	Outcome Outcome
	Scope   Scope
}

// Approver decides whether one immutable request may run.
type Approver interface {
	Approve(context.Context, Request) (Decision, error)
}

// Request is an immutable, value-redacted description of one prepared command.
type Request struct {
	operationID      string
	purpose          string
	executable       string
	arguments        []string
	workingDirectory string
	mutation         model.MutationClass
	network          model.NetworkClass
	startsService    bool
	timeout          time.Duration
	terminationGrace time.Duration
	outputLimit      int64
	environmentNames []string
	dataBundles      []model.DataBundleDescriptor
	fingerprint      string
}

// Manager applies exact in-memory approval scopes around an Approver.
type Manager struct {
	approver Approver
	mu       sync.Mutex
	grants   map[string]Scope
}

// NewManager creates a fail-closed approval manager.
func NewManager(approver Approver) *Manager {
	return &Manager{approver: approver, grants: make(map[string]Scope)}
}

// Authorize approves the exact prepared command or returns a denial/unavailable decision.
func (manager *Manager) Authorize(ctx context.Context, prepared privacy.PreparedCommand, environmentNames []string) (Decision, error) {
	if manager == nil || manager.approver == nil {
		return Decision{Outcome: Unavailable}, nil
	}
	request, err := newRequest(prepared, environmentNames)
	if err != nil {
		return Decision{}, err
	}

	manager.mu.Lock()
	cachedScope, cached := manager.grants[request.fingerprint]
	manager.mu.Unlock()
	if cached {
		return Decision{Outcome: Approved, Scope: cachedScope}, nil
	}

	decision, err := manager.approver.Approve(ctx, request)
	if err != nil {
		return Decision{}, err
	}
	if decision.Outcome != Approved {
		if decision.Outcome != Denied && decision.Outcome != Unavailable {
			return Decision{}, fmt.Errorf("invalid approval outcome %q", decision.Outcome)
		}
		return decision, nil
	}
	if decision.Scope != ScopeOnce && decision.Scope != ScopeThisCheck && decision.Scope != ScopeThisRun {
		return Decision{}, fmt.Errorf("invalid approval scope %q", decision.Scope)
	}
	if decision.Scope != ScopeOnce {
		manager.mu.Lock()
		manager.grants[request.fingerprint] = decision.Scope
		manager.mu.Unlock()
	}
	return decision, nil
}

func newRequest(prepared privacy.PreparedCommand, environmentNames []string) (Request, error) {
	spec := prepared.Spec()
	valueFingerprints := make(map[string]string, len(spec.Environment.Set))
	for name, value := range spec.Environment.Set {
		digest := sha256.Sum256([]byte(value))
		valueFingerprints[name] = hex.EncodeToString(digest[:])
	}
	material := struct {
		ProjectRoot                  string
		Spec                         model.CommandSpec
		EnvironmentNames             []string
		EnvironmentValueFingerprints map[string]string
	}{
		ProjectRoot:                  prepared.ProjectRoot(),
		Spec:                         spec,
		EnvironmentNames:             append([]string(nil), environmentNames...),
		EnvironmentValueFingerprints: valueFingerprints,
	}
	encoded, err := json.Marshal(material)
	if err != nil {
		return Request{}, fmt.Errorf("fingerprint approval request: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return Request{
		operationID:      spec.OperationID,
		purpose:          spec.Purpose,
		executable:       spec.Executable,
		arguments:        append([]string(nil), spec.Arguments...),
		workingDirectory: spec.WorkingDirectory,
		mutation:         spec.Mutation,
		network:          spec.Network,
		startsService:    spec.StartsService,
		timeout:          spec.Timeout,
		terminationGrace: spec.TerminationGrace,
		outputLimit:      spec.OutputLimit,
		environmentNames: append([]string(nil), environmentNames...),
		dataBundles:      append([]model.DataBundleDescriptor(nil), spec.DataBundles...),
		fingerprint:      hex.EncodeToString(digest[:]),
	}, nil
}

// OperationID returns the stable check or operation identifier.
func (request Request) OperationID() string { return request.operationID }

// Purpose returns the beginner-facing reason for the command.
func (request Request) Purpose() string { return request.purpose }

// Executable returns the exact canonical executable path.
func (request Request) Executable() string { return request.executable }

// Arguments returns a defensive copy of the exact argument vector.
func (request Request) Arguments() []string { return append([]string(nil), request.arguments...) }

// WorkingDirectory returns the exact canonical working directory.
func (request Request) WorkingDirectory() string { return request.workingDirectory }

// Mutation returns the declared mutation classification.
func (request Request) Mutation() model.MutationClass { return request.mutation }

// Network returns the declared network classification.
func (request Request) Network() model.NetworkClass { return request.network }

// StartsService reports whether the command declares a long-running service start.
func (request Request) StartsService() bool { return request.startsService }

// Timeout returns the approved execution timeout.
func (request Request) Timeout() time.Duration { return request.timeout }

// TerminationGrace returns the approved graceful termination window.
func (request Request) TerminationGrace() time.Duration { return request.terminationGrace }

// OutputLimit returns the independent per-stream capture limit.
func (request Request) OutputLimit() int64 { return request.outputLimit }

// EnvironmentNames returns only environment variable names, never values.
func (request Request) EnvironmentNames() []string {
	return append([]string(nil), request.environmentNames...)
}

// DataBundles returns a defensive copy of declared data descriptors.
func (request Request) DataBundles() []model.DataBundleDescriptor {
	return append([]model.DataBundleDescriptor(nil), request.dataBundles...)
}
