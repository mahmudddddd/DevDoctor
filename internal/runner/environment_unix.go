//go:build !windows

package runner

func baselineEnvironmentNames() []string {
	return []string{"HOME", "LANG", "LC_ALL", "PATH", "TMPDIR"}
}

func environmentCaseInsensitive() bool {
	return false
}

func environmentKey(name string) string {
	return name
}
