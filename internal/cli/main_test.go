package cli

import (
	"os"
	"testing"

	"github.com/andy-esch/taskflow/internal/userconfig"
)

// TestMain pins the home-config directory at an empty temp dir for every test in
// this package.
//
// This is not hygiene, it is load-bearing. These tests execute NewRootCmd in-process
// and diff GOLDEN output; setStyle loads the user config, so without this a
// developer who happens to have a ~/.config/tskflwctl/config.toml pinning a theme
// would watch unrelated goldens fail for reasons that look inexplicable. Nothing in
// this package may read or write a real $HOME.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tskflwctl-cfghome-")
	if err != nil {
		panic("create temp config home: " + err.Error())
	}
	if err := os.Setenv(userconfig.DirEnv, dir); err != nil {
		panic("pin config home: " + err.Error())
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}
