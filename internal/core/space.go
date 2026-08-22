package core

// SpaceRole describes how a registered checkout reaches its planning tree.
type SpaceRole string

const (
	SpaceRoleDirect  SpaceRole = "direct"
	SpaceRolePointer SpaceRole = "pointer"
	SpaceRoleUnknown SpaceRole = "unknown"
)

// SpaceState is the stable health vocabulary for one registered checkout.
type SpaceState string

const (
	SpaceStateOK         SpaceState = "ok"
	SpaceStateEmpty      SpaceState = "empty"
	SpaceStateMissing    SpaceState = "missing"
	SpaceStateNotARepo   SpaceState = "not-a-repo"
	SpaceStateUnreadable SpaceState = "unreadable"
	SpaceStateMismatch   SpaceState = "mismatch"
)

// Healthy reports whether an entry can be used to read its planning tree. Keeping this
// derived from State prevents adapters from supplying contradictory state/health values.
func (s SpaceState) Healthy() bool {
	return s == SpaceStateOK || s == SpaceStateEmpty
}

// SpaceEntryPoint is one registered checkout that reaches a logical planning space.
// ID is the machine-local address; PlanningID is the durable identity shared by every
// entry point into the same planning tree.
type SpaceEntryPoint struct {
	ID   string
	Path string
	// Checkout is the resolved local path used to enter this registered checkout.
	// Path retains the persisted, human-facing spelling (including ~) for output.
	Checkout   string
	VerifyID   string
	PlanningID string
	Role       SpaceRole
	Label      string
	Accent     string
	Added      string
	State      SpaceState
	Root       string
	Detail     string
	Remedy     string
}

func (e SpaceEntryPoint) Healthy() bool { return e.State.Healthy() }

// SpaceGroup is one logical planning identity and all registered ways into it. Entries
// retain registry order.
type SpaceGroup struct {
	PlanningID string
	Entries    []SpaceEntryPoint
}

// SpaceCatalog is one registry snapshot in its two useful projections. Entries retain
// registry order for machine output and completion; Groups retain first-seen group and
// entry order for cross-space consumers.
type SpaceCatalog struct {
	Entries []SpaceEntryPoint
	Groups  []SpaceGroup
}

// SpaceRegistration is a validated planning checkout prepared by the secondary adapter.
// The service supplies ID after applying its defaulting and validation policy.
type SpaceRegistration struct {
	ID       string
	Path     string
	VerifyID string
}

// SpaceMutation is the reusable receipt for a registry add or forget operation.
type SpaceMutation struct {
	Entry   SpaceEntryPoint
	Changed bool
	DryRun  bool
}
