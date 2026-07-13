package privacy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

const (
	defaultCommandTimeout         = 2 * time.Minute
	defaultTerminationGrace       = 2 * time.Second
	defaultOutputLimit      int64 = 256 * 1024
	maxCommandTimeout             = 30 * time.Minute
	maxTerminationGrace           = 30 * time.Second
	maxOutputLimit          int64 = 16 * 1024 * 1024
)

// PathPolicy validates the filesystem identities bound to an approved command.
type PathPolicy struct{}

// PreparedCommand is an immutable command whose paths and limits were validated.
type PreparedCommand struct {
	spec               model.CommandSpec
	projectRoot        string
	rootInput          string
	workingInput       string
	executableInput    string
	rootIdentity       os.FileInfo
	workingIdentity    os.FileInfo
	executableIdentity os.FileInfo
}

// NewPathPolicy returns the default project and executable path policy.
func NewPathPolicy() PathPolicy {
	return PathPolicy{}
}

// PrepareCommand canonicalizes and validates a command before it is shown for consent.
func (PathPolicy) PrepareCommand(projectRoot string, requested model.CommandSpec) (PreparedCommand, error) {
	spec, err := normalizeCommandSpec(requested)
	if err != nil {
		return PreparedCommand{}, err
	}

	rootInput, root, rootInfo, err := canonicalDirectory(projectRoot)
	if err != nil {
		return PreparedCommand{}, fmt.Errorf("validate project root: %w", err)
	}

	workingInput := spec.WorkingDirectory
	if workingInput == "" {
		workingInput = root
	} else if !filepath.IsAbs(workingInput) {
		workingInput = filepath.Join(root, workingInput)
	}
	workingInput, working, workingInfo, err := canonicalDirectory(workingInput)
	if err != nil {
		return PreparedCommand{}, fmt.Errorf("validate working directory: %w", err)
	}
	if !pathWithin(root, working) {
		return PreparedCommand{}, fmt.Errorf("working directory is outside the project root")
	}

	if !filepath.IsAbs(spec.Executable) {
		return PreparedCommand{}, fmt.Errorf("executable must be an absolute path")
	}
	executableInput, executable, executableInfo, err := canonicalRegularFile(spec.Executable)
	if err != nil {
		return PreparedCommand{}, fmt.Errorf("validate executable: %w", err)
	}

	spec.Executable = executable
	spec.WorkingDirectory = working

	return PreparedCommand{
		spec:               cloneCommandSpec(spec),
		projectRoot:        root,
		rootInput:          rootInput,
		workingInput:       workingInput,
		executableInput:    executableInput,
		rootIdentity:       rootInfo,
		workingIdentity:    workingInfo,
		executableIdentity: executableInfo,
	}, nil
}

// Revalidate confirms that the exact approved filesystem identities still exist.
func (p PreparedCommand) Revalidate() error {
	_, root, rootInfo, err := canonicalDirectory(p.rootInput)
	if err != nil {
		return fmt.Errorf("revalidate project root: %w", err)
	}
	if root != p.projectRoot || !os.SameFile(rootInfo, p.rootIdentity) {
		return fmt.Errorf("project root identity changed after approval")
	}

	_, working, workingInfo, err := canonicalDirectory(p.workingInput)
	if err != nil {
		return fmt.Errorf("revalidate working directory: %w", err)
	}
	if working != p.spec.WorkingDirectory || !pathWithin(root, working) || !os.SameFile(workingInfo, p.workingIdentity) {
		return fmt.Errorf("working directory identity changed after approval")
	}

	_, executable, executableInfo, err := canonicalRegularFile(p.executableInput)
	if err != nil {
		return fmt.Errorf("revalidate executable: %w", err)
	}
	if executable != p.spec.Executable || !os.SameFile(executableInfo, p.executableIdentity) {
		return fmt.Errorf("executable identity changed after approval")
	}
	return nil
}

// Spec returns a defensive copy of the exact command approved by the user.
func (p PreparedCommand) Spec() model.CommandSpec {
	return cloneCommandSpec(p.spec)
}

// ProjectRoot returns the canonical selected project root.
func (p PreparedCommand) ProjectRoot() string {
	return p.projectRoot
}

func canonicalDirectory(path string) (string, string, os.FileInfo, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", "", nil, fmt.Errorf("make path absolute: %w", err)
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve path: %w", err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return "", "", nil, fmt.Errorf("make resolved path absolute: %w", err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return "", "", nil, fmt.Errorf("inspect path: %w", err)
	}
	if !info.IsDir() {
		return "", "", nil, fmt.Errorf("path is not a directory")
	}
	return absolute, filepath.Clean(canonical), info, nil
}

func canonicalRegularFile(path string) (string, string, os.FileInfo, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", "", nil, fmt.Errorf("make path absolute: %w", err)
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", "", nil, fmt.Errorf("resolve path: %w", err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return "", "", nil, fmt.Errorf("make resolved path absolute: %w", err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return "", "", nil, fmt.Errorf("inspect path: %w", err)
	}
	if !info.Mode().IsRegular() {
		return "", "", nil, fmt.Errorf("path is not a regular file")
	}
	return absolute, filepath.Clean(canonical), info, nil
}

func pathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}

func normalizeCommandSpec(spec model.CommandSpec) (model.CommandSpec, error) {
	spec.OperationID = strings.TrimSpace(spec.OperationID)
	spec.Purpose = strings.TrimSpace(spec.Purpose)
	if spec.OperationID == "" {
		return model.CommandSpec{}, fmt.Errorf("operation ID is required")
	}
	if spec.Purpose == "" {
		return model.CommandSpec{}, fmt.Errorf("command purpose is required")
	}
	if spec.Timeout == 0 {
		spec.Timeout = defaultCommandTimeout
	}
	if spec.TerminationGrace == 0 {
		spec.TerminationGrace = defaultTerminationGrace
	}
	if spec.OutputLimit == 0 {
		spec.OutputLimit = defaultOutputLimit
	}
	if spec.Timeout < 0 || spec.Timeout > maxCommandTimeout {
		return model.CommandSpec{}, fmt.Errorf("timeout must be between zero and %s", maxCommandTimeout)
	}
	if spec.TerminationGrace < 0 || spec.TerminationGrace > maxTerminationGrace {
		return model.CommandSpec{}, fmt.Errorf("termination grace must be between zero and %s", maxTerminationGrace)
	}
	if spec.OutputLimit < 1 || spec.OutputLimit > maxOutputLimit {
		return model.CommandSpec{}, fmt.Errorf("output limit must be between 1 and %d bytes", maxOutputLimit)
	}
	if !validMutation(spec.Mutation) {
		return model.CommandSpec{}, fmt.Errorf("invalid mutation classification %q", spec.Mutation)
	}
	if !validNetwork(spec.Network) {
		return model.CommandSpec{}, fmt.Errorf("invalid network classification %q", spec.Network)
	}
	return cloneCommandSpec(spec), nil
}

func validMutation(value model.MutationClass) bool {
	switch value {
	case model.MutationNone, model.MutationProject, model.MutationSystem, model.MutationUnknown:
		return true
	default:
		return false
	}
}

func validNetwork(value model.NetworkClass) bool {
	switch value {
	case model.NetworkNone, model.NetworkRead, model.NetworkWrite, model.NetworkUnknown:
		return true
	default:
		return false
	}
}

func cloneCommandSpec(spec model.CommandSpec) model.CommandSpec {
	spec.Arguments = append([]string(nil), spec.Arguments...)
	spec.Environment.Pass = append([]string(nil), spec.Environment.Pass...)
	if spec.Environment.Set != nil {
		values := spec.Environment.Set
		spec.Environment.Set = make(map[string]string, len(values))
		for name, value := range values {
			spec.Environment.Set[name] = value
		}
	}
	spec.DataBundles = append([]model.DataBundleDescriptor(nil), spec.DataBundles...)
	return spec
}
