package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
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

// verifyThreadSourceSnapshot is the unreadable-aware half of every graph-aware
// repository write CAS. It compares opaque source evidence without exposing it
// to planners or renderers and remains valid for a future repair path that may
// deliberately continue past malformed Thread documents.
func verifyThreadSourceSnapshot(expected, current core.ThreadRead) error {
	if !sameThreadSourceSnapshot(expected, current) {
		return domain.ErrConflict
	}
	return nil
}

func sameThreadSourceSnapshot(left, right core.ThreadRead) bool {
	if len(left.Records) != len(right.Records) || len(left.Problems) != len(right.Problems) {
		return false
	}
	leftThreads := append([]core.VersionedRecord[domain.Thread](nil), left.Records...)
	rightThreads := append([]core.VersionedRecord[domain.Thread](nil), right.Records...)
	sort.SliceStable(leftThreads, func(i, j int) bool { return threadSourceKey(leftThreads[i]) < threadSourceKey(leftThreads[j]) })
	sort.SliceStable(rightThreads, func(i, j int) bool { return threadSourceKey(rightThreads[i]) < threadSourceKey(rightThreads[j]) })
	if !slices.EqualFunc(leftThreads, rightThreads, sameThreadSource) {
		return false
	}
	leftProblems := append([]core.ThreadReadProblem(nil), left.Problems...)
	rightProblems := append([]core.ThreadReadProblem(nil), right.Problems...)
	sort.SliceStable(leftProblems, func(i, j int) bool {
		return threadProblemSourceKey(leftProblems[i]) < threadProblemSourceKey(leftProblems[j])
	})
	sort.SliceStable(rightProblems, func(i, j int) bool {
		return threadProblemSourceKey(rightProblems[i]) < threadProblemSourceKey(rightProblems[j])
	})
	return slices.EqualFunc(leftProblems, rightProblems, sameThreadProblemSource)
}

func threadSourceKey(thread core.VersionedRecord[domain.Thread]) string {
	return strings.Join([]string{
		thread.Record.Source.ID, thread.Record.Source.Location, thread.LocalPath,
		thread.Record.Value.ID, thread.Record.Value.Slug, thread.SourceVersion,
	}, "\x00")
}

func sameThreadSource(left, right core.VersionedRecord[domain.Thread]) bool {
	return left.Record.Source == right.Record.Source &&
		left.LocalPath == right.LocalPath &&
		left.Record.Value.ID == right.Record.Value.ID &&
		left.Record.Value.Slug == right.Record.Value.Slug &&
		left.SourceVersion != "" && left.SourceVersion == right.SourceVersion
}

func threadProblemSourceKey(problem core.ThreadReadProblem) string {
	return strings.Join([]string{
		problem.ThreadID, problem.ThreadSlug, problem.Location, problem.Message, problem.SourceVersion,
	}, "\x00")
}

func sameThreadProblemSource(left, right core.ThreadReadProblem) bool {
	return left.ThreadID == right.ThreadID && left.ThreadSlug == right.ThreadSlug &&
		left.Location == right.Location && left.Message == right.Message &&
		left.SourceVersion != "" && left.SourceVersion == right.SourceVersion
}

// verifyUnchanged is the version-CAS precondition shared by every write: called
// immediately before the write, it re-resolves the file by its explicit selected
// adapter key, never a raw fuzzy query or an identity guessed from a filename,
// and re-reads the source to confirm nothing raced us between
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
// The caller establishes key and supplies its matching resolver. Epics use whole
// stems even when they resemble an ID-led task name; this helper does not decode
// either convention. Callers with a distinct diagnostic label use
// verifyUnchangedWithLabel without changing the resolver key.
func verifyUnchanged(resolve func(string) (string, error), key, path, ifVersion, noun, op string) error {
	return verifyUnchangedWithLabel(resolve, key, key, path, ifVersion, noun, op)
}

func verifyUnchangedWithLabel(resolve func(string) (string, error), key, label, path, ifVersion, noun, op string) error {
	conflict := func() error {
		return fmt.Errorf("%s %q changed on disk during %s; retry: %w", noun, label, op, domain.ErrConflict)
	}
	curPath, err := resolve(key)
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
		return fmt.Errorf("re-resolve %s %q before writing: %w", noun, label, err)
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
		return fmt.Errorf("re-read %s %q: %w", noun, label, err)
	}
	if hashContent(cur) != ifVersion {
		return conflict()
	}
	return nil
}
