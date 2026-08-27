//go:build !unix

package store

import (
	"fmt"

	"github.com/andy-esch/taskflow/internal/domain"
)

// platformWriteLock fails explicitly on targets for which taskflow has no tested
// cooperating-writer primitive. Silent no-op locking would make a successful CAS
// claim untrue and is therefore less safe than rejecting mutation.
func (s *FS) platformWriteLockChecked() (func() error, error) {
	return nil, fmt.Errorf("%w: repository mutation locking is unsupported on this platform", domain.ErrValidation)
}
