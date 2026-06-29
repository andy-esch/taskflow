package domain

// Transition is one lifecycle action: the verb a user names (from the CLI or the
// TUI action menu) mapped to the destination state it moves a document to. It is
// the verb→state half of the lifecycle vocabulary, the missing peer to
// AllStatuses()/AllAuditBuckets()/AllEpicStatuses() (which own the state half).
//
// To is a string because the same shape serves both entities: it is a task
// Status for TaskTransitions() and an AuditBucket for AuditTransitions(); each
// consumer casts it back to the typed state for the entity in view. Destructive
// marks a move an interactive surface should gate behind a confirm (an
// archiving move); it is shared data — a non-interactive surface (the CLI) is
// free to ignore it.
//
// One registry, two adapters: the CLI builds its per-verb commands from these
// tables and the TUI's action menu is a thin view of them, so adding or renaming
// a verb is a single edit here rather than a drift between surfaces. Param marks a
// verb that takes an extra optional argument (the task defer's revisit date), so
// adapters read it from the registry instead of hardcoding which verb is special.
type Transition struct {
	Verb        string
	To          string
	Destructive bool
	Param       TransitionParam
}

// TransitionParam marks an optional argument a lifecycle verb accepts beyond the
// move itself. Today the only non-None case is the task `defer`, which takes an
// optional revisit (snooze) date. It is a typed marker, NOT a full param spec: the
// registry declares THAT a verb takes a date so adapters stop hardcoding "is this
// defer?", while HOW each collects it (a --until flag, a TUI date input) stays
// adapter-specific. Add a value here if another verb ever grows a parameter.
type TransitionParam int

const (
	ParamNone         TransitionParam = iota // takes only the target document(s)
	ParamOptionalDate                        // an optional date — defer's revisit/snooze
)

// taskTransitions are the task status moves (the working-set lifecycle). Order is
// the declared verb order both surfaces present: the CLI command list and the TUI
// menu. deprecate is the one archiving move, so it is the only destructive row.
var taskTransitions = []Transition{
	{Verb: "start", To: string(StatusInProgress), Destructive: false},
	{Verb: "next", To: string(StatusNextUp), Destructive: false},
	{Verb: "ready", To: string(StatusReadyToStart), Destructive: false},
	{Verb: "complete", To: string(StatusCompleted), Destructive: false},
	{Verb: "defer", To: string(StatusDeferred), Destructive: false, Param: ParamOptionalDate},
	{Verb: "deprecate", To: string(StatusDeprecated), Destructive: true},
}

// auditTransitions are the audit bucket moves, mirroring `audit close/reopen/defer`.
// None is flagged destructive: close/defer to a non-open bucket are guarded by the
// store on still-open findings (that rejection surfaces as a normal error), not by
// a TUI confirm.
var auditTransitions = []Transition{
	{Verb: "close", To: string(AuditClosed), Destructive: false},
	{Verb: "reopen", To: string(AuditOpen), Destructive: false},
	{Verb: "defer", To: string(AuditDeferred), Destructive: false},
}

// TaskTransitions returns the task lifecycle verb→status moves, in declared order.
func TaskTransitions() []Transition { return taskTransitions }

// AuditTransitions returns the audit lifecycle verb→bucket moves, in declared order.
func AuditTransitions() []Transition { return auditTransitions }
