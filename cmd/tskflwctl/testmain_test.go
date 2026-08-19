package main

import (
	"os"
	"testing"

	"github.com/andy-esch/taskflow/internal/userconfig"
)

// TestMain pins the home-config dir at an empty temp dir. These tests exec the real
// binary as a SUBPROCESS, which inherits this environment — so without it a
// developer's own ~/.config/tskflwctl/config.toml would leak into the smoke tests'
// output and exit codes.
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
