package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/andy-esch/taskflow/internal/core"
	"github.com/andy-esch/taskflow/internal/domain"
)

// hashContent is the version-CAS token (epic 24 / ADR-0003): the SHA-256 of a file's
// exact on-disk bytes, hex-encoded. It is computed ON READ and NEVER stored in the file —
// a stored token would be self-referential (it would change the very bytes it
// fingerprints). It is a *strong* validator (byte equality), which is what a lost-update
// guard needs: two writers who touched different fields must both be caught, not waved
// through by a normalized/"weak" hash. Cryptographic on purpose — the token IS the safety
// check, so a collision would be a silent lost update, and the bytes are agent-authored
// (attacker-influenceable) task/audit content.
func hashContent(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// verifyTaskGraphSourceSnapshot is the shared pre-write whole-repository CAS for
// every graph-aware mutation, including the forthcoming broken-graph repair path.
// It is deliberately valid for broken snapshots: unreadable source revisions are
// part of SameSourceSnapshot, so a raw edit cannot hide behind an unchanged parse
// error. Callers add operation-specific conflict context around the sentinel.
func verifyTaskGraphSourceSnapshot(expected, current *core.TaskGraph) error {
	if !expected.SameSourceSnapshot(current) {
		return domain.ErrConflict
	}
	return nil
}

// verifyUnchanged is the version-CAS precondition shared by every write: called
// immediately before the write, it re-resolves the file (by its canonical slug, not the
// caller's raw fuzzy query) and re-reads the source to confirm nothing raced us between
// the read-and-transform and the write. It returns
// domain.ErrConflict (exit 14 — the same sentinel today's path-CAS produces) when the
// file either:
//
//   - moved or vanished (curPath != path) — a concurrent relocation; the check the
//     path-CAS re-resolve has always made, guarding against resurrecting the slug in its
//     old status/bucket dir; or
//   - changed content (hash(current) != ifVersion) — a concurrent in-place body/field
//     edit, which the path-CAS alone silently missed. This is the coverage version-CAS
//     adds, so this guard is a strict superset of the old re-resolve block.
//
// ifVersion is the hash of the bytes the caller read; an empty ifVersion skips the
// content check (an unconditional write — the caller held no prior version to compare).
// resolve is s.resolve / s.resolveAudit adapted to (path, error), keeping the guard
// entity-agnostic; noun/op shape the error to match the existing per-site wording
// ("task %q changed on disk during move", …) so routing the call sites through here is
// behaviour-preserving.
func verifyUnchanged(resolve func(string) (string, error), slug, path, ifVersion, noun, op string) error {
	conflict := func() error {
		return fmt.Errorf("%s %q changed on disk during %s; retry: %w", noun, slug, op, domain.ErrConflict)
	}
	// Re-resolve by the CANONICAL slug (the resolved file's basename), never the caller's
	// raw query: re-running a fuzzy/partial query could newly match a concurrently-created
	// same-prefix file and return ErrAmbiguous, which we'd read as a conflict — a spurious
	// rejection of an edit to a file that never changed. The canonical slug matches
	// exactly (resolveID is exact-first) and deterministically, so only a genuine move or
	// dup of THIS file trips the path check.
	// Re-resolve by a key the resolver actually matches: for a flat, id-led entity
	// (`<id>-<slug>`) that is the stable id — the full stem matches neither the id
	// nor the slug candidate key. For a bucketed audit or an epic (basename not
	// id-led) the full stem IS the resolution key. Either way it matches THIS file
	// exactly and deterministically.
	canonical := strings.TrimSuffix(filepath.Base(path), ".md")
	if entityID, _, ok := splitFlatName(canonical); ok {
		canonical = entityID
	}
	curPath, err := resolve(canonical)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		// The file we read is no longer resolvable under its own id: moved, renamed,
		// or deleted under us. That IS the conflict this guard exists for, and a
		// retry re-reads from scratch and can legitimately succeed.
		return conflict()
	case err != nil:
		// Everything else — a genuine duplicate id (ErrAmbiguous), or a real ReadDir
		// I/O failure — is NOT a conflict, and no number of retries can fix it.
		// Collapsing it into one would burn the whole OCC budget and then hand back
		// "changed on disk; retry": advice that cannot work, for a condition it
		// misnames. Propagating keeps the sentinel the CLI maps to its own exit code
		// (13 for the ambiguity, which names the colliding files). This mirrors the
		// os.IsNotExist split in the re-read below.
		return fmt.Errorf("re-resolve %s %q before writing: %w", noun, slug, err)
	case curPath != path:
		return conflict()
	}
	if ifVersion == "" {
		return nil // unconditional: no prior version to compare against
	}
	cur, err := os.ReadFile(curPath)
	if err != nil {
		// A vanished file is a genuine conflict (it changed under us → retry). Any OTHER
		// read failure — permission, I/O, fd exhaustion — is a real error the caller must
		// see, not be misled into retrying a "changed on disk" conflict that isn't one.
		if os.IsNotExist(err) {
			return conflict()
		}
		return fmt.Errorf("re-read %s %q: %w", noun, slug, err)
	}
	if hashContent(cur) != ifVersion {
		return conflict()
	}
	return nil
}
