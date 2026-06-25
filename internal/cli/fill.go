package cli

import (
	"fmt"
	"sort"
	"strings"

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

// fillTags resolves the required tags. Tags are FREE-FORM (no fixed vocabulary),
// so the prompt is a comma-separated TEXT input — not a multiselect over a list
// that doesn't really exist — with the tags already in use offered as a
// placeholder hint. At least one tag is required. hintFn is called only on the
// prompt path (a richer multiselect-of-suggestions is a future enhancement).
func (a *App) fillTags(tags []string, hintFn func() string) ([]string, error) {
	if done, err := a.needPrompt(len(tags) > 0, "--tags is required (at least one)"); done {
		return tags, err
	}
	raw, err := a.Prompt.Text("Tags (comma-separated)", hintFn())
	if err != nil {
		return nil, err
	}
	parsed := parseTags(raw)
	if len(parsed) == 0 {
		return nil, fmt.Errorf("%w: at least one tag is required", domain.ErrValidation)
	}
	return parsed, nil
}

// parseTags splits a comma-separated string into trimmed, de-duplicated,
// non-empty tags, preserving first-seen order.
func parseTags(s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range strings.Split(s, ",") {
		t := strings.TrimSpace(p)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

// tagHint lists a few tags already in use as a placeholder suggestion for the
// free-form tags prompt. Falls back to a generic example when none are in use or
// the read fails (a hint must never block creating a task).
func (a *App) tagHint() string {
	tasks, _, err := a.Svc.ListTasks(core.TaskFilter{All: true})
	if err != nil {
		return "e.g. net, ui"
	}
	seen := map[string]bool{}
	var tags []string
	for _, t := range tasks {
		for _, tag := range t.Tags {
			if !seen[tag] {
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	if len(tags) == 0 {
		return "e.g. net, ui"
	}
	sort.Strings(tags)
	if len(tags) > 6 {
		tags = tags[:6]
	}
	return "in use: " + strings.Join(tags, ", ")
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

// resolveOne returns the single positional target, falling back to a picker on a
// TTY when none was given (exit 11 non-interactively) — the bare-invocation twin of
// fillSelect for the `<noun> show` / `task set` / `task append` commands, which take
// at most one arg.
func (a *App) resolveOne(args []string, requiredMsg, emptyMsg, title string, optsFn func() ([]prompt.Option, error)) (string, error) {
	value := ""
	if len(args) == 1 {
		value = args[0]
	}
	return a.fillSelect(value, requiredMsg, emptyMsg, title, optsFn)
}

// taskOptions lists the active tasks as a picker source for a bare `task
// edit/show/set/append`. An explicit slug resolves across every status regardless;
// the picker just offers the working set so a human doesn't have to remember a slug.
func (a *App) taskOptions() ([]prompt.Option, error) {
	tasks, _, err := a.Svc.ListTasks(core.TaskFilter{})
	if err != nil {
		return nil, err
	}
	opts := make([]prompt.Option, 0, len(tasks))
	for _, t := range tasks {
		opts = append(opts, labeledOption(t.Slug, t.Description))
	}
	return opts, nil
}

// auditOptions lists every audit (all buckets) as a picker source for a bare
// `audit show`.
func (a *App) auditOptions() ([]prompt.Option, error) {
	audits, _, err := a.Svc.ListAudits("", true)
	if err != nil {
		return nil, err
	}
	opts := make([]prompt.Option, 0, len(audits))
	for _, ad := range audits {
		opts = append(opts, labeledOption(ad.Slug, ad.Area))
	}
	return opts, nil
}

// auditMoveOptions lists audits NOT already in bucket `to` — the picker for a bare
// `audit close/reopen/defer`, mirroring transitionOptions(to) for the task verbs.
func (a *App) auditMoveOptions(to domain.AuditBucket) func() ([]prompt.Option, error) {
	return func() ([]prompt.Option, error) {
		audits, _, err := a.Svc.ListAudits("", true)
		if err != nil {
			return nil, err
		}
		opts := make([]prompt.Option, 0, len(audits))
		for _, ad := range audits {
			if ad.Bucket == to {
				continue
			}
			opts = append(opts, labeledOption(ad.Slug, ad.Area))
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
