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

// verifyTaskGraphSourceSnapshot is the whole-repository CAS for graph-aware writes,
// including repair. Unreadable source revisions participate in equality, so an
// unchanged parse error cannot hide changed bytes. Callers add conflict context.
func verifyTaskGraphSourceSnapshot(expected, current *core.TaskGraph) error {
	if !expected.SameSourceSnapshot(current) {
		return domain.ErrConflict
	}
	return nil
}

// verifyThreadSourceSnapshot compares complete Thread source evidence, including
// unreadable revisions, without exposing persistence tokens to planners/renderers.
// This also protects repair operations that tolerate malformed Threads.
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

// verifyUnchanged checks selected ownership and content immediately before writing.
// The caller holds the write guard and supplies an established key with its exact
// resolver, never a fuzzy query or a key inferred here from filename shape. Epics
// use whole stems even when those resemble ID-led task names.
//
// A missing/moved source or changed hash returns ErrConflict; ambiguity and I/O
// errors retain their sentinels. Empty ifVersion skips only the content comparison.
// Use verifyUnchangedWithLabel when diagnostics need a label distinct from key.
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
		// Retry policy belongs to the caller; absence never authorizes alias fallback.
		return conflict()
	case err != nil:
		// Preserve actionable ambiguity/I/O errors instead of inviting futile retries.
		return fmt.Errorf("re-resolve %s %q before writing: %w", noun, label, err)
	case curPath != path:
		return conflict()
	}
	if ifVersion == "" {
		return nil // unconditional: no prior version to compare against
	}
	cur, err := os.ReadFile(curPath)
	if err != nil {
		// Disappearance is drift; other read failures are not evidence of contention.
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
