---
schema: 1
id: 6g54ay3njm8y
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: 'Configure repository ruleset as code for main branch protection: required checks on today''s CI jobs (Lint/Tests/Planning Lint), no deletions/force-push'
effort: 1 day
tier: 3
priority: medium
autonomy_level: 3
tags: [ci, github, security, infrastructure, quality]
created: "2026-08-30"
audited: "2026-10-11"
audit_sources: [2026-10-11-weekly-task-sweep]
updated_at: "2026-10-11"
---
# Adopt branch protection ruleset as code (.github/ruleset.json)

## Objective

Establish declarative GitHub branch protection for the `main` branch as code in `.github/ruleset.json`, modeled after the pattern in `andy-esch/desirelines`. Enforce the unified `Go Quality` CI status check, prevent accidental branch deletion and force pushes (non-fast-forward), and provide an idempotent `just ruleset-apply` recipe using the `gh` CLI.

## Background & Comparative Context

Currently, `andy-esch/taskflow` has no active branch protection or ruleset configured (`gh ruleset list -R andy-esch/taskflow` returns empty). 

In comparison:
- `andy-esch/desirelines` maintains `.github/ruleset.json` (active ruleset ID `10028286`), which protects `~DEFAULT_BRANCH` with `deletion`, `non_fast_forward`, and strict `required_status_checks`.
- Modern GitHub Repository Rulesets supersede classic branch protection by supporting declarative JSON schemas, multi-branch targeting, and programmatic synchronization via the GitHub REST API (`gh api`).
- `taskflow`'s CI workflow (`.github/workflows/ci.yml`) executes a single comprehensive job named **`Go Quality`**, which runs:
  1. `just tidy-check` (`go.mod` / `go.sum` cleanliness)
  2. `just docs-check` (CLI reference doc sync)
  3. `gofmt` (Go code formatting)
  4. `golangci-lint` (v2 linter suite)
  5. `govulncheck` (package-level vulnerability scanner)
  6. `go test -race -coverprofile=...` (data race detection and unit test suite)
  7. Codecov coverage upload

Requiring `"Go Quality"` in the branch ruleset guarantees that all seven hygiene, quality, and security checks must pass on a branch up-to-date with `main` before any PR can be merged.

## Ruleset Specification (`.github/ruleset.json`)

```json
{
  "name": "main",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["~DEFAULT_BRANCH"],
      "exclude": []
    }
  },
  "rules": [
    {
      "type": "deletion"
    },
    {
      "type": "non_fast_forward"
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "strict_required_status_checks_policy": true,
        "do_not_enforce_on_create": false,
        "required_status_checks": [
          { "context": "Go Quality" }
        ]
      }
    }
  ],
  "bypass_actors": []
}
```

### Policy Decisions & Rationale:
1. **Target `~DEFAULT_BRANCH`:** Dynamically tracks the repository's default branch (`main`) without hardcoding branch names.
2. **`deletion: true`:** Blocks accidental branch deletion via git push or GitHub UI.
3. **`non_fast_forward: true`:** Blocks `git push --force` or `--force-with-lease`, preserving unbroken git commit history on `main`.
4. **`strict_required_status_checks_policy: true`:** Requires PR branches to be updated with latest `main` before merging, preventing semantic conflicts.
5. **No Required PR Approvals:** By omitting the `pull_request` approval rule, solo maintainers and AI pair-programming agents can merge pull requests immediately once all CI checks pass, avoiding blocked workflows while maintaining strict code quality gates.

## Automation & Justfile Recipe

Add an idempotent synchronization recipe to `justfile`:

```just
# Sync branch protection ruleset to GitHub via gh CLI
ruleset-apply:
	#!/usr/bin/env bash
	set -euo pipefail
	REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner)
	ID=$(gh api "repos/${REPO}/rulesets" --jq '.[] | select(.name=="main") | .id' 2>/dev/null || true)
	if [ -n "$ID" ]; then
		echo "Updating ruleset ${ID} on ${REPO}..."
		gh api -X PUT "repos/${REPO}/rulesets/${ID}" --input .github/ruleset.json
	else
		echo "Creating ruleset on ${REPO}..."
		gh api -X POST "repos/${REPO}/rulesets" --input .github/ruleset.json
	fi
	echo "✔ Ruleset synchronized"
```

## Scope

- Create `.github/ruleset.json` with the specified configuration.
- Add `ruleset-apply` to `justfile`.
- Execute `just ruleset-apply` using `gh` CLI to apply the ruleset to `andy-esch/taskflow`.
- Verify the active ruleset using `gh ruleset list` and `gh ruleset view`.
- Ensure repository lint, docs-check, and test suite pass.

## Acceptance Criteria

- [ ] `.github/ruleset.json` is committed to the repository with `deletion`, `non_fast_forward`, and `Go Quality` required status checks.
- [ ] `just ruleset-apply` is added to `justfile` and successfully syncs rulesets via `gh api`.
- [ ] `gh ruleset list -R andy-esch/taskflow` displays active ruleset `main`.
- [ ] `gh ruleset view main -R andy-esch/taskflow` confirms `deletion`, `non_fast_forward`, and `"Go Quality"` status checks are active.
- [ ] `just docs-check`, `tskflwctl lint`, and `go test -race ./...` pass cleanly.

## Verification Commands

```bash
just ruleset-apply
gh ruleset list -R andy-esch/taskflow
gh ruleset view main -R andy-esch/taskflow
gh ruleset check main -R andy-esch/taskflow
go run ./cmd/tskflwctl lint
```

## Sweep refresh (2026-10-11)

2026-10-11: automated weekly sweep — the ruleset spec's required status check no longer
exists; applying this task as written would make `main` permanently unmergeable. Body and
acceptance criteria left intact; the correction is a recommendation, not an edit.

> ⚠️ **The `Go Quality` job is gone.** Everything in "Background & Comparative Context"
> and the "Ruleset Specification" block below rests on the claim that
> `.github/workflows/ci.yml` "executes a single comprehensive job named **`Go Quality`**".
> That was true when this task was filed (2026-08-30). It stopped being true on
> 2026-09-26, and the task was never updated.

**What CI actually runs today.** `.github/workflows/ci.yml` declares three jobs:

| job id | `name:` (the status-check context) |
| :--- | :--- |
| `lint` | `Lint` |
| `test` | `Tests` |
| `planning-lint` | `Planning Lint` |

The split happened in two commits on 2026-09-26: `bd375d5`
("chore(ci): split tests into separate job in workflow") removed the single `Go Quality`
job, and `835f91b` ("ci: dogfood planning lint action in workflow") added the third. A
later commit, `9c804a6` (2026-10-09, "build: guard Go and linter toolchain coordination"),
added a toolchain-coordination step the body's seven-step list also does not mention.

**Why this matters more than a stale sentence.** The ruleset JSON in this task requires
exactly one context, `{ "context": "Go Quality" }`, with
`strict_required_status_checks_policy: true`. No job reports that context any more, so a
required check that never arrives would block every PR into `main` indefinitely — and
`bypass_actors: []` leaves no escape hatch. **Do not apply the JSON block below verbatim.**

**Recommendation.** Replace the single `Go Quality` context with the three contexts CI
actually publishes:

```json
"required_status_checks": [
  { "context": "Lint" },
  { "context": "Tests" },
  { "context": "Planning Lint" }
]
```

Three contexts rather than one is also strictly better for the task's stated goal: the
split means a failure in any one job is independently visible to the ruleset instead of
being hidden behind one aggregate. Note that the contexts are the jobs' `name:` values,
not their ids, and that they will need revisiting whenever `ci.yml`'s job names change —
which is exactly the coupling that produced this drift. Worth deciding, when this is
picked up, whether to add a check that the ruleset's contexts still match `ci.yml`.

**Also verified (2026-10-11):**
- `.github/ruleset.json` still does not exist, and `.github/` holds only
  `workflows/ci.yml` and `workflows/release.yml` — so the whole task is still unstarted.
- No `ruleset-apply` recipe exists yet. Note the file is `Justfile` (capital J), not
  `justfile` as the body's "Automation & Justfile Recipe" snippet implies.
- `tidy-check` (`Justfile:76`) and `docs-check` (`Justfile:91`) both still exist, so the
  last acceptance criterion's commands are still valid.

**Acceptance criteria untouched.** The first and fourth criteria name `Go Quality` and are
now built on a false premise, but "stale" is not "decided" — nothing here is met and
nothing has been consciously waived, so no box was ticked or given a state. Rewriting them
is a human call.
