package privacy

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DefaultMaxMetadataBytes bounds each metadata read to one mebibyte.
const DefaultMaxMetadataBytes int64 = 1 << 20

// ErrFileDenied indicates that a metadata path violated the discovery policy.
var ErrFileDenied = errors.New("file is denied by the discovery policy")

// FilePolicy controls which metadata files discovery may read and how much data it accepts.
type FilePolicy struct {
	MaxBytes int64
}

// NewFilePolicy returns the default local-only metadata policy.
func NewFilePolicy() FilePolicy {
	return FilePolicy{MaxBytes: DefaultMaxMetadataBytes}
}

// ReadMetadata safely reads one allowlisted metadata file beneath root.
func (p FilePolicy) ReadMetadata(root, relativePath string) ([]byte, error) {
	resolved, err := p.ResolveMetadataPath(root, relativePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(resolved)
	if err != nil {
		return nil, fmt.Errorf("open metadata file: %w", err)
	}
	defer func() { _ = file.Close() }()

	openedInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect opened metadata file: %w", err)
	}
	pathInfo, err := os.Lstat(resolved)
	if err != nil {
		return nil, fmt.Errorf("revalidate metadata path: %w", err)
	}
	if pathInfo.Mode()&os.ModeSymlink != 0 || !os.SameFile(openedInfo, pathInfo) {
		return nil, fmt.Errorf("%w: metadata path changed or became a symbolic link", ErrFileDenied)
	}
	if !openedInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: metadata path is not a regular file", ErrFileDenied)
	}

	limit := p.MaxBytes
	if limit <= 0 {
		limit = DefaultMaxMetadataBytes
	}
	if openedInfo.Size() > limit {
		return nil, fmt.Errorf("%w: metadata file exceeds %d bytes", ErrFileDenied, limit)
	}

	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read metadata file: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("%w: metadata file exceeds %d bytes", ErrFileDenied, limit)
	}

	return data, nil
}

// ResolveMetadataPath validates and resolves an allowlisted direct child of root.
func (p FilePolicy) ResolveMetadataPath(root, relativePath string) (string, error) {
	if filepath.IsAbs(relativePath) || relativePath == "" {
		return "", fmt.Errorf("%w: expected a non-empty relative path", ErrFileDenied)
	}
	if !isAllowedMetadataName(filepath.ToSlash(relativePath)) {
		return "", fmt.Errorf("%w: %s is not allowlisted", ErrFileDenied, relativePath)
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return "", fmt.Errorf("resolve project root symlinks: %w", err)
	}

	candidate := filepath.Join(canonicalRoot, filepath.FromSlash(relativePath))
	if !isWithinRoot(canonicalRoot, candidate) {
		return "", fmt.Errorf("%w: metadata path escapes the project root", ErrFileDenied)
	}
	info, err := os.Lstat(candidate)
	if err != nil {
		return "", fmt.Errorf("inspect metadata path: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("%w: symbolic links are not read during discovery", ErrFileDenied)
	}

	return candidate, nil
}

func isAllowedMetadataName(relativePath string) bool {
	clean := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(relativePath)), "./")
	if strings.Contains(clean, "/") {
		return false
	}
	if runtime.GOOS == "windows" {
		clean = strings.ToLower(clean)
	}

	switch clean {
	case "package.json", "tsconfig.json", "jsconfig.json", "package-lock.json", "npm-shrinkwrap.json",
		"pnpm-lock.yaml", "pnpm-workspace.yaml", "yarn.lock", "bun.lock", "bun.lockb",
		"Dockerfile", "dockerfile", "compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml",
		".env.example":
		return true
	default:
		return isAllowedConfigName(clean)
	}
}

func isAllowedConfigName(name string) bool {
	prefixes := []string{
		"next.config.", "nuxt.config.", "vite.config.", "svelte.config.", "astro.config.",
		"angular.json", "nest-cli.json", "remix.config.", "webpack.config.",
	}
	for _, prefix := range prefixes {
		if name == prefix || strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func isWithinRoot(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}
