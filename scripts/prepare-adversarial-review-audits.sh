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
		printf '> Finding grammar is exact: use `#### M1. <title> · **Status:** open` (or H1/L1). Codes must match `[A-Z]+[0-9]+`; no hyphens, no em dash in place of the period, and no free-standing status line.\n\n'
		printf '> Required second pass: after completing the brief checklist, review the change again for systemic failure modes. Take an explicitly adversarial stance toward shared abstractions, test helpers that can mask broad defect classes, state changing between projection and action, and boundaries that only appear to fail closed. Prefer one demonstrated systemic issue over several speculative findings, and settle each challenged pattern with hostile evidence.\n\n'
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
	printf '\n[%s]\nReview the audit assigned to you at %s. Complete both the contract/checklist pass and the required systemic second pass, including code that may still be in flux. Replace only its Reviewer report placeholder, preserve the brief, and leave every finding open for implementation-owner triage.\n' \
		"${reviewers[$i]}" "$audit_path"
done
