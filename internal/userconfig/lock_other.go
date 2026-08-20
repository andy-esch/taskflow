//go:build !unix

package userconfig

// registryWriteLock is a no-op on unsupported platforms. Release artifacts target only
// darwin and linux (both unix); keeping this fallback preserves package portability for
// downstream builds without pretending to provide a cross-process guarantee there.
func registryWriteLock(_ string) (func(), error) {
	return func() {}, nil
}
