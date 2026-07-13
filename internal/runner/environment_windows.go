//go:build windows

package runner

import "strings"

func baselineEnvironmentNames() []string {
	return []string{
		"COMSPEC",
		"HOMEDRIVE",
		"HOMEPATH",
		"LOCALAPPDATA",
		"PATH",
		"PATHEXT",
		"SystemRoot",
		"TEMP",
		"TMP",
		"USERPROFILE",
		"WINDIR",
	}
}

func environmentCaseInsensitive() bool {
	return true
}

func environmentKey(name string) string {
	return strings.ToUpper(name)
}
