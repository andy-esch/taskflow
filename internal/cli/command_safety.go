package cli

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/andy-esch/taskflow/internal/wire"
)

const (
	commandSafetyAnnotation = "safety"
	commandSafetyReadOnly   = "read-only"
	commandSafetyMutating   = "mutating"
)

// commandSafetyState is the invocation-scoped bridge between Cobra's primary
// adapter metadata and secondary adapters that can persist data. Stores know
// only that a mutation must be authorized; they do not depend on Cobra or on
// the read-only/mutating vocabulary.
type commandSafetyState struct {
	path   string
	safety string
}

func (s *commandSafetyState) bind(cmd *cobra.Command) error {
	if cmd == nil {
		return fmt.Errorf("command safety invariant: selected command is nil")
	}
	safety := cmd.Annotations[commandSafetyAnnotation]
	// Cobra constructs its private __complete transport only after Execute has
	// begun, so it cannot be annotated alongside the stable registered tree.
	// It is nevertheless runnable and reaches our persistent pre-run; bind its
	// fixed read-only capability here and keep ordinary commands fail-closed.
	if safety == "" && isCompletionCommand(cmd) {
		safety = commandSafetyReadOnly
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[commandSafetyAnnotation] = safety
	}
	if !recognizedCommandSafety(safety) {
		return fmt.Errorf("command safety invariant: runnable command %q has unrecognized %s annotation %q",
			cmd.CommandPath(), commandSafetyAnnotation, safety)
	}
	s.path = cmd.CommandPath()
	s.safety = safety
	return nil
}

func (s *commandSafetyState) authorizeMutation() error {
	if s.safety == commandSafetyMutating {
		return nil
	}
	if s.safety == commandSafetyReadOnly {
		return fmt.Errorf("command safety violation: read-only command %q reached a mutating persistence path", s.path)
	}
	return fmt.Errorf("command safety violation: mutation reached persistence before a command safety classification was bound")
}

func recognizedCommandSafety(safety string) bool {
	return safety == commandSafetyReadOnly || safety == commandSafetyMutating
}

// commandSafetySurface returns every executable command, including hidden and
// deprecated compatibility leaves, in byte-stable command-path order. Grouping
// nodes are deliberately absent: they have no Run/RunE and cannot produce an
// operation whose side effects need classifying.
func commandSafetySurface(root *cobra.Command) ([]wire.SchemaCommand, error) {
	commands := make([]wire.SchemaCommand, 0)
	var visit func(*cobra.Command) error
	visit = func(cmd *cobra.Command) error {
		if cmd.Runnable() {
			safety := cmd.Annotations[commandSafetyAnnotation]
			if !recognizedCommandSafety(safety) {
				return fmt.Errorf("command safety invariant: runnable command %q has unrecognized %s annotation %q",
					cmd.CommandPath(), commandSafetyAnnotation, safety)
			}
			commands = append(commands, wire.SchemaCommand{
				Path: cmd.CommandPath(), Safety: safety,
				Hidden: cmd.Hidden, Deprecated: cmd.Deprecated != "",
			})
		}
		for _, child := range cmd.Commands() {
			if err := visit(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(root); err != nil {
		return nil, err
	}
	slices.SortFunc(commands, func(a, b wire.SchemaCommand) int {
		if a.Path < b.Path {
			return -1
		}
		if a.Path > b.Path {
			return 1
		}
		return 0
	})
	return commands, nil
}

func markCommandTreeReadOnly(cmd *cobra.Command) {
	if cmd == nil {
		return
	}
	if cmd.Runnable() {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[commandSafetyAnnotation] = commandSafetyReadOnly
	}
	for _, child := range cmd.Commands() {
		markCommandTreeReadOnly(child)
	}
}
