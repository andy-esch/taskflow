//go:build !unix

package store

// writeLock is a no-op on non-unix platforms — syscall.Flock isn't available there, so the
// cooperating-writer serialization the version-CAS relies on is NOT provided. This is a
// KNOWN GAP for Windows (tracked as a follow-up): concurrent writers can still drop
// updates as they did before the flock fix. The version-CAS itself still runs and catches
// a non-cooperating (out-of-band) edit; only the same-tool concurrency guarantee is
// missing. The primary deployment (Linux container + macOS host) is unix, where the real
// lock applies.
func (s *FS) writeLock() (func(), error) {
	return func() {}, nil
}
