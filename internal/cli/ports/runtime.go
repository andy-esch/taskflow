// Package ports defines the CLI's invocation-level composition contracts.
// It contains no adapter construction, discovery, or package-global defaults.
package ports

import (
	"io"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/design"
)

// Planning is an opened local invocation, not a semantic planning entity.
// Repository contains the startup identity and presentation settings observed
// during the same discovery that selected Service; Layout is explicitly local.
type Planning struct {
	Repository core.RepositoryConfiguration
	Service    *core.Service
	Layout     core.Layout
}

// Services is the named application bundle supplied to one command tree.
// OpenPlanning remains lazy: calling Compose must not discover a repository.
// Its implementations must preserve the supplied mutation authorization in all
// planning stores, including those opened later by the TUI workspace service.
type Services struct {
	Configuration *core.ConfigurationService
	Spaces        *core.SpaceRegistryService
	Overview      *core.SpaceOverviewService
	Workspaces    *core.WorkspaceService
	OpenPlanning  func(start string) (Planning, error)
}

// Bindings are explicit per-invocation dependencies, not a service locator.
// The binary selects their implementations; the CLI never selects a local
// fallback. Empty bindings support metadata-only command-tree construction.
// ReadRepository/ReadUser are local presentation reads needed before command
// execution (help chrome) or outside a planning repository (init/version).
type Bindings struct {
	// Compose validates persistence policy without invoking authorizeMutation.
	// On failure the controller discards the entire bundle, including partial values.
	Compose           func(authorizeMutation func() error) (Services, error)
	ReadRepository    func(start string) (core.RepositoryConfiguration, error)
	ReadUser          func() (core.UserConfiguration, error)
	IsMissingPlanning func(error) bool
	RunBrowser        func(Browser) error
	RunConfiguration  func(ConfigurationEditor) error
}

// Browser is an explicit launch request, independent of Bubble Tea options.
// UI startup requires the full Configuration/Spaces/Workspaces/Overview bundle
// before either landing route; opening in a repo still exposes atlas and config.
type Browser struct {
	Workspace          core.Workspace
	Theme              design.Theme
	AtlasTheme         design.Theme
	Configuration      *core.ConfigurationService
	ConfigurationStart string
	Overrides          core.ConfigurationOverrides
	Workspaces         *core.WorkspaceService
	Overview           *core.SpaceOverviewService
	LandOnAtlas        bool
}

type ConfigurationEditor struct {
	Service   *core.ConfigurationService
	Start     string
	Overrides core.ConfigurationOverrides
	Dark      bool
	In        io.Reader
	Out       io.Writer
}
