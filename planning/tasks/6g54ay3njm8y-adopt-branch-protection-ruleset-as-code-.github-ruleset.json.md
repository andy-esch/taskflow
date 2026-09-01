---
schema: 1
id: 6g54ay3njm8y
status: ready-to-start
epic: 21-code-quality-architecture-hardening
description: Configure repository ruleset as code for main branch protection, enforcing Go Quality status check and preventing deletions/force-pushes
effort: 1 day
tier: 3
priority: medium
autonomy_level: 3
tags: [ci, github, security, infrastructure, quality]
created: "2026-08-30"
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
