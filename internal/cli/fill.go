package cli

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/cli/prompt"
	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// needPrompt resolves the gate half of the flag-twin contract, shared by every
// fillX helper so the rule lives in ONE place no matter the value type: a value
// already present is kept; a missing value with the gate closed is the
// required-input error (exit 11); a missing value with the gate open means "go
// prompt" (done=false). `have` is the caller's already-set test (`v != ""`, or
// `len(tags) > 0` for a future multiselect), kept out of here so single- and
// multi-value fills can both reuse the skeleton.
func (a *App) needPrompt(have bool, requiredMsg string) (done bool, err error) {
	if have {
		return true, nil
	}
	if !a.Gate.On() {
		return true, fmt.Errorf("%w: %s", domain.ErrValidation, requiredMsg)
	}
	return false, nil
}

// fillSelect resolves a required single-value param: flag value → prompt on a TTY
// → exit 11 otherwise. optsFn is called only on the prompt path; an empty option
// set fails with emptyMsg rather than opening a dead-end picker (clig.dev: a
// prompt that can't succeed should fail with guidance).
func (a *App) fillSelect(value, requiredMsg, emptyMsg, title string, optsFn func() ([]prompt.Option, error)) (string, error) {
	if done, err := a.needPrompt(value != "", requiredMsg); done {
		return value, err
	}
	opts, err := optsFn()
	if err != nil {
		return "", err
	}
	if len(opts) == 0 {
		return "", fmt.Errorf("%w: %s", domain.ErrValidation, emptyMsg)
	}
	return a.Prompt.SelectOne(title, opts)
}

// fillText resolves a required free-text param: flag value → text prompt on a TTY
// → exit 11 otherwise.
func (a *App) fillText(value, requiredMsg, title, placeholder string) (string, error) {
	if done, err := a.needPrompt(value != "", requiredMsg); done {
		return value, err
	}
	return a.Prompt.Text(title, placeholder)
}

// labeledOption builds a picker option from an id plus an optional description —
// the shared label shape for the epic, transition, and (future) tags lists, so
// the separator and empty-desc guard aren't copy-pasted per builder.
func labeledOption(id, desc string) prompt.Option {
	label := id
	if desc != "" {
		label += "  ·  " + desc
	}
	return prompt.Option{Label: label, Value: id}
}

// transitionOptions lists the active tasks eligible to move to `to` (everything
// not already there), as a picker source for a bare transition verb.
func (a *App) transitionOptions(to domain.Status) func() ([]prompt.Option, error) {
	return func() ([]prompt.Option, error) {
		tasks, _, err := a.Svc.ListTasks(core.TaskFilter{})
		if err != nil {
			return nil, err
		}
		opts := make([]prompt.Option, 0, len(tasks))
		for _, t := range tasks {
			if t.Status == to {
				continue
			}
			opts = append(opts, labeledOption(t.Slug, t.Description))
		}
		return opts, nil
	}
}

// epicOptions lists epics as pickable options (id + description), for the
// `task new` epic prompt. Read through the service like every other read.
func (a *App) epicOptions() ([]prompt.Option, error) {
	epics, _, err := a.Svc.ListEpics()
	if err != nil {
		return nil, err
	}
	opts := make([]prompt.Option, 0, len(epics))
	for _, e := range epics {
		opts = append(opts, labeledOption(e.Epic.ID, e.Epic.Description))
	}
	return opts, nil
}
