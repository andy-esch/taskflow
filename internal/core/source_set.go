package core

import (
	"fmt"
	"sync/atomic"

	"github.com/andy-esch/taskflow/internal/domain"
)

// SourceSetID identifies one adapter-owned planning corpus within this process.
// It is not a record key, a location, a revision, or a durable planning-space ID.
// A secondary adapter mints one value and shares it with every independently
// composable capability that addresses that corpus. The zero value is invalid.
// This is a wiring witness, not a security credential.
type SourceSetID struct{ nonce uint64 }

var nextSourceSetID atomic.Uint64

func NewSourceSetID() SourceSetID { return SourceSetID{nonce: nextSourceSetID.Add(1)} }

func (id SourceSetID) IsZero() bool { return id.nonce == 0 }

// SourceSetProvider is the small composition witness for planning-data ports.
// One complete adapter may implement many ports; split readers, path sources,
// and mutators must publish the same non-zero ID before they can be paired.
// This method must be stable and side-effect-free; construction calls it before
// any planning-data operation.
type SourceSetProvider interface {
	SourceSetID() SourceSetID
}

// ErrIncompatibleCapabilities classifies a missing or mismatched source-set
// witness. It also wraps ErrValidation so existing CLI exit classification holds.
var ErrIncompatibleCapabilities = fmt.Errorf("%w: incompatible planning capabilities", domain.ErrValidation)

type sourceSetCapability struct {
	name  string
	value any
}

func validateSourceSet(capabilities []sourceSetCapability) error {
	var firstName string
	var firstID SourceSetID
	for _, capability := range capabilities {
		if isNilCapability(capability.value) {
			continue
		}
		provider, ok := capability.value.(SourceSetProvider)
		if !ok {
			return fmt.Errorf("%w: %s does not publish a source-set identity", ErrIncompatibleCapabilities, capability.name)
		}
		id := provider.SourceSetID()
		if id.IsZero() {
			return fmt.Errorf("%w: %s publishes an empty source-set identity", ErrIncompatibleCapabilities, capability.name)
		}
		if id != provider.SourceSetID() {
			return fmt.Errorf("%w: %s publishes an unstable source-set identity", ErrIncompatibleCapabilities, capability.name)
		}
		if firstName == "" {
			firstName, firstID = capability.name, id
			continue
		}
		if id != firstID {
			return fmt.Errorf("%w: %s and %s address different source sets", ErrIncompatibleCapabilities, firstName, capability.name)
		}
	}
	return nil
}
