package runner

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/mahmudddddd/DevDoctor/internal/model"
)

var secretNameFragments = []string{
	"API_KEY",
	"AUTH",
	"COOKIE",
	"CREDENTIAL",
	"PASSWD",
	"PASSWORD",
	"PRIVATE_KEY",
	"SECRET",
	"TOKEN",
}

// EnvironmentNames returns the names that would be passed without exposing values.
func EnvironmentNames(spec model.EnvironmentSpec) ([]string, error) {
	explicit, err := validateExplicitEnvironment(spec)
	if err != nil {
		return nil, err
	}

	names := append([]string(nil), baselineEnvironmentNames()...)
	for _, item := range explicit {
		names = append(names, item.name)
	}
	names = uniqueEnvironmentNames(names)
	sort.Slice(names, func(i, j int) bool {
		return environmentKey(names[i]) < environmentKey(names[j])
	})
	return names, nil
}

func buildEnvironment(spec model.EnvironmentSpec) ([]string, error) {
	explicit, err := validateExplicitEnvironment(spec)
	if err != nil {
		return nil, err
	}

	values := make(map[string]environmentValue)
	for _, name := range baselineEnvironmentNames() {
		if value, ok := lookupEnvironment(name); ok {
			values[environmentKey(name)] = environmentValue{name: name, value: value}
		}
	}
	for _, item := range explicit {
		if item.set {
			values[environmentKey(item.name)] = environmentValue{name: item.name, value: item.value}
			continue
		}
		if value, ok := lookupEnvironment(item.name); ok {
			values[environmentKey(item.name)] = environmentValue{name: item.name, value: value}
		}
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	environment := make([]string, 0, len(keys))
	for _, key := range keys {
		item := values[key]
		environment = append(environment, item.name+"="+item.value)
	}
	return environment, nil
}

type explicitEnvironment struct {
	name  string
	value string
	set   bool
}

type environmentValue struct {
	name  string
	value string
}

func validateExplicitEnvironment(spec model.EnvironmentSpec) ([]explicitEnvironment, error) {
	seen := make(map[string]struct{}, len(spec.Pass)+len(spec.Set))
	items := make([]explicitEnvironment, 0, len(spec.Pass)+len(spec.Set))

	for _, name := range spec.Pass {
		if err := validateEnvironmentName(name); err != nil {
			return nil, err
		}
		key := environmentKey(name)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("environment variable %q is declared more than once", name)
		}
		seen[key] = struct{}{}
		items = append(items, explicitEnvironment{name: name})
	}
	for name, value := range spec.Set {
		if err := validateEnvironmentName(name); err != nil {
			return nil, err
		}
		key := environmentKey(name)
		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf("environment variable %q is declared more than once", name)
		}
		seen[key] = struct{}{}
		items = append(items, explicitEnvironment{name: name, value: value, set: true})
	}
	return items, nil
}

func validateEnvironmentName(name string) error {
	if name == "" {
		return fmt.Errorf("environment variable name is empty")
	}
	for index, character := range name {
		if index == 0 {
			if character != '_' && !unicode.IsLetter(character) {
				return fmt.Errorf("environment variable name %q is malformed", name)
			}
			continue
		}
		if character != '_' && !unicode.IsLetter(character) && !unicode.IsDigit(character) {
			return fmt.Errorf("environment variable name %q is malformed", name)
		}
	}
	upper := strings.ToUpper(name)
	for _, fragment := range secretNameFragments {
		if strings.Contains(upper, fragment) {
			return fmt.Errorf("environment variable %q is secret-bearing and is not permitted", name)
		}
	}
	return nil
}

func uniqueEnvironmentNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	unique := make([]string, 0, len(names))
	for _, name := range names {
		key := environmentKey(name)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, name)
	}
	return unique
}

func lookupEnvironment(name string) (string, bool) {
	if !environmentCaseInsensitive() {
		return os.LookupEnv(name)
	}
	key := environmentKey(name)
	for _, entry := range os.Environ() {
		currentName, value, ok := strings.Cut(entry, "=")
		if ok && environmentKey(currentName) == key {
			return value, true
		}
	}
	return "", false
}
