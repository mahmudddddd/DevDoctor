package detect

import (
	"container/heap"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/mahmudddddd/DevDoctor/internal/model"
	"github.com/mahmudddddd/DevDoctor/internal/privacy"
)

const maxRootEntries = 10_000

// ProjectDetector identifies supported project technologies from safe root metadata.
type ProjectDetector struct {
	policy privacy.FilePolicy
}

type packageManifest struct {
	Name                 string            `json:"name"`
	PackageManager       string            `json:"packageManager"`
	Engines              map[string]string `json:"engines"`
	Dependencies         map[string]string `json:"dependencies"`
	DevDependencies      map[string]string `json:"devDependencies"`
	PeerDependencies     map[string]string `json:"peerDependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	Workspaces           json.RawMessage   `json:"workspaces"`
}

type frameworkMarker struct {
	ID           string
	Name         string
	Dependencies []string
	Configs      []string
}

type packageManagerMarker struct {
	ID        string
	Name      string
	Lockfiles []string
}

var frameworkMarkers = []frameworkMarker{
	{ID: "angular", Name: "Angular", Dependencies: []string{"@angular/core"}, Configs: []string{"angular.json"}},
	{ID: "astro", Name: "Astro", Dependencies: []string{"astro"}, Configs: []string{"astro.config."}},
	{ID: "electron", Name: "Electron", Dependencies: []string{"electron"}},
	{ID: "express", Name: "Express", Dependencies: []string{"express"}},
	{ID: "fastify", Name: "Fastify", Dependencies: []string{"fastify"}},
	{ID: "nestjs", Name: "NestJS", Dependencies: []string{"@nestjs/core"}, Configs: []string{"nest-cli.json"}},
	{ID: "nextjs", Name: "Next.js", Dependencies: []string{"next"}, Configs: []string{"next.config."}},
	{ID: "nuxt", Name: "Nuxt", Dependencies: []string{"nuxt"}, Configs: []string{"nuxt.config."}},
	{ID: "react", Name: "React", Dependencies: []string{"react"}},
	{ID: "remix", Name: "Remix", Dependencies: []string{"@remix-run/react", "@remix-run/node"}, Configs: []string{"remix.config."}},
	{ID: "svelte", Name: "Svelte", Dependencies: []string{"svelte"}, Configs: []string{"svelte.config."}},
	{ID: "sveltekit", Name: "SvelteKit", Dependencies: []string{"@sveltejs/kit"}, Configs: []string{"svelte.config."}},
	{ID: "vite", Name: "Vite", Dependencies: []string{"vite"}, Configs: []string{"vite.config."}},
	{ID: "vue", Name: "Vue", Dependencies: []string{"vue"}},
}

var packageManagerMarkers = []packageManagerMarker{
	{ID: "bun", Name: "Bun", Lockfiles: []string{"bun.lock", "bun.lockb"}},
	{ID: "npm", Name: "npm", Lockfiles: []string{"package-lock.json", "npm-shrinkwrap.json"}},
	{ID: "pnpm", Name: "pnpm", Lockfiles: []string{"pnpm-lock.yaml"}},
	{ID: "yarn", Name: "Yarn", Lockfiles: []string{"yarn.lock"}},
}

// NewProjectDetector creates a detector governed by the supplied file policy.
func NewProjectDetector(policy privacy.FilePolicy) ProjectDetector {
	return ProjectDetector{policy: policy}
}

// Discover inspects bounded project-root metadata without executing project code.
func (d ProjectDetector) Discover(ctx context.Context, root string) (model.ProjectSummary, error) {
	if err := ctx.Err(); err != nil {
		return model.ProjectSummary{}, err
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return model.ProjectSummary{}, fmt.Errorf("resolve project path: %w", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return model.ProjectSummary{}, fmt.Errorf("resolve project path: %w", err)
	}
	info, err := os.Stat(canonicalRoot)
	if err != nil {
		return model.ProjectSummary{}, fmt.Errorf("inspect project path: %w", err)
	}
	if !info.IsDir() {
		return model.ProjectSummary{}, fmt.Errorf("project path is not a directory: %s", canonicalRoot)
	}

	entries, truncated, err := readRootEntries(ctx, canonicalRoot, maxRootEntries)
	if err != nil {
		return model.ProjectSummary{}, err
	}
	files, err := rootFiles(ctx, entries)
	if err != nil {
		return model.ProjectSummary{}, err
	}

	summary := model.ProjectSummary{
		Root: canonicalRoot,
		Name: filepath.Base(canonicalRoot),
	}
	if truncated {
		summary.Warnings = append(summary.Warnings, model.Warning{
			Code:    "project.entry-limit-reached",
			Message: fmt.Sprintf("The project root contains more than %d entries. Discovery inspected a bounded subset.", maxRootEntries),
		})
	}

	relevantFiles, fileWarnings, err := d.relevantFiles(ctx, canonicalRoot, files)
	if err != nil {
		return model.ProjectSummary{}, err
	}
	summary.RelevantFiles = relevantFiles
	summary.Warnings = append(summary.Warnings, fileWarnings...)

	manifest, manifestAvailable, manifestWarnings := d.readPackageManifest(canonicalRoot, files)
	summary.Warnings = append(summary.Warnings, manifestWarnings...)
	if manifestAvailable && manifest.Name != "" {
		summary.Name = manifest.Name
	}

	dependencies := mergedDependencies(manifest)
	summary.Languages = detectLanguages(files, manifestAvailable, dependencies)
	summary.Frameworks = detectFrameworks(files, dependencies)
	summary.Runtimes = detectRuntimes(manifest, manifestAvailable)
	packageManagers, manifestPackageManager := detectPackageManagers(files, manifest)
	summary.PackageManagers = packageManagers
	summary.Workspaces = detectWorkspaces(files, manifest)

	lockfileManagerCount := 0
	var lockfileEvidence []model.Evidence
	for _, manager := range summary.PackageManagers {
		if len(manager.Lockfiles) > 0 {
			lockfileManagerCount++
			lockfileEvidence = append(lockfileEvidence, manager.Evidence...)
		}
	}
	if lockfileManagerCount > 1 {
		summary.Warnings = append(summary.Warnings, model.Warning{
			Code:     "package-manager.conflicting-lockfiles",
			Message:  "Lockfiles from more than one package manager were found. DevDoctor will not guess which one should be used.",
			Evidence: onlyFileEvidence(lockfileEvidence),
		})
	}
	if manifestPackageManager != "" && !containsPackageManager(summary.PackageManagers, manifestPackageManager) {
		summary.Warnings = append(summary.Warnings, model.Warning{
			Code:    "package-manager.unknown-declaration",
			Message: fmt.Sprintf("package.json declares an unrecognized package manager %q.", manifestPackageManager),
			Evidence: []model.Evidence{{
				Kind:   "manifest-field",
				Path:   "package.json",
				Detail: "packageManager = " + manifest.PackageManager,
			}},
		})
	}

	if len(summary.Languages) == 0 && len(summary.Runtimes) == 0 && len(summary.Frameworks) == 0 {
		summary.Warnings = append(summary.Warnings, model.Warning{
			Code:    "project.no-supported-markers",
			Message: "No currently supported project markers were found at this directory level.",
		})
	}

	sortSummary(&summary)
	return summary, nil
}

type entryMaxHeap []os.DirEntry

func (entries entryMaxHeap) Len() int { return len(entries) }

func (entries entryMaxHeap) Less(i, j int) bool {
	return entryComesBefore(entries[j], entries[i])
}

func (entries entryMaxHeap) Swap(i, j int) { entries[i], entries[j] = entries[j], entries[i] }

func (entries *entryMaxHeap) Push(value any) {
	*entries = append(*entries, value.(os.DirEntry))
}

func (entries *entryMaxHeap) Pop() any {
	previous := *entries
	last := len(previous) - 1
	value := previous[last]
	*entries = previous[:last]
	return value
}

func readRootEntries(ctx context.Context, root string, limit int) ([]os.DirEntry, bool, error) {
	directory, err := os.Open(root)
	if err != nil {
		return nil, false, fmt.Errorf("open project directory: %w", err)
	}
	defer func() { _ = directory.Close() }()

	if limit <= 0 {
		return nil, false, fmt.Errorf("root entry limit must be positive")
	}

	entries := &entryMaxHeap{}
	heap.Init(entries)
	totalEntries := 0
	for {
		batch, readErr := directory.ReadDir(256)
		totalEntries += len(batch)
		for _, entry := range batch {
			if entry.IsDir() {
				continue
			}
			if entries.Len() < limit {
				heap.Push(entries, entry)
				continue
			}
			if entryComesBefore(entry, (*entries)[0]) {
				heap.Pop(entries)
				heap.Push(entries, entry)
			}
		}
		if err := ctx.Err(); err != nil {
			return nil, false, err
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return nil, false, fmt.Errorf("read project directory: %w", readErr)
		}
	}

	result := append([]os.DirEntry(nil), (*entries)...)
	sort.Slice(result, func(i, j int) bool { return result[i].Name() < result[j].Name() })
	return result, totalEntries > limit, nil
}

func entryComesBefore(left, right os.DirEntry) bool {
	leftRelevant := isRelevantFilename(left.Name())
	rightRelevant := isRelevantFilename(right.Name())
	if leftRelevant != rightRelevant {
		return leftRelevant
	}
	return left.Name() < right.Name()
}

func rootFiles(ctx context.Context, entries []os.DirEntry) (map[string]string, error) {
	files := make(map[string]string)
	for index, entry := range entries {
		if index%256 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		if entry.IsDir() {
			continue
		}
		files[normalizedFileKey(entry.Name())] = entry.Name()
	}
	return files, nil
}

func normalizedFileKey(name string) string {
	if runtime.GOOS == "windows" {
		return strings.ToLower(name)
	}
	return name
}

func (d ProjectDetector) relevantFiles(ctx context.Context, root string, files map[string]string) ([]string, []model.Warning, error) {
	var relevant []string
	var warnings []model.Warning
	names := make([]string, 0, len(files))
	for _, actualName := range files {
		names = append(names, actualName)
	}
	sort.Strings(names)
	for index, actualName := range names {
		if index%256 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
		}
		if !isRelevantFilename(actualName) {
			continue
		}
		if _, err := d.policy.ResolveMetadataPath(root, actualName); err != nil {
			warnings = append(warnings, model.Warning{
				Code:    "privacy.metadata-file-denied",
				Message: fmt.Sprintf("Skipped metadata file %q because it did not pass the safe-read policy.", actualName),
				Evidence: []model.Evidence{{
					Kind:   "file-policy",
					Path:   actualName,
					Detail: err.Error(),
				}},
			})
			continue
		}
		relevant = append(relevant, actualName)
	}
	sort.Strings(relevant)
	return relevant, warnings, nil
}

func (d ProjectDetector) readPackageManifest(root string, files map[string]string) (packageManifest, bool, []model.Warning) {
	actualName, ok := files["package.json"]
	if !ok {
		return packageManifest{}, false, nil
	}

	data, err := d.policy.ReadMetadata(root, actualName)
	if err != nil {
		return packageManifest{}, true, []model.Warning{{
			Code:     "manifest.package-json-unreadable",
			Message:  "package.json was found but could not be read safely.",
			Evidence: []model.Evidence{{Kind: "file-read", Path: actualName, Detail: err.Error()}},
		}}
	}

	var manifest packageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return packageManifest{}, true, []model.Warning{{
			Code:     "manifest.package-json-invalid",
			Message:  "package.json is not valid JSON.",
			Evidence: []model.Evidence{{Kind: "parse-error", Path: actualName, Detail: err.Error()}},
		}}
	}
	return manifest, true, nil
}

func mergedDependencies(manifest packageManifest) map[string]string {
	dependencies := make(map[string]string)
	for _, source := range []map[string]string{
		manifest.Dependencies,
		manifest.DevDependencies,
		manifest.PeerDependencies,
		manifest.OptionalDependencies,
	} {
		for name, version := range source {
			dependencies[name] = version
		}
	}
	return dependencies
}

func detectLanguages(files map[string]string, hasManifest bool, dependencies map[string]string) []model.Detection {
	var languages []model.Detection
	var jsEvidence []model.Evidence
	var tsEvidence []model.Evidence

	if hasManifest {
		jsEvidence = append(jsEvidence, model.Evidence{Kind: "manifest", Path: "package.json", Detail: "Node package manifest"})
	}
	if actual, ok := files["jsconfig.json"]; ok {
		jsEvidence = append(jsEvidence, model.Evidence{Kind: "config-file", Path: actual, Detail: "JavaScript project configuration"})
	}
	if actual, ok := files["tsconfig.json"]; ok {
		tsEvidence = append(tsEvidence, model.Evidence{Kind: "config-file", Path: actual, Detail: "TypeScript project configuration"})
	}
	if version, ok := dependencies["typescript"]; ok {
		tsEvidence = append(tsEvidence, model.Evidence{Kind: "dependency", Path: "package.json", Detail: "typescript " + version})
	}

	for lowerName, actualName := range files {
		extension := strings.ToLower(filepath.Ext(lowerName))
		switch extension {
		case ".js", ".jsx", ".mjs", ".cjs":
			jsEvidence = append(jsEvidence, model.Evidence{Kind: "source-marker", Path: actualName, Detail: "JavaScript source file at project root"})
		case ".ts", ".tsx", ".mts", ".cts":
			tsEvidence = append(tsEvidence, model.Evidence{Kind: "source-marker", Path: actualName, Detail: "TypeScript source file at project root"})
		}
	}

	if len(jsEvidence) > 0 {
		languages = append(languages, model.Detection{ID: "javascript", Name: "JavaScript", Confidence: model.ConfidenceHigh, Evidence: jsEvidence})
	}
	if len(tsEvidence) > 0 {
		languages = append(languages, model.Detection{ID: "typescript", Name: "TypeScript", Confidence: model.ConfidenceHigh, Evidence: tsEvidence})
	}
	return languages
}

func detectFrameworks(files map[string]string, dependencies map[string]string) []model.Detection {
	var detections []model.Detection
	for _, marker := range frameworkMarkers {
		var evidence []model.Evidence
		confidence := model.ConfidenceMedium
		for _, dependency := range marker.Dependencies {
			if requirement, ok := dependencies[dependency]; ok {
				evidence = append(evidence, model.Evidence{Kind: "dependency", Path: "package.json", Detail: dependency + " " + requirement})
				confidence = model.ConfidenceHigh
			}
		}
		for _, config := range marker.Configs {
			for lowerName, actualName := range files {
				if lowerName == config || strings.HasPrefix(lowerName, config) {
					evidence = append(evidence, model.Evidence{Kind: "config-file", Path: actualName, Detail: marker.Name + " configuration marker"})
				}
			}
		}
		if len(evidence) > 0 {
			detections = append(detections, model.Detection{
				ID: marker.ID, Name: marker.Name, Confidence: confidence, Evidence: evidence,
			})
		}
	}
	return detections
}

func detectRuntimes(manifest packageManifest, hasManifest bool) []model.RuntimeDetection {
	if !hasManifest {
		return nil
	}

	runtime := model.RuntimeDetection{
		ID:         "nodejs",
		Name:       "Node.js",
		Confidence: model.ConfidenceHigh,
		Evidence: []model.Evidence{{
			Kind: "manifest", Path: "package.json", Detail: "Node package manifest",
		}},
	}
	if requirement := manifest.Engines["node"]; requirement != "" {
		runtime.Requirement = requirement
		runtime.Evidence = append(runtime.Evidence, model.Evidence{
			Kind: "manifest-field", Path: "package.json", Detail: "engines.node = " + requirement,
		})
	}
	return []model.RuntimeDetection{runtime}
}

func detectPackageManagers(files map[string]string, manifest packageManifest) ([]model.PackageManagerDetection, string) {
	detections := make(map[string]*model.PackageManagerDetection)
	declaredName, declaredVersion := parsePackageManager(manifest.PackageManager)

	for _, marker := range packageManagerMarkers {
		for _, lockfile := range marker.Lockfiles {
			actualName, ok := files[strings.ToLower(lockfile)]
			if !ok {
				continue
			}
			detection := ensurePackageManager(detections, marker.ID, marker.Name)
			detection.Lockfiles = append(detection.Lockfiles, actualName)
			detection.Evidence = append(detection.Evidence, model.Evidence{
				Kind: "lockfile", Path: actualName, Detail: marker.Name + " lockfile",
			})
		}
	}

	if declaredName != "" {
		for _, marker := range packageManagerMarkers {
			if declaredName != marker.ID {
				continue
			}
			detection := ensurePackageManager(detections, marker.ID, marker.Name)
			detection.DeclaredVersion = declaredVersion
			detection.Evidence = append(detection.Evidence, model.Evidence{
				Kind: "manifest-field", Path: "package.json", Detail: "packageManager = " + manifest.PackageManager,
			})
			break
		}
	}

	if actualName, ok := files["pnpm-workspace.yaml"]; ok {
		detection := ensurePackageManager(detections, "pnpm", "pnpm")
		detection.Evidence = append(detection.Evidence, model.Evidence{
			Kind: "workspace-file", Path: actualName, Detail: "pnpm workspace configuration",
		})
	}

	result := make([]model.PackageManagerDetection, 0, len(detections))
	for _, detection := range detections {
		if len(detection.Lockfiles) > 0 || detection.DeclaredVersion != "" {
			detection.Confidence = model.ConfidenceHigh
		} else {
			detection.Confidence = model.ConfidenceMedium
		}
		result = append(result, *detection)
	}
	return result, declaredName
}

func ensurePackageManager(detections map[string]*model.PackageManagerDetection, id, name string) *model.PackageManagerDetection {
	if detection, ok := detections[id]; ok {
		return detection
	}
	detection := &model.PackageManagerDetection{ID: id, Name: name}
	detections[id] = detection
	return detection
}

func parsePackageManager(value string) (string, string) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "", ""
	}
	separator := strings.Index(value, "@")
	if separator < 0 {
		return value, ""
	}
	return value[:separator], value[separator+1:]
}

func detectWorkspaces(files map[string]string, manifest packageManifest) []model.WorkspaceDetection {
	var workspaces []model.WorkspaceDetection
	if patterns := parseWorkspacePatterns(manifest.Workspaces); len(patterns) > 0 {
		workspaces = append(workspaces, model.WorkspaceDetection{
			ID:       "package-workspaces",
			Name:     "Package workspaces",
			Patterns: patterns,
			Evidence: []model.Evidence{{Kind: "manifest-field", Path: "package.json", Detail: "workspaces field"}},
		})
	}
	if actualName, ok := files["pnpm-workspace.yaml"]; ok {
		workspaces = append(workspaces, model.WorkspaceDetection{
			ID:   "pnpm-workspace",
			Name: "pnpm workspace",
			Evidence: []model.Evidence{{
				Kind: "workspace-file", Path: actualName, Detail: "pnpm workspace configuration",
			}},
		})
	}
	return workspaces
}

func parseWorkspacePatterns(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var patterns []string
	if err := json.Unmarshal(raw, &patterns); err == nil {
		sort.Strings(patterns)
		return patterns
	}
	var object struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(raw, &object); err == nil {
		sort.Strings(object.Packages)
		return object.Packages
	}
	return nil
}

func containsPackageManager(detections []model.PackageManagerDetection, id string) bool {
	for _, detection := range detections {
		if detection.ID == id {
			return true
		}
	}
	return false
}

func onlyFileEvidence(evidence []model.Evidence) []model.Evidence {
	var result []model.Evidence
	for _, item := range evidence {
		if item.Kind == "lockfile" {
			result = append(result, item)
		}
	}
	return result
}

func isRelevantFilename(name string) bool {
	normalizedName := name
	if runtime.GOOS == "windows" {
		normalizedName = strings.ToLower(name)
	}
	knownNames := []string{
		"package.json", "tsconfig.json", "jsconfig.json", "package-lock.json", "npm-shrinkwrap.json",
		"pnpm-lock.yaml", "pnpm-workspace.yaml", "yarn.lock", "bun.lock", "bun.lockb", "Dockerfile", "dockerfile",
		"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml", ".env.example",
		"angular.json", "nest-cli.json",
	}
	for _, knownName := range knownNames {
		if normalizedName == knownName {
			return true
		}
	}
	for _, prefix := range []string{
		"next.config.", "nuxt.config.", "vite.config.", "svelte.config.", "astro.config.",
		"remix.config.", "webpack.config.",
	} {
		if strings.HasPrefix(normalizedName, prefix) {
			return true
		}
	}
	return false
}

func sortSummary(summary *model.ProjectSummary) {
	normalizeSummarySlices(summary)
	for i := range summary.Languages {
		sortEvidence(summary.Languages[i].Evidence)
	}
	for i := range summary.Frameworks {
		sortEvidence(summary.Frameworks[i].Evidence)
	}
	for i := range summary.Runtimes {
		sortEvidence(summary.Runtimes[i].Evidence)
	}
	for i := range summary.PackageManagers {
		sort.Strings(summary.PackageManagers[i].Lockfiles)
		sortEvidence(summary.PackageManagers[i].Evidence)
	}
	for i := range summary.Workspaces {
		sort.Strings(summary.Workspaces[i].Patterns)
		sortEvidence(summary.Workspaces[i].Evidence)
	}
	for i := range summary.Warnings {
		sortEvidence(summary.Warnings[i].Evidence)
	}

	sort.Slice(summary.Languages, func(i, j int) bool { return summary.Languages[i].ID < summary.Languages[j].ID })
	sort.Slice(summary.Frameworks, func(i, j int) bool { return summary.Frameworks[i].ID < summary.Frameworks[j].ID })
	sort.Slice(summary.Runtimes, func(i, j int) bool { return summary.Runtimes[i].ID < summary.Runtimes[j].ID })
	sort.Slice(summary.PackageManagers, func(i, j int) bool { return summary.PackageManagers[i].ID < summary.PackageManagers[j].ID })
	sort.Slice(summary.Workspaces, func(i, j int) bool { return summary.Workspaces[i].ID < summary.Workspaces[j].ID })
	sort.Slice(summary.Warnings, func(i, j int) bool {
		left := warningSortKey(summary.Warnings[i])
		right := warningSortKey(summary.Warnings[j])
		return left < right
	})
}

func normalizeSummarySlices(summary *model.ProjectSummary) {
	if summary.Languages == nil {
		summary.Languages = []model.Detection{}
	}
	if summary.Frameworks == nil {
		summary.Frameworks = []model.Detection{}
	}
	if summary.Runtimes == nil {
		summary.Runtimes = []model.RuntimeDetection{}
	}
	if summary.PackageManagers == nil {
		summary.PackageManagers = []model.PackageManagerDetection{}
	}
	if summary.Workspaces == nil {
		summary.Workspaces = []model.WorkspaceDetection{}
	}
	if summary.RelevantFiles == nil {
		summary.RelevantFiles = []string{}
	}
	if summary.Warnings == nil {
		summary.Warnings = []model.Warning{}
	}
	for index := range summary.Languages {
		if summary.Languages[index].Evidence == nil {
			summary.Languages[index].Evidence = []model.Evidence{}
		}
	}
	for index := range summary.Frameworks {
		if summary.Frameworks[index].Evidence == nil {
			summary.Frameworks[index].Evidence = []model.Evidence{}
		}
	}
	for index := range summary.Runtimes {
		if summary.Runtimes[index].Evidence == nil {
			summary.Runtimes[index].Evidence = []model.Evidence{}
		}
	}
	for index := range summary.PackageManagers {
		if summary.PackageManagers[index].Lockfiles == nil {
			summary.PackageManagers[index].Lockfiles = []string{}
		}
		if summary.PackageManagers[index].Evidence == nil {
			summary.PackageManagers[index].Evidence = []model.Evidence{}
		}
	}
	for index := range summary.Workspaces {
		if summary.Workspaces[index].Patterns == nil {
			summary.Workspaces[index].Patterns = []string{}
		}
		if summary.Workspaces[index].Evidence == nil {
			summary.Workspaces[index].Evidence = []model.Evidence{}
		}
	}
}

func warningSortKey(warning model.Warning) string {
	var key strings.Builder
	key.WriteString(warning.Code)
	key.WriteByte(0)
	key.WriteString(warning.Message)
	for _, evidence := range warning.Evidence {
		key.WriteByte(0)
		key.WriteString(evidence.Path)
		key.WriteByte(0)
		key.WriteString(evidence.Kind)
		key.WriteByte(0)
		key.WriteString(evidence.Detail)
	}
	return key.String()
}

func sortEvidence(evidence []model.Evidence) {
	sort.Slice(evidence, func(i, j int) bool {
		if evidence[i].Path == evidence[j].Path {
			if evidence[i].Kind == evidence[j].Kind {
				return evidence[i].Detail < evidence[j].Detail
			}
			return evidence[i].Kind < evidence[j].Kind
		}
		return evidence[i].Path < evidence[j].Path
	})
}
