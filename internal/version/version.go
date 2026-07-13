package version

import "fmt"

// Build metadata is replaced by release linker flags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a concise human-readable build identifier.
func String() string {
	if Version == "dev" {
		return Version
	}

	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
}
