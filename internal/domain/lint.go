package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Issue is a single frontmatter lint finding.
type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// LintTask returns the frontmatter issues for an (active) task. validEpic
// reports whether an epic id exists; pass nil to skip the epic-existence check.
func LintTask(t Task, validEpic func(string) bool) []Issue {
	var issues []Issue
	add := func(field, msg string) { issues = append(issues, Issue{Field: field, Message: msg}) }

	switch {
	case t.Epic == "":
		add("epic", "missing")
	case validEpic != nil && !validEpic(t.Epic):
		add("epic", fmt.Sprintf("unknown epic %q", t.Epic))
	}
	switch {
	case t.Tier == 0:
		add("tier", "missing")
	case t.Tier < 1 || t.Tier > 5:
		add("tier", "must be 1-5")
	}
	switch {
	case t.Priority == "":
		add("priority", "missing")
	case !validPriorities[t.Priority]:
		add("priority", "must be high|medium|low")
	}
	if t.Autonomy != 0 && (t.Autonomy < 1 || t.Autonomy > 5) {
		add("autonomy_level", "must be 1-5")
	}
	if t.Effort == "" {
		add("effort", "missing")
	}
	switch {
	case t.Created == "":
		add("created", "missing")
	case ValidateDate(t.Created) != nil:
		add("created", "must be YYYY-MM-DD")
	}
	if len(t.Tags) == 0 {
		add("tags", "missing")
	}

	switch {
	case t.Description == "":
		if t.Status == StatusNextUp || t.Status == StatusInProgress {
			add("description", "required for next-up/in-progress")
		}
	case strings.ContainsAny(t.Description, "\r\n"):
		add("description", "must be a single line")
	default:
		// Count CHARACTERS, not bytes, to match ValidateDescription — a multibyte
		// description must not be flagged here after passing creation validation.
		if n := utf8.RuneCountInString(t.Description); n > MaxDescriptionLen {
			add("description", fmt.Sprintf("too long (%d > %d chars)", n, MaxDescriptionLen))
		}
	}

	issues = append(issues, MisfiledIssues(t)...)
	issues = append(issues, FrontmatterStatusIssues(t)...)
	issues = append(issues, MissingIDIssue(t.ID)...)
	return issues
}

// FrontmatterStatusIssues flags a task whose frontmatter `status` was missing or named
// no recognized status — now that frontmatter is the authority (ADR-0003 Phase A) that's
// a real defect; the folder only silently covered for it (parseTask fell back). Applies
// in ANY status (an archived task's broken status surfaces too, beside MisfiledIssues),
// and is fail-open: the task still lists and resolves via the fallback, it's just flagged.
func FrontmatterStatusIssues(t Task) []Issue {
	if !t.StatusFellBack {
		return nil
	}
	return []Issue{{Field: "status", Message: "frontmatter status missing or unrecognized — set it with the lifecycle verb for its state (`task start`/`next`/`ready`/`complete`/`deprecate`)"}}
}

// MissingIDIssue flags an entity (task or audit) that has no stable id yet — the
// pre-assignment state a one-time `lint --fix` backfills (ADR-0003). Applies in
// ANY status: an archived task still needs a stable key for links and reopening,
// so — like MisfiledIssues — it's checked outside the active-only field block.
func MissingIDIssue(id string) []Issue {
	if strings.TrimSpace(id) != "" {
		return nil
	}
	return []Issue{{Field: "id", Message: MissingIDMessage}}
}

// MissingIDMessage is the plain-lint wording for an entity with no stable id yet —
// before `--fix` has had a chance to backfill one. Exported so the fix flow can
// recognize this exact finding among leftovers and restate it (see
// UnrepairedIDMessage) without matching on the loosely-shared "id" field.
const MissingIDMessage = "missing stable id — `lint --fix` assigns one"

// UnrepairedIDMessage restates a missing-id finding that survived `lint --fix`:
// the backfiller found no date to mint an id from (no created/…/deprecated_at
// field and no YYYY-MM-DD filename prefix), so plain lint's "assigns one" wording
// would misdirect — the fix already ran. The fix flow swaps in this remedy.
const UnrepairedIDMessage = "no date to mint an id from — add a `created: YYYY-MM-DD` field (or a YYYY-MM-DD- filename prefix), then re-run `lint --fix`"

// LintEpic returns the frontmatter issues for an epic. Mirrors LintTask, but
// epics have no validEpic dependency (they're the join target, not a referrer)
// and no status-directory drift (status is a flat frontmatter field, not a
// folder). The status vocabulary is always checked; a `deprecated` epic is
// withdrawn, so — like an archived task — it's spared the field nags (no point
// demanding a priority/description on a dead goal).
func LintEpic(e Epic) []Issue {
	var issues []Issue
	add := func(field, msg string) { issues = append(issues, Issue{Field: field, Message: msg}) }

	// Always: the status must be in the closed vocabulary. Files predating the
	// enum (or hand-edited ones) surface here regardless of active/deprecated.
	if err := ValidateEpicStatus(e.Status); err != nil {
		add("status", err.Error())
	}
	// A deprecated epic is dead/withdrawn — stop at the status check, don't nag
	// about the active-only fields below (mirrors MisfiledIssues-only for
	// archived tasks).
	if e.Status == "deprecated" {
		return issues
	}

	switch {
	case e.Priority == "":
		add("priority", "missing")
	case !validPriorities[e.Priority]:
		add("priority", "must be high|medium|low")
	}
	if e.Description == "" {
		add("description", "missing")
	}
	return issues
}

// MisfiledIssues reports the status/folder mismatch for a task, if any. It is
// separate from the active-only field checks so archived tasks (completed/…)
// can still be flagged for drift without nagging about missing fields.
func MisfiledIssues(t Task) []Issue {
	if !t.Misfiled() {
		return nil
	}
	return []Issue{{
		Field: "status",
		Message: fmt.Sprintf("frontmatter says %q but file is in %s/ — frontmatter wins; `lint --fix` moves it",
			t.Status, t.FolderStatus),
	}}
}

// AuditMisfiledIssues reports the bucket/folder mismatch for an audit, if any — the
// audit analog of MisfiledIssues (frontmatter bucket is authoritative; a stale folder
// is the drift).
func AuditMisfiledIssues(a Audit) []Issue {
	if !a.Misfiled() {
		return nil
	}
	return []Issue{{
		Field: "bucket",
		Message: fmt.Sprintf("frontmatter says %q but file is in %s/ — frontmatter wins; `lint --fix` moves it",
			a.Bucket, a.FolderBucket),
	}}
}

// FrontmatterBucketIssues flags an audit whose frontmatter `bucket` was missing or named
// no recognized bucket — the audit analog of FrontmatterStatusIssues. Fail-open: the audit
// still lists (with BucketFellBack set); clear it by setting the bucket via a lifecycle verb.
func FrontmatterBucketIssues(a Audit) []Issue {
	if !a.BucketFellBack {
		return nil
	}
	return []Issue{{Field: "bucket", Message: "frontmatter bucket missing or unrecognized — set it with `audit close`/`reopen`/`defer`"}}
}
