//go:build !unix

package config

// Release targets are Unix. This fallback preserves package portability without
// claiming a cross-process guarantee on platforms where syscall.Flock is absent.
func writeLock(_ string) (func(), error) { return func() {}, nil }
