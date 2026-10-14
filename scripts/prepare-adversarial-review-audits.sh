#!/usr/bin/env bash

set -euo pipefail

usage() {
	cat <<'EOF'
Create reviewer-specific audit entities from one evidence-rich implementation-review brief.

Usage:
  scripts/prepare-adversarial-review-audits.sh \
    --area <slug> --title <title> --brief-file <path> \
    [--reviewer <slug>]... [--date YYYY-MM-DD] [-C <planning-root>] [--dry-run]

The brief must begin with review sections rather than frontmatter or an H1. It must contain:
  Review brief, Review target, Intended contract to challenge,
  Mandatory evidence floor, Required hostile angles,
  Validation and restoration, Deliverable, and Reviewer report.

When no --reviewer is supplied, the default pair is claude and antigravity.
Every generated audit requires the reviewer to use `isolated-review-workspace.sh`
before inspecting or mutating the source checkout.
Set TSKFLWCTL to an alternate binary path when needed.
EOF
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 2
}

area=""
title=""
brief_file=""
audit_date="$(date +%F)"
planning_root="."
dry_run=false
reviewers=()

while (($# > 0)); do
	case "$1" in
	--area)
		(($# >= 2)) || die "--area needs a value"
		area="$2"
		shift 2
		;;
	--title)
		(($# >= 2)) || die "--title needs a value"
		title="$2"
		shift 2
		;;
	--brief-file)
		(($# >= 2)) || die "--brief-file needs a value"
		brief_file="$2"
		shift 2
		;;
	--reviewer)
		(($# >= 2)) || die "--reviewer needs a value"
		reviewers+=("$2")
		shift 2
		;;
	--date)
		(($# >= 2)) || die "--date needs a value"
		audit_date="$2"
		shift 2
		;;
	-C | --chdir)
		(($# >= 2)) || die "$1 needs a value"
		planning_root="$2"
		shift 2
		;;
	--dry-run)
		dry_run=true
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*) die "unknown argument: $1" ;;
	esac
done

[[ -n "$area" ]] || die "--area is required"
[[ "$area" =~ ^[a-z0-9][a-z0-9-]*$ ]] || die "--area must be a lowercase slug"
[[ -n "$title" ]] || die "--title is required"
[[ -n "$brief_file" ]] || die "--brief-file is required"
[[ -f "$brief_file" ]] || die "brief file does not exist: $brief_file"
[[ "$audit_date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || die "--date must be YYYY-MM-DD"

if ((${#reviewers[@]} == 0)); then
	reviewers=(claude antigravity)
fi

seen_reviewers=()
for reviewer in "${reviewers[@]}"; do
	[[ "$reviewer" =~ ^[a-z0-9][a-z0-9-]*$ ]] || die "reviewer must be a lowercase slug: $reviewer"
	for seen in "${seen_reviewers[@]:-}"; do
		[[ "$reviewer" != "$seen" ]] || die "duplicate reviewer: $reviewer"
	done
	seen_reviewers+=("$reviewer")
done

required_sections=(
	"## Review brief"
	"## Review target"
	"## Intended contract to challenge"
	"## Mandatory evidence floor"
	"## Required hostile angles"
	"## Validation and restoration"
	"## Deliverable"
	"## Reviewer report"
)
for section in "${required_sections[@]}"; do
	grep -Fqx "$section" "$brief_file" || die "brief is missing required section: $section"
done
grep -Eqi 'consumer inventory' "$brief_file" || die "brief must require a consumer inventory"
grep -Eqi 'mutation' "$brief_file" || die "brief must require mutation evidence"

if [[ -n "${TSKFLWCTL:-}" ]]; then
	cli=("$TSKFLWCTL")
elif [[ -x "./bin/tskflwctl" ]]; then
	cli=("./bin/tskflwctl")
else
	cli=(tskflwctl)
fi
command -v "${cli[0]}" >/dev/null 2>&1 || die "tskflwctl binary not found: ${cli[0]}"

scratch="$(mktemp -d "${TMPDIR:-/tmp}/taskflow-adversarial-review.XXXXXX")"
trap 'rm -rf "$scratch"' EXIT

body_files=()
audit_slugs=()
for reviewer in "${reviewers[@]}"; do
	audit_area="${area}-${reviewer}"
	audit_slug="${audit_date}-${audit_area}"
	body_file="${scratch}/${reviewer}.md"
	{
		printf '# Audit: %s — %s — %s\n\n' "$title" "$reviewer" "$audit_date"
		printf '> Reviewer assignment: %s. This document is the review brief and the only file the reviewer should update.\n>\n' "$reviewer"
		# The regex and Markdown code span are intentional literals.
		# shellcheck disable=SC2016
		printf '> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens or em dash in place of the period. Status may be inline after `·` or on its own `**Status:** open` line within that finding section.\n\n'
		printf '> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.\n\n'
		printf '> Review-effectiveness floor: for the critical regressions selected in the brief, execute a compiler-valid mutation and require the intended test to fail; exercise changed optional wire branches with non-default values in semantic validators; run changed repair advice against the state that recommends it; and use coordinated mutations when a nearby caller would otherwise preserve an invariant accidentally. A compile failure is an invalid probe, not a killed mutation. Apply the brief\047s bounded evidence floor rather than exhaustively mutating every new test.\n\n'
		printf '> Evidence-integrity floor: verify every named symbol, field, file, lock path, test, fixture, and shipped capability in the sandbox, citing an exact path/line or command result. Distinguish implemented code and committed fixtures from requirements that exist only in planning documents. A ready/no-findings verdict is inadmissible if its inventory contains unverified names or treats planned evidence as already shipped; omit uncertain claims or label them unknown.\n\n'
		cat <<'EOF'
> Shared-worktree isolation is mandatory. Treat the checkout named in the handoff as a read-only
> source. Before inspecting implementation, running tests or generators, or making mutation probes,
> create the independent workspace below. Do not use `git worktree`, a symlink, or any arrangement
> whose `.git` metadata points back to the shared checkout. The general shell helper owns isolation,
> the baseline, verification, and the guarded one-file transfer.

## Mandatory reviewer sandbox

The implementation owner and another reviewer may be using the handoff checkout concurrently.
Reading this brief and performing the initial copy are the only operations allowed there until the
final guarded audit transfer. Your handoff prompt prints exact `SOURCE_ROOT` and `AUDIT_REL`
values. Substitute them below; never derive them from your current directory, which may be a
different checkout. Then invoke the repository's general isolated-review tool:

```sh
SOURCE_ROOT="<SOURCE_ROOT from your handoff prompt>"
AUDIT_REL="<AUDIT_REL from your handoff prompt>"
SANDBOX="$("$SOURCE_ROOT/scripts/isolated-review-workspace.sh" create \
  --source "$SOURCE_ROOT" \
  --deliverable "$AUDIT_REL" \
  --print-path)" && cd "${SANDBOX:?create failed: stop and report the blocker}"
```

Every later block begins from `${SANDBOX:?…}`, so it cannot run in the shared checkout after a
failed `create`. Run the helper from `$SOURCE_ROOT`, not from the sandbox: you may mutate the
sandbox's copy while probing, and only the read-only source copy can be trusted to verify your work.

The helper creates an independent `--no-hardlinks` clone, overlays the current staged, unstaged,
untracked, and deleted source state, detects a changing handoff, and records the result in a
sandbox-only baseline commit. That checkpoint—not the source branch's last commit—is the restoration
baseline for probes and the only commit the reviewer may create. Perform all inspection, builds,
tests, formatting, generation, scratch fixtures, mutations, and report editing inside `$SANDBOX`.
Never commit again, switch branches, stage, restore, clean, stash, reset, or run a write-capable
project command in `$SOURCE_ROOT`. The sandbox's push URLs are disabled; never re-enable them or
push to the source path. If creation fails, report the blocker; never fall back to the shared
checkout.

Before transfer, restore every probe so only the assigned audit differs, inspect its diff, then run
the helper's fail-closed verification:

```sh
cd "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "$SANDBOX" &&
  git diff -- "$AUDIT_REL"
```

Copy that `verify` output into the report's isolation attestation: workspace path, resolved Git
directory, baseline commit, captured source blob/fingerprint, deliverable, push guard,
`verify=passed`, and `transfer=pending`. The report is final from this point, so `transfer=pending`
is the truthful value; never edit it to claim delivery. Re-run `verify` (its fields do not depend on
the report text), then transfer:

```sh
"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" verify --sandbox "${SANDBOX:?create failed: stop and report the blocker}" &&
  "$SOURCE_ROOT/scripts/isolated-review-workspace.sh" transfer --sandbox "$SANDBOX"
```

The helper refuses commits, staging, unrelated changes, a non-independent or nested `.git`, a
removed or weakened push guard, source-deliverable drift, and empty reports; it copies back only
the assigned audit through a same-directory atomic rename. A successful transfer records a
helper-owned receipt and prints it (`transfer=succeeded` with the transferred blob). A refusal exits
2 and records nothing. Exit 3 means the report was already delivered when a later step failed: run
`"$SOURCE_ROOT/scripts/isolated-review-workspace.sh" receipt --sandbox "$SANDBOX"` and put its
output, or its refusal, in your handoff. Put the printed receipt in your handoff message. Do not
copy anything else manually. Leave the workspace in place and report its path until the
implementation owner confirms delivery with that same `receipt` command. On refusal, preserve it and
report the conflict rather than resolving it in the shared checkout.

A report without the verification attestation, or a handoff without the transfer receipt, is
incomplete even if its technical findings are otherwise sound.

EOF
		cat "$brief_file"
	} >"$body_file"
	body_files+=("$body_file")
	audit_slugs+=("$audit_slug")
done

# Preflight every audit before creating any, so ordinary validation and collision
# failures do not leave a half-created reviewer pair.
for i in "${!reviewers[@]}"; do
	audit_area="${area}-${reviewers[$i]}"
	"${cli[@]}" -C "$planning_root" audit new "$audit_area" \
		--date "$audit_date" --body-file "${body_files[$i]}" --dry-run --no-input >/dev/null
done

if $dry_run; then
	printf 'preflight ok: %s\n' "${audit_slugs[*]}"
	exit 0
fi

for i in "${!reviewers[@]}"; do
	audit_area="${area}-${reviewers[$i]}"
	"${cli[@]}" -C "$planning_root" audit new "$audit_area" \
		--date "$audit_date" --body-file "${body_files[$i]}" --no-input >/dev/null
	"${cli[@]}" -C "$planning_root" audit lint "${audit_slugs[$i]}" >/dev/null
	audit_path="$("${cli[@]}" -C "$planning_root" audit path "${audit_slugs[$i]}")"
	# The sandbox source is the repository holding the audit, which may differ
	# from the reviewer's working directory or from the planning root.
	audit_dir="$(cd "$(dirname "$audit_path")" && pwd -P)" || die "cannot resolve the audit directory: $audit_path"
	source_root="$(git -C "$audit_dir" rev-parse --show-toplevel)" || die "the audit is not inside a Git repository: $audit_path"
	source_root="$(cd "$source_root" && pwd -P)"
	audit_rel="$(basename "$audit_path")"
	[[ "$audit_dir" == "$source_root" ]] || audit_rel="${audit_dir#"$source_root"/}/$audit_rel"
	printf '\n[%s]\nReview the audit assigned to you at %s.\nSOURCE_ROOT=%s\nAUDIT_REL=%s\nThe current checkout is a shared, read-only handoff: before inspecting or testing, follow the audit\047s Mandatory reviewer sandbox protocol with exactly these SOURCE_ROOT and AUDIT_REL values, and do all work in that independent clone. Complete both review passes, then copy back only the assigned audit after its origin-hash guard passes. Preserve the brief and leave every finding open for implementation-owner triage. Put the verification attestation in your report and the helper\047s printed transfer receipt in your handoff message.\n' \
		"${reviewers[$i]}" "$audit_path" "$source_root" "$audit_rel"
done

printf '\n[handoff freeze]\nDo not launch either reviewer until the implementation snapshot and both audit briefs are final for this review round. Once a reviewer starts, do not edit its audit or silently advance the review target. If either must change, cancel and relaunch that review or accept a report explicitly scoped to the captured sandbox baseline.\n'

# The Markdown code span is an intentional literal.
# shellcheck disable=SC2016
printf '\n[delivery]\nBefore triaging or cleaning up after a handoff, confirm delivery with the source checkout\047s `scripts/isolated-review-workspace.sh receipt --sandbox <sandbox>`, not the sandbox\047s copy. A report\047s transfer=pending attestation is verification evidence, not delivery.\n'
