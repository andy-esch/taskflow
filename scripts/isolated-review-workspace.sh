#!/usr/bin/env bash

set -euo pipefail

readonly state_dir_name="isolated-review-workspace"

usage() {
	cat <<'EOF'
Create, verify, or transfer an isolated review workspace.

Usage:
  isolated-review-workspace.sh create --deliverable <repo-relative-file> [--source <repo>] [--sandbox-parent <dir>] [--print-path]
  isolated-review-workspace.sh verify --sandbox <dir>
  isolated-review-workspace.sh transfer --sandbox <dir>

The workspace is an independent local clone overlaid with the source repository's
current staged, unstaged, untracked, and deleted state. A sandbox-only baseline
commit makes temporary review probes easy to restore. Verification permits only
one unstaged deliverable delta, and transfer copies only that file back after
checking that its source version has not changed.

The sandbox is retained after transfer for owner-confirmed cleanup.
EOF
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 2
}

canonical_dir() {
	(cd "$1" 2>/dev/null && pwd -P) || return 1
}

validate_relative_file() {
	local value="$1"
	[[ -n "$value" ]] || die "--deliverable must not be empty"
	[[ "$value" != /* ]] || die "--deliverable must be repository-relative"
	case "$value" in
	.|..|./*|../*|*/./*|*/../*|*/.|*/..)
		die "--deliverable must not contain dot or parent traversal segments: $value"
		;;
	esac
	[[ "$value" != *$'\n'* && "$value" != *$'\r'* ]] || die "--deliverable must not contain line breaks"
}

validate_target() {
	local root="$1"
	local relative="$2"
	local target="$root/$relative"
	[[ -f "$target" && ! -L "$target" ]] || die "deliverable must be a regular non-symlink file: $target"
	local parent
	parent="$(canonical_dir "$(dirname "$target")")" || die "cannot resolve deliverable parent: $target"
	[[ "$parent" == "$root" || "$parent" == "$root/"* ]] || die "deliverable resolves outside repository: $target"
}

# HEAD plus the net tracked diff and non-ignored untracked content. Comparing it
# before/after the overlay catches an owner changing the handoff mid-copy.
source_fingerprint() {
	local root="$1"
	local record relative
	record="$(mktemp "${TMPDIR:-/tmp}/isolated-review-fingerprint.XXXXXX")"
	{
		git -C "$root" rev-parse HEAD
		git -C "$root" --no-pager diff --binary --full-index --no-ext-diff HEAD --
		while IFS= read -r -d '' relative; do
			printf 'untracked\0%s\0' "$relative"
			git -C "$root" hash-object -- "$relative"
		done < <(git -C "$root" ls-files --others --exclude-standard -z)
	} >"$record"
	git hash-object "$record"
	rm -f "$record"
}

write_state() {
	printf '%s\n' "$3" >"$1/$2"
}

read_state() {
	local file="$1/$2"
	local value=""
	[[ -f "$file" && ! -L "$file" ]] || die "workspace state is missing $2"
	IFS= read -r value <"$file" || [[ -n "$value" ]] || die "workspace state $2 is empty"
	printf '%s' "$value"
}

assert_independent_clone() {
	local sandbox="$1"
	local git_dir="$2"
	[[ "$git_dir" == "$sandbox/.git" && -d "$git_dir" && ! -L "$git_dir" ]] ||
		die "workspace Git directory is not an independent in-tree .git directory: $git_dir"
	[[ ! -s "$git_dir/objects/info/alternates" ]] || die "workspace Git objects use an alternates file"
	local count=0
	local observed=""
	while IFS= read -r line; do
		case "$line" in
		worktree\ *)
			count=$((count + 1))
			observed="${line#worktree }"
			;;
		esac
	done < <(git -C "$sandbox" worktree list --porcelain)
	[[ "$count" -eq 1 && "$(canonical_dir "$observed")" == "$sandbox" ]] ||
		die "workspace Git metadata does not identify exactly this one checkout"
	[[ -z "$(git -C "$sandbox" config --get core.worktree || true)" ]] || die "workspace has core.worktree configured"
}

create_workspace() {
	local source_arg="."
	local deliverable=""
	local parent="${TMPDIR:-/tmp}"
	local print_path=false
	while (($# > 0)); do
		case "$1" in
		--source) (($# >= 2)) || die "--source needs a value"; source_arg="$2"; shift 2 ;;
		--deliverable) (($# >= 2)) || die "--deliverable needs a value"; deliverable="$2"; shift 2 ;;
		--sandbox-parent) (($# >= 2)) || die "--sandbox-parent needs a value"; parent="$2"; shift 2 ;;
		--print-path) print_path=true; shift ;;
		*) die "unknown create argument: $1" ;;
		esac
	done
	validate_relative_file "$deliverable"

	local source_root
	source_root="$(git -C "$source_arg" rev-parse --show-toplevel 2>/dev/null)" || die "--source is not inside a Git repository: $source_arg"
	source_root="$(canonical_dir "$source_root")" || die "cannot resolve source repository"
	[[ "$source_root" != *$'\n'* && "$source_root" != *$'\r'* ]] || die "source path must not contain line breaks"
	validate_target "$source_root" "$deliverable"
	parent="$(canonical_dir "$parent")" || die "sandbox parent does not exist: $parent"
	[[ -w "$parent" ]] || die "sandbox parent is not writable: $parent"
	[[ "$parent" != "$source_root" && "$parent" != "$source_root/"* ]] || die "sandbox parent must be outside the source repository"

	local source_blob before sandbox after sandbox_state baseline git_dir state_dir
	source_blob="$(git -C "$source_root" hash-object -- "$deliverable")"
	before="$(source_fingerprint "$source_root")"
	sandbox="$(mktemp -d "$parent/isolated-review.XXXXXX")"
	trap 'if [[ -n "${sandbox:-}" && -d "$sandbox" ]]; then rm -rf "$sandbox"; fi' EXIT
	git clone --quiet --no-hardlinks "$source_root" "$sandbox"
	rsync -a --delete --exclude='/.git/' "$source_root/" "$sandbox/"
	after="$(source_fingerprint "$source_root")"
	sandbox_state="$(source_fingerprint "$sandbox")"
	[[ "$before" == "$after" ]] || die "source working state changed during the copy; freeze the handoff and retry"
	[[ "$after" == "$sandbox_state" ]] || die "workspace overlay does not match the source working state"

	git -C "$sandbox" add -A
	git -C "$sandbox" \
		-c user.name='Isolated Review Workspace' \
		-c user.email='isolated-review@invalid' \
		-c commit.gpgsign=false \
		-c core.hooksPath=/dev/null \
		commit --quiet --allow-empty --no-verify -m 'chore: capture isolated review baseline'
	baseline="$(git -C "$sandbox" rev-parse HEAD)"
	git_dir="$(git -C "$sandbox" rev-parse --absolute-git-dir)"
	assert_independent_clone "$sandbox" "$git_dir"
	state_dir="$git_dir/$state_dir_name"
	(umask 077 && mkdir -p "$state_dir")
	write_state "$state_dir" source-root "$source_root"
	write_state "$state_dir" deliverable "$deliverable"
	write_state "$state_dir" source-blob "$source_blob"
	write_state "$state_dir" source-fingerprint "$after"
	write_state "$state_dir" baseline-commit "$baseline"
	trap - EXIT

	if $print_path; then
		printf '%s\n' "$sandbox"
		return
	fi
	printf 'sandbox_path=%s\ngit_dir=%s\nbaseline_commit=%s\n' "$sandbox" "$git_dir" "$baseline"
	printf 'deliverable=%s\nsource_blob=%s\nsource_fingerprint=%s\n' "$deliverable" "$source_blob" "$after"
}

# Sets globals for the verification/transfer attestation.
load_workspace() {
	local requested="$1"
	sandbox_root="$(canonical_dir "$requested")" || die "workspace does not exist: $requested"
	git_dir="$(git -C "$sandbox_root" rev-parse --absolute-git-dir 2>/dev/null)" || die "workspace is not a Git repository"
	assert_independent_clone "$sandbox_root" "$git_dir"
	state_dir="$git_dir/$state_dir_name"
	[[ -d "$state_dir" && ! -L "$state_dir" ]] || die "isolated-review state is missing"
	source_root="$(read_state "$state_dir" source-root)"
	deliverable="$(read_state "$state_dir" deliverable)"
	source_blob="$(read_state "$state_dir" source-blob)"
	source_snapshot="$(read_state "$state_dir" source-fingerprint)"
	baseline="$(read_state "$state_dir" baseline-commit)"
	validate_relative_file "$deliverable"
	[[ "$(canonical_dir "$source_root")" == "$source_root" ]] || die "recorded source repository no longer resolves"
	[[ "$source_root" != "$sandbox_root" ]] || die "workspace and source repository are the same directory"
	[[ "$(canonical_dir "$(git -C "$source_root" rev-parse --show-toplevel 2>/dev/null)")" == "$source_root" ]] ||
		die "recorded source is no longer a Git repository"
	validate_target "$source_root" "$deliverable"
	validate_target "$sandbox_root" "$deliverable"
	[[ -s "$sandbox_root/$deliverable" ]] || die "workspace deliverable must not be empty: $deliverable"
	[[ "$(git -C "$sandbox_root" rev-parse HEAD)" == "$baseline" ]] || die "workspace HEAD changed; reviewer commits are not permitted"
	git -C "$sandbox_root" diff --cached --quiet --exit-code -- || die "workspace has staged changes"

	local dirty=0 invalid="" entry status relative
	while IFS= read -r -d '' entry; do
		status="${entry:0:2}"
		relative="${entry:3}"
		dirty=$((dirty + 1))
		if [[ "$relative" != "$deliverable" || "$status" == *R* || "$status" == *C* ]]; then
			invalid="${invalid}${status} ${relative}"$'\n'
		fi
	done < <(git -C "$sandbox_root" status --porcelain=v1 -z --untracked-files=all)
	[[ -z "$invalid" && "$dirty" -le 1 ]] || die $'workspace has changes outside the deliverable:\n'"$invalid"
	git -C "$sandbox_root" diff --check -- "$deliverable"
	[[ "$(git -C "$source_root" hash-object -- "$deliverable")" == "$source_blob" ]] ||
		die "source deliverable changed after workspace creation; preserve the workspace"
	deliverable_changed=false
	git -C "$sandbox_root" diff --quiet --exit-code -- "$deliverable" || deliverable_changed=true
}

attest() {
	printf 'sandbox_path=%s\ngit_dir=%s\nbaseline_commit=%s\n' "$sandbox_root" "$git_dir" "$baseline"
	printf 'source_blob=%s\nsource_fingerprint=%s\ndeliverable=%s\n' "$source_blob" "$source_snapshot" "$deliverable"
	printf 'deliverable_changed=%s\ntransfer=%s\n' "$deliverable_changed" "$1"
}

parse_sandbox_arg() {
	sandbox_arg=""
	while (($# > 0)); do
		case "$1" in
		--sandbox) (($# >= 2)) || die "--sandbox needs a value"; sandbox_arg="$2"; shift 2 ;;
		*) die "unknown argument: $1" ;;
		esac
	done
	[[ -n "$sandbox_arg" ]] || die "--sandbox is required"
}

verify_workspace() {
	parse_sandbox_arg "$@"
	load_workspace "$sandbox_arg"
	attest pending
}

transfer_deliverable() {
	parse_sandbox_arg "$@"
	load_workspace "$sandbox_arg"
	$deliverable_changed || die "deliverable has no review delta to transfer"
	local source_target="$source_root/$deliverable"
	local sandbox_target="$sandbox_root/$deliverable"
	local transfer_file
	transfer_file="$(mktemp "${source_target}.review-transfer.XXXXXX")"
	trap 'if [[ -n "${transfer_file:-}" && -f "$transfer_file" ]]; then rm -f "$transfer_file"; fi' EXIT
	cp -p "$sandbox_target" "$transfer_file"
	[[ "$(git -C "$source_root" hash-object -- "$deliverable")" == "$source_blob" ]] ||
		die "source deliverable changed before transfer; preserve the workspace"
	mv "$transfer_file" "$source_target"
	transfer_file=""
	cmp -s "$sandbox_target" "$source_target" || die "transferred deliverable does not match the workspace"
	trap - EXIT
	attest succeeded
	printf 'source_deliverable=%s\nsandbox_retained=%s\n' "$source_target" "$sandbox_root"
}

main() {
	command -v git >/dev/null 2>&1 || die "git is required"
	command -v rsync >/dev/null 2>&1 || die "rsync is required"
	(($# > 0)) || { usage; exit 2; }
	local command="$1"
	shift
	case "$command" in
	create) create_workspace "$@" ;;
	verify) verify_workspace "$@" ;;
	transfer) transfer_deliverable "$@" ;;
	-h | --help | help) usage ;;
	*) die "unknown command: $command" ;;
	esac
}

main "$@"
