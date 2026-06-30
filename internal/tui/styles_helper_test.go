package tui

import "github.com/andy-esch/taskflow/internal/design"

// testStyles is the default-themed styles bundle for tests that call the chrome
// render helpers directly (production builds it in New / repopulates in Run).
var testStyles = newStyles(design.Default().Dark)
