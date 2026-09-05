package domain

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// corpusHeadings are real heading lines taken from the repository's own audits —
// the shapes a code-agnostic near-miss pattern was measured at 116 false positives
// against. They are pinned here as a fixture so the recognizer's narrowing cannot
// regress silently even when the live corpus changes.
//
// Sourced from the 12 audits that use numbered section headings, plus the report
// scaffolding every reviewer brief carries.
var corpusHeadings = []string{
	"### 1. Lifecycle and Dependency Semantics",
	"### 2. Coverage of Every First-Party Start / Status Path",
	"### 3. Force-Scope Separation",
	"### 4. Global Graph-Health Behavior",
	"### 5. Resulting Active-Task Validity",
	"### 1. Thread Domain Invariants, Document Parsing, Stable Identity",
	"### 2. Projection Semantics",
	"### 3. Guarded Creation",
	"### 4. Safe Legacy Projects Retirement",
	"### 1. Domain Semantics",
	"### 2. Mutation Authorization and Safety",
	"### 3. Read Projections",
	"### 4. API and Compatibility Contracts",
	"### 1. Full Lifecycle Transition Matrix",
	"### 2. User Reference Resolution & Planner Execution",
	"### 3. Store Transaction Sequence",
	"### 4. Concurrency & Independent-Store Serialization",
	"### 1. File Formats & Strict Manifest Parsing",
	"### 2. External Gate (`member: false`) Reachability",
	"### 3. Compose End-to-End Execution",
	"### 4. Apply Authorization & Live Rediscovery",
	"### 1. Complete Read Inventory vs. Guarded Mutations",
	"### 2. `NewService` Construction Semantics",
	"### 3. Port Shape & Consumer Ownership",
	"### 4. Split-Source Snapshot Consistency",
	"### 1. Global DAG Node & Edge Derivation",
	"### 2. Topology Semantics & Health Qualification",
	"### 3. Member Waves vs. External Gates",
	"### 4. Determinism & Scan Order Invariance",
	"#### 1. End-to-End Planning Space Probe (9 Required Shapes)",
	"#### 2. Hostile Identity Fixtures",
	"#### 3. Completion-Order & Stale Message Interleaving Probes",
	"#### 4. Capability-Boundary Probes",
	"#### 1. Consumer and composition inventory",
	"#### 2. End-to-end throwaway-space probes ledger",
	"#### 3. Split-authority adapter probe ledger",
	"#### 4. Multi-space probes ledger",
	"### 1. What is the precise repair invariant?",
	"### 3. Is “every durable prefix must improve” achievable for cycles?",
	"### 1. Executive Verdict",
	"### 2. Environment, Validation, and Isolation Attestation",
	"### 3. Source-to-CAS Data Flow and Production Consumer Inventory",
	"#### 3.1 Data Flow Architecture",
	"## Context: how it stayed red",
	"## Findings",
	"## Candidate tasks",
	"## Reviewer report",
	"### Findings",
	"## Mandatory reviewer sandbox",
	"## Review brief",
	"## Validation and restoration",
	"## Deliverable",
}

// AC: zero false positives across the shapes the existing corpus actually uses.
func TestNearMissFindingHeaders_ZeroFalsePositivesOnCorpusHeadings(t *testing.T) {
	for _, line := range corpusHeadings {
		if hits := NearMissFindingHeaders(line + "\n"); len(hits) != 0 {
			t.Errorf("false positive on corpus heading %q → would rewrite to %q", line, hits[0].Canonical)
		}
	}
}

// The live corpus is the real gate: run the recognizer over every audit in the
// repository and require zero hits. If someone later widens the pattern, this fails
// with the exact heading it would have corrupted. Skipped when the planning tree is
// not beside the package (a consumer vendoring the module).
func TestNearMissFindingHeaders_LiveCorpusIsClean(t *testing.T) {
	dir := filepath.Join("..", "..", "planning", "audits")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("planning corpus unavailable: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		scanned++
		for _, h := range NearMissFindingHeaders(string(b)) {
			t.Errorf("%s line %d: recognizer claims %q (would rewrite to %q)", e.Name(), h.Line, h.Text, h.Canonical)
		}
	}
	if scanned == 0 {
		t.Skip("no audits found")
	}
	t.Logf("scanned %d audits, zero near-miss claims", scanned)
}

// The corpus must also still parse the findings it has — the recognizer must not
// have stolen any canonical header away from ParseFindings.
func TestNearMissFindingHeaders_DoesNotShadowCanonicalFindings(t *testing.T) {
	dir := filepath.Join("..", "..", "planning", "audits")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("planning corpus unavailable: %v", err)
	}
	total := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		body := string(b)
		total += len(ParseFindings(body))
		// Repair must be a no-op on a corpus that is already canonical.
		fixed, changed := CanonicalizeFindingHeaders(body)
		if len(changed) != 0 || fixed != body {
			t.Errorf("%s: `lint --fix` would rewrite an already-canonical audit", e.Name())
		}
	}
	if total == 0 {
		t.Skip("no findings found")
	}
	t.Logf("corpus parses %d canonical findings, all untouched", total)
}
