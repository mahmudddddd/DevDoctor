package model

// ReportSchemaVersion identifies the JSON report contract emitted by this build.
const ReportSchemaVersion = "1.0"

// ProjectReport is the top-level versioned discovery result.
type ProjectReport struct {
	SchemaVersion string         `json:"schemaVersion"`
	ToolVersion   string         `json:"toolVersion"`
	Project       ProjectSummary `json:"project"`
}

// ProjectSummary contains technologies and metadata detected at a project root.
type ProjectSummary struct {
	Root            string                    `json:"root"`
	Name            string                    `json:"name"`
	Languages       []Detection               `json:"languages"`
	Frameworks      []Detection               `json:"frameworks"`
	Runtimes        []RuntimeDetection        `json:"runtimes"`
	PackageManagers []PackageManagerDetection `json:"packageManagers"`
	Workspaces      []WorkspaceDetection      `json:"workspaces"`
	RelevantFiles   []string                  `json:"relevantFiles"`
	Warnings        []Warning                 `json:"warnings"`
}

// Detection describes a technology inferred from one or more evidence items.
type Detection struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Confidence Confidence `json:"confidence"`
	Evidence   []Evidence `json:"evidence"`
}

// RuntimeDetection describes a detected runtime and its declared requirement.
type RuntimeDetection struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Requirement string     `json:"requirement,omitempty"`
	Confidence  Confidence `json:"confidence"`
	Evidence    []Evidence `json:"evidence"`
}

// PackageManagerDetection describes package-manager declarations and lockfiles.
type PackageManagerDetection struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	DeclaredVersion string     `json:"declaredVersion,omitempty"`
	Lockfiles       []string   `json:"lockfiles"`
	Confidence      Confidence `json:"confidence"`
	Evidence        []Evidence `json:"evidence"`
}

// WorkspaceDetection describes a detected multi-package workspace.
type WorkspaceDetection struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Patterns []string   `json:"patterns"`
	Evidence []Evidence `json:"evidence"`
}

// Evidence records the observable metadata behind a detection or warning.
type Evidence struct {
	Kind   string `json:"kind"`
	Path   string `json:"path,omitempty"`
	Detail string `json:"detail"`
}

// Warning records an ambiguity, policy skip, or unsupported project state.
type Warning struct {
	Code     string     `json:"code"`
	Message  string     `json:"message"`
	Evidence []Evidence `json:"evidence,omitempty"`
}

// Confidence communicates how strongly the available evidence supports a detection.
type Confidence string

// Supported confidence levels, from weakest to strongest evidence.
const (
	ConfidenceLow    Confidence = "low"
	ConfidenceMedium Confidence = "medium"
	ConfidenceHigh   Confidence = "high"
)
