package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

// standardEpicNameRe matches the epic filename convention NN-<slug> (a zero-padded
// number, 2+ digits, a dash, then a non-empty slug). Epics are ordered by that number
// (epicNum) and read consistently when every stem carries it; a name without it — or
// with an empty slug (`01-`) — still lists and resolves (fail-open) but is lint-flagged
// (EpicNameIssue).
var standardEpicNameRe = regexp.MustCompile(`^\d{2,}-.`)

// EpicNameIssue flags an epic whose filename stem does not follow the NN-<slug>
// convention. Fail-open, like FrontmatterStatusIssues: the epic is untouched and
// still usable, the name is just called out for a rename.
func EpicNameIssue(id string) []Issue {
	if standardEpicNameRe.MatchString(id) {
		return nil
	}
	return []Issue{{Field: "filename", Message: fmt.Sprintf("epic filename %q should be NN-<slug> (a zero-padded number) — rename it so epics order consistently", id)}}
}

// Issue is a single frontmatter lint finding.
type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// DuplicateEpicNNIssues flags epics that share a leading NN key. Two epics on the same key
// (e.g. `01-a`, `01-b`) co-mingle their tasks in the rollup and canonicalEpic silently
// resolves an `epic:` ref to the first — an invalid state nothing else enforces. Returns one
// issue per colliding epic, keyed by epic id (nothing for a unique key). Fail-open, like the
// other epic checks: the epics still list and resolve, the clash is just called out for a
// renumber. It's cross-epic (needs the whole set), so it lives here, not in per-epic LintEpic.
func DuplicateEpicNNIssues(epicIDs []string) map[string]Issue {
	byKey := make(map[string][]string, len(epicIDs))
	for _, id := range epicIDs {
		key := EpicRefKey(id)
		byKey[key] = append(byKey[key], id)
	}
	out := make(map[string]Issue)
	for key, ids := range byKey {
		if len(ids) < 2 {
			continue
		}
		sort.Strings(ids)
		for _, id := range ids {
			peers := make([]string, 0, len(ids)-1)
			for _, other := range ids {
				if other != id {
					peers = append(peers, other)
				}
			}
			out[id] = Issue{
				Field: "filename",
				Message: fmt.Sprintf(
					"duplicate epic NN key %q (shared with %s) — tasks co-mingle in the rollup and epic refs resolve to the first; give each epic a unique number",
					key, strings.Join(peers, ", ")),
			}
		}
	}
	return out
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

	issues = append(issues, FrontmatterStatusIssues(t)...)
	issues = append(issues, MissingIDIssue(t.ID)...)
	issues = append(issues, IDDriftIssue(t.ID, t.FilenameID)...)
	return issues
}

// FrontmatterStatusIssues flags a task whose frontmatter `status` is missing or names no
// recognized status — under the flat layout (ADR-0003 §4) frontmatter is the sole
// authority, so that's a real defect with no directory to cover for it. Applies in ANY
// status, and is fail-open: the task still lists (with its raw status), it's just flagged.
func FrontmatterStatusIssues(t Task) []Issue {
	if !t.StatusFellBack {
		return nil
	}
	return []Issue{{Field: "status", Message: "frontmatter status missing or unrecognized — set it with the lifecycle verb for its state (`task start`/`next`/`ready`/`complete`/`deprecate`)"}}
}

// MissingIDIssue flags an entity (task or audit) that has no stable id yet — the
// pre-assignment state a one-time `lint --fix` backfills (ADR-0003). Applies in
// ANY status: an archived task still needs a stable key for links and reopening,
// so — like FrontmatterStatusIssues — it's checked outside the active-only field block.
func MissingIDIssue(id string) []Issue {
	if strings.TrimSpace(id) != "" {
		return nil
	}
	return []Issue{{Field: "id", Message: MissingIDMessage}}
}

// IDDriftIssue flags a frontmatter `id:` that disagrees with the id in the flat
// filename (filenameID — the canonical key resolveID/CAS match on). Post-flatten the
// two are minted together and kept in lock-step, so a hand-edit to one but not the
// other is a silent drift that would make the frontmatter id lie about the file it
// names — surfaced in ANY status, like MissingIDIssue. An empty side is left to
// MissingIDIssue / the id-led scan gate, not double-reported here.
func IDDriftIssue(frontmatterID, filenameID string) []Issue {
	// A blank or whitespace-only frontmatter id is MissingIDIssue's job (it trims too),
	// so defer to it rather than double-reporting the same file as both missing AND drifted.
	if strings.TrimSpace(frontmatterID) == "" || filenameID == "" || frontmatterID == filenameID {
		return nil
	}
	return []Issue{{Field: "id", Message: fmt.Sprintf("frontmatter id %q disagrees with the filename id %q — rename the file or fix the field", frontmatterID, filenameID)}}
}

// MissingIDMessage is the lint wording for an id-led entity whose frontmatter `id:`
// is absent — a copy of the id already in its filename that `lint --fix` fills in
// (backfillMissingID). Post-flatten the filename always carries the id, so --fix can
// always repair this; there is no "unrepairable id" state to restate.
const MissingIDMessage = "missing stable id — `lint --fix` assigns one"

// LintEpic returns the frontmatter issues for an epic. Mirrors LintTask, but
// epics have no validEpic dependency (they're the join target, not a referrer)
// and no status-directory drift (status is a flat frontmatter field, not a
// folder). The status vocabulary is always checked; a `deprecated` epic is
// withdrawn, so — like an archived task — it's spared the field nags (no point
// demanding a priority/description on a dead goal).
func LintEpic(e Epic) []Issue {
	var issues []Issue
	add := func(field, msg string) { issues = append(issues, Issue{Field: field, Message: msg}) }

	// Always (like the status check): the filename must follow the NN-<slug>
	// convention. Applies regardless of active/deprecated, since a stray-named epic
	// mis-orders the roster either way.
	issues = append(issues, EpicNameIssue(e.ID)...)
	// Always: the status must be in the closed vocabulary. Files predating the
	// enum (or hand-edited ones) surface here regardless of active/deprecated.
	if err := ValidateEpicStatus(e.Status); err != nil {
		add("status", err.Error())
	}
	// A deprecated epic is dead/withdrawn — stop at the status check, don't nag
	// about the active-only fields below (as archived tasks are spared them).
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

// FrontmatterBucketIssues flags an audit whose frontmatter `bucket` was missing or named
// no recognized bucket — the audit analog of FrontmatterStatusIssues. Fail-open: the audit
// still lists (with BucketFellBack set); clear it by setting the bucket via a lifecycle verb.
func FrontmatterBucketIssues(a Audit) []Issue {
	if !a.BucketFellBack {
		return nil
	}
	return []Issue{{Field: "bucket", Message: "frontmatter bucket missing or unrecognized — set it with `audit close`/`reopen`/`defer`"}}
}
