package tui

import (
	"fmt"
	"sort"
	"strings"
)

// CommandID is a closed identifier for a registered DevDoctor action.
type CommandID string

// Registered command identifiers.
const (
	CommandDiagnose CommandID = "diagnose"
	CommandProject  CommandID = "project"
	CommandWarnings CommandID = "warnings"
	CommandExport   CommandID = "export"
	CommandHelp     CommandID = "help"
	CommandClear    CommandID = "clear"
	CommandQuit     CommandID = "quit"
)

// Command describes an allowlisted composer action.
type Command struct {
	ID          CommandID
	Name        string
	Description string
	Rank        int
	Available   func(CommandContext) bool
}

// CommandContext describes the state used to expose safe actions.
type CommandContext struct {
	State     RunState
	HasReport bool
}

// Action is the typed result of parsing composer input.
type Action struct {
	ID CommandID
}

var commandRegistry = []Command{
	{ID: CommandDiagnose, Name: "/diagnose", Description: "Inspect project metadata", Rank: 0, Available: availableWhenIdle},
	{ID: CommandProject, Name: "/project", Description: "View project details", Rank: 1, Available: availableWhenIdle},
	{ID: CommandWarnings, Name: "/warnings", Description: "Review warnings", Rank: 2, Available: availableWhenIdle},
	{ID: CommandExport, Name: "/export", Description: "Preview report", Rank: 3, Available: availableWhenIdle},
	{ID: CommandHelp, Name: "/help", Description: "Show help", Rank: 4, Available: availableWhenIdle},
	{ID: CommandClear, Name: "/clear", Description: "Clear current view", Rank: 5, Available: availableWhenIdle},
	{ID: CommandQuit, Name: "/quit", Description: "Exit DevDoctor", Rank: 6, Available: availableAlways},
}

func availableWhenIdle(context CommandContext) bool {
	return context.State != StateRunning && context.State != StateWaiting
}

func availableAlways(CommandContext) bool {
	return true
}

// Commands returns a copy of the registered command catalog.
func Commands() []Command {
	commands := make([]Command, len(commandRegistry))
	copy(commands, commandRegistry)
	return commands
}

// FilterCommands returns all registered commands matching a composer draft.
func FilterCommands(draft string) []Command {
	return filterCommands(draft, CommandContext{}, false)
}

// FilterAvailableCommands returns matching commands allowed in the current state.
func FilterAvailableCommands(draft string, context CommandContext) []Command {
	return filterCommands(draft, context, true)
}

func filterCommands(draft string, context CommandContext, availableOnly bool) []Command {
	query := strings.ToLower(strings.TrimSpace(draft))
	query = strings.TrimPrefix(query, "/")
	matches := make([]Command, 0, len(commandRegistry))
	for _, command := range commandRegistry {
		if availableOnly && command.Available != nil && !command.Available(context) {
			continue
		}
		name := strings.TrimPrefix(strings.ToLower(command.Name), "/")
		description := strings.ToLower(command.Description)
		if query == "" || strings.Contains(name, query) || strings.Contains(description, query) {
			matches = append(matches, command)
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		leftScore := commandMatchScore(matches[i], query)
		rightScore := commandMatchScore(matches[j], query)
		if leftScore != rightScore {
			return leftScore < rightScore
		}
		if matches[i].Rank != matches[j].Rank {
			return matches[i].Rank < matches[j].Rank
		}
		return matches[i].Name < matches[j].Name
	})
	return matches
}

func commandMatchScore(command Command, query string) int {
	if query == "" {
		return 0
	}
	name := strings.TrimPrefix(strings.ToLower(command.Name), "/")
	description := strings.ToLower(command.Description)
	switch {
	case strings.HasPrefix(name, query):
		return 0
	case strings.Contains(name, query):
		return 1
	case strings.Contains(description, query):
		return 2
	default:
		return 3
	}
}

// IsActionAvailable reports whether an action is valid in the current state.
func IsActionAvailable(action Action, context CommandContext) bool {
	for _, command := range commandRegistry {
		if command.ID == action.ID {
			return command.Available == nil || command.Available(context)
		}
	}
	return false
}

// ParseAction accepts only a complete registered command with no arguments.
func ParseAction(input string) (Action, error) {
	fields := strings.Fields(strings.TrimSpace(input))
	if len(fields) == 0 {
		return Action{}, fmt.Errorf("enter a registered DevDoctor command beginning with /")
	}
	if !strings.HasPrefix(fields[0], "/") {
		return Action{}, fmt.Errorf("DevDoctor accepts registered slash commands only; shell input was not run")
	}
	if len(fields) != 1 {
		return Action{}, fmt.Errorf("%s does not accept arguments", fields[0])
	}
	for _, command := range commandRegistry {
		if strings.EqualFold(fields[0], command.Name) {
			return Action{ID: command.ID}, nil
		}
	}
	return Action{}, fmt.Errorf("no matching DevDoctor command for %q", fields[0])
}
