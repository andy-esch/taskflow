package appwiring

import (
	"github.com/andy-esch/taskflow/internal/cli/ports"
	"github.com/andy-esch/taskflow/internal/configui"
	"github.com/andy-esch/taskflow/internal/tui"
)

func runBrowser(request ports.Browser) error {
	options := []tui.Option{
		tui.WithConfiguration(request.Configuration, request.ConfigurationStart, request.Overrides),
		tui.WithWorkspaceOpening(request.Workspaces),
		tui.WithAtlas(request.Overview),
		tui.WithAtlasTheme(request.AtlasTheme),
	}
	if request.LandOnAtlas {
		options = append(options, tui.WithAtlasLanding())
	}
	return tui.Run(request.Workspace, request.Theme, options...)
}

func runConfiguration(request ports.ConfigurationEditor) error {
	return configui.Run(request.Service, request.Start, request.Overrides, request.Dark, request.In, request.Out)
}
