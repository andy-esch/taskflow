#!/usr/bin/env bash

set -euo pipefail

readonly state_dir_name="isolated-review-workspace"
# An unknown transport scheme makes Git fail before contacting anything, and its
# error names the guard. A clone otherwise pushes to the live source checkout.
readonly push_guard_url="isolated-review-push-disabled://source-is-read-only"
# A report must be final before transfer, so it can attest only verification.
# This helper-owned file is the delivery evidence, written only after transfer.
readonly receipt_name="transfer-receipt"
# Written before the source is touched, so the receipt cannot be lost to a
# failure after the source rename; published as the receipt once confirmed.
readonly pending_name="transfer-receipt.pending"

usage() {
	cat <<'EOF'
Create, verify, transfer, or confirm delivery from an isolated review workspace.

Usage:
  isolated-review-workspace.sh create --deliverable <repo-relative-file> [--source <repo>] [--sandbox-parent <dir>] [--print-path]
  isolated-review-workspace.sh verify --sandbox <dir>
  isolated-review-workspace.sh transfer --sandbox <dir>
  isolated-review-workspace.sh receipt --sandbox <dir>

The workspace is an independent local clone overlaid with the source repository's
current staged, unstaged, untracked, and deleted state. A sandbox-only baseline
commit makes temporary review probes easy to restore. Verification permits only
one unstaged deliverable delta, and transfer copies only that file back after
checking that its source version has not changed.

A report is final before transfer, so its verify attestation says
transfer=pending. A successful transfer records a receipt in the workspace's
private state and prints it; receipt fails until then, and afterwards confirms
that the source still holds the transferred bytes. Verify and transfer refuse
once a receipt exists. Refusals exit 2 and leave the source unchanged. Exit 3
means the source report was already replaced when the failure hit; run receipt,
which records an interrupted delivery whose bytes are intact or explains why it
cannot.

Every workspace remote pushes only to a disabled URL, so an ordinary `git push`
fails before reaching the source; verify and transfer reject a removed or
weakened guard. This prevents accidents. It is not a security boundary against
a push to an explicit source path or a command-line configuration override.

Workspaces are created under /tmp unless --sandbox-parent names another
directory. A workspace is retained after transfer until its review is
processed, then it is disposable.
EOF
}

die() {
	printf 'error: %s\n' "$*" >&2
	exit 2
}

# Exit 3 means the source report had already been replaced when the failure hit;
# exit 2 is reserved for refusals that left the source unchanged.
die_delivered() {
	printf 'error: %s\n' "$*" >&2
	exit 3
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

# Reads of the source must not write it. Porcelain diff otherwise rewrites a
# stat-dirty index (for a linked worktree, the shared repository's copy).
read_git() {
	local root="$1"
	shift
	git --no-optional-locks --no-pager -C "$root" -c diff.autoRefreshIndex=false "$@"
}

# HEAD plus the net tracked diff and non-ignored untracked content. Comparing it
# before/after the overlay catches an owner changing the handoff mid-copy. Each
# step is checked explicitly: bash 3.2 suspends errexit inside $(...), and an
# unhashable path such as a nested repository must fail rather than be skipped.
source_fingerprint() {
	local root="$1"
	local record untracked relative ok=true
	record="$(mktemp "${TMPDIR:-/tmp}/isolated-review-fingerprint.XXXXXX")" || return 1
	untracked="$record.untracked"
	if read_git "$root" rev-parse HEAD >"$record" &&
		read_git "$root" diff --binary --full-index --no-ext-diff HEAD -- >>"$record" &&
		read_git "$root" ls-files --others --exclude-standard -z >"$untracked"; then
		while IFS= read -r -d '' relative; do
			printf 'untracked\0%s\0' "$relative" >>"$record"
			# A symlink is content by its link text, as git add stores it; hashing
			# would follow it and fail on a dangling or directory target.
			if [[ -L "$root/$relative" ]]; then
				readlink "$root/$relative" >>"$record" || ok=false
			else
				read_git "$root" hash-object -- "$relative" >>"$record" || ok=false
			fi
			$ok || break
		done <"$untracked"
	else
		ok=false
	fi
	if $ok; then
		git hash-object "$record" || ok=false
	fi
	rm -f "$record" "$untracked"
	$ok
}

# Succeeds when a symlink's target, resolved lexically from the link's own
# directory, is absolute or climbs above the workspace root.
link_escapes() {
	local entry="$1"
	local target="$2"
	local depth part
	local -a parts
	[[ "$target" != /* ]] || return 0
	IFS=/ read -r -a parts <<<"$entry"
	depth=$((${#parts[@]} - 1))
	IFS=/ read -r -a parts <<<"$target"
	for part in "${parts[@]}"; do
		case "$part" in
		'' | .) ;;
		..) depth=$((depth - 1)) && ((depth >= 0)) || return 0 ;;
		*) depth=$((depth + 1)) ;;
		esac
	done
	return 1
}

# Only the workspace's own root .git may bind it to a repository. A nested
# .git in any letter case (case-insensitive filesystems read .GIT as a gitfile)
# or a symlink leading out of the workspace would let Git commands or writes
# reach another repository while the root still looks independent. The scan
# runs from inside the workspace, so path characters are never patterns.
refuse_external_bindings() {
	local sandbox="$1"
	local entry target
	while IFS= read -r -d '' entry; do
		entry="${entry#./}"
		case "${entry##*/}" in
		[.][gG][iI][tT])
			die "workspace contains nested Git metadata: $sandbox/$entry; it can act on another repository, so move it outside the workspace"
			;;
		esac
		target="$(readlink "$sandbox/$entry")" || die "cannot read workspace symlink: $entry"
		! link_escapes "$entry" "$target" ||
			die "workspace symlink $entry points outside the workspace ($target), so Git commands or writes through it reach other files; remove it or make it relative within the repository"
	done < <(cd "$sandbox" && find . -path ./.git -prune -o \( -iname .git -o -type l \) -print0)
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

# One command for the reviewer to run, quoted with printf %q so a name with
# spaces, quotes, or glob characters stays a single argument in bash or zsh.
restore_line() {
	local word line="restore:"
	for word in "$@"; do
		line="$line $(printf '%q' "$word")"
	done
	printf '%s' "$line"
}

# Workspace-scoped settings for one key pattern, as NUL-delimited origin and
# key/value records: the local file, its include/includeIf files, and
# config.worktree. User-wide and system scopes are documented limits.
workspace_config() {
	local sandbox="$1"
	local pattern="$2"
	local scope
	for scope in --local --worktree; do
		git -C "$sandbox" config -z --show-origin --includes "$scope" --get-regexp "$pattern" 2>/dev/null || true
	done
}

origin_file() {
	local path="${2#file:}"
	[[ "$path" == /* ]] || path="$1/$path"
	printf '%s' "$path"
}

# Every remote must push only to the disabled URL after Git's own insteadOf
# expansion, and no workspace setting may name a push destination other than a
# guarded remote. This catches accidental reconfiguration; explicit-URL pushes
# and user-wide or command-line overrides are outside a workspace check.
check_push_guard() {
	local sandbox="$1"
	local source_root="$2"
	local remotes remote urls origin entry key value rule rules file legacy
	remotes="$(git -C "$sandbox" remote)"
	[[ $'\n'"$remotes"$'\n' == *$'\n'origin$'\n'* ]] ||
		die "push guard missing: remote 'origin' was removed"$'\n'"$(restore_line git -C "$sandbox" remote add origin "$source_root")"$'\n'"$(restore_line git -C "$sandbox" remote set-url --push origin "$push_guard_url")"
	# A conditional include can activate later, for example on a branch that does
	# not exist yet, so the push settings it would bring cannot be checked now.
	while IFS= read -r -d '' origin && IFS= read -r -d '' entry; do
		key="${entry%%$'\n'*}"
		die "push guard bypassed: conditional include $key can change push settings later, for example on another branch"$'\n'"$(restore_line git config --file "$(origin_file "$sandbox" "$origin")" --unset-all "$key")"
	done < <(workspace_config "$sandbox" '^includeif\.')
	while IFS= read -r remote; do
		[[ -n "$remote" ]] || continue
		urls="$(git -C "$sandbox" remote get-url --push --all "$remote" 2>&1)" ||
			die "push guard unreadable for remote '$remote': $urls"
		[[ "$urls" != "$push_guard_url" ]] || continue
		rules=""
		# A push URL set in an included file or config.worktree survives a local
		# replace-all, so remove it from the file that sets it.
		while IFS= read -r -d '' origin && IFS= read -r -d '' entry; do
			[[ "${entry%%$'\n'*}" == "remote.$remote.pushurl" ]] || continue
			file="$(origin_file "$sandbox" "$origin")"
			[[ "$file" != "$sandbox/.git/config" ]] || continue
			rule="$(restore_line git config --file "$file" --unset-all "remote.$remote.pushurl")"
			[[ "$rules"$'\n' == *$'\n'"$rule"$'\n'* ]] || rules="$rules"$'\n'"$rule"
		done < <(workspace_config "$sandbox" '^remote\..*\.pushurl$')
		while IFS= read -r -d '' origin && IFS= read -r -d '' entry; do
			key="${entry%%$'\n'*}"
			value=""
			[[ "$entry" != *$'\n'* ]] || value="${entry#*$'\n'}"
			[[ -n "$value" && "$push_guard_url" == "$value"* ]] || continue
			# --worktree reads the local file again unless worktreeConfig is on.
			rule="$(restore_line git config --file "$(origin_file "$sandbox" "$origin")" --unset-all "$key")"
			[[ "$rules"$'\n' == *$'\n'"$rule"$'\n'* ]] || rules="$rules"$'\n'"$rule"
		done < <(workspace_config "$sandbox" '^url\..*\.insteadof$')
		die "push guard weakened: remote '$remote' would push to: $(printf '%s' "$urls" | tr '\n' ' ')"$'\n'"$(restore_line git -C "$sandbox" config --replace-all "remote.$remote.pushurl" "$push_guard_url")$rules"
	done <<<"$remotes"
	while IFS= read -r -d '' origin && IFS= read -r -d '' entry; do
		key="${entry%%$'\n'*}"
		value=""
		[[ "$entry" != *$'\n'* ]] || value="${entry#*$'\n'}"
		[[ "$value" != *$'\n'* ]] && [[ "$value" == "." || $'\n'"$remotes"$'\n' == *$'\n'"$value"$'\n'* ]] ||
			die "push guard bypassed: $key names '$value' rather than a guarded remote"$'\n'"$(restore_line git config --file "$(origin_file "$sandbox" "$origin")" --unset-all "$key")"
	done < <(workspace_config "$sandbox" '^(remote\.pushdefault|branch\..*\.(pushremote|remote))$')
	# Any non-directory entry, including a symlink, can define a legacy remote.
	while IFS= read -r -d '' legacy; do
		die "push guard bypassed: legacy remote file $legacy can name a push URL that git remote does not list"$'\n'"$(restore_line rm -- "$legacy")"
	done < <(find "$sandbox/.git/remotes" "$sandbox/.git/branches" ! -type d -print0 2>/dev/null)
}

create_workspace() {
	local source_arg="."
	local deliverable=""
	# Workspaces are disposable once their review is processed; /tmp lets the
	# operating system's temporary-file cleanup reclaim any left behind.
	local parent="/tmp"
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
	before="$(source_fingerprint "$source_root")" ||
		die "cannot fingerprint the source working state (an unhashable path such as a nested repository?)"
	sandbox="$(mktemp -d "$parent/isolated-review.XXXXXX")"
	trap 'if [[ -n "${sandbox:-}" && -d "$sandbox" ]]; then rm -rf "$sandbox"; fi' EXIT
	git clone --quiet --no-hardlinks "$source_root" "$sandbox"
	git -C "$sandbox" remote set-url --push origin "$push_guard_url"
	# Exclude every .git, file or directory, at any depth: a linked-worktree source
	# has a root .git file, and nested worktrees carry their own .git files, all
	# naming the shared repository's administration.
	rsync -a --delete --exclude='.git' "$source_root/" "$sandbox/"
	# Re-check before any workspace Git write: an overlay that replaced or nested
	# .git would otherwise stage and commit the baseline into another repository.
	refuse_external_bindings "$sandbox"
	assert_independent_clone "$sandbox" "$(git -C "$sandbox" rev-parse --absolute-git-dir)"
	after="$(source_fingerprint "$source_root")" || die "cannot fingerprint the source working state after the copy"
	sandbox_state="$(source_fingerprint "$sandbox")" || die "cannot fingerprint the workspace overlay"
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
	check_push_guard "$sandbox" "$source_root"
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
	printf 'push_guard=%s\n' "$push_guard_url"
}

# Sets globals describing the workspace and its recorded capture.
load_state() {
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
}

# Sets globals for the verification/transfer attestation.
load_workspace() {
	load_state "$1"
	[[ ! -e "$state_dir/$receipt_name" ]] ||
		die "deliverable already transferred; confirm delivery with: $0 receipt --sandbox '$sandbox_root'"
	local source_now
	source_now="$(git -C "$source_root" hash-object -- "$deliverable")"
	if [[ -e "$state_dir/$pending_name" ]]; then
		[[ "$source_now" == "$source_blob" ]] ||
			die "a transfer replaced the source report but its receipt was not recorded; resolve it with: $0 receipt --sandbox '$sandbox_root'"
		refuse_unresolved_pending
	fi
	[[ -s "$sandbox_root/$deliverable" ]] || die "workspace deliverable must not be empty: $deliverable"
	[[ "$(git -C "$sandbox_root" rev-parse HEAD)" == "$baseline" ]] || die "workspace HEAD changed; reviewer commits are not permitted"
	git -C "$sandbox_root" diff --cached --quiet --exit-code -- || die "workspace has staged changes"
	refuse_external_bindings "$sandbox_root"
	check_push_guard "$sandbox_root" "$source_root"

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
	if [[ "$source_now" != "$source_blob" ]]; then
		[[ "$source_now" != "$(git -C "$sandbox_root" hash-object -- "$deliverable")" ]] ||
			die "the source already holds this workspace's report but no receipt was recorded; inspect it with: $0 receipt --sandbox '$sandbox_root'"
		die "source deliverable changed after workspace creation; preserve the workspace"
	fi
	deliverable_changed=false
	git -C "$sandbox_root" diff --quiet --exit-code -- "$deliverable" || deliverable_changed=true
}

# A pending record with the source still unchanged means a transfer is running,
# or was killed before its rename. Neither verify nor receipt may delete it: a
# running transfer still needs it. Print the owner's cleanup for a killed one.
refuse_unresolved_pending() {
	local message staged
	message="a transfer is in progress, or one was interrupted before replacing the source report, which is unchanged; if no transfer is running, remove its leftovers and retry"$'\n'"$(restore_line rm -f -- "$state_dir/$pending_name")"
	while IFS= read -r -d '' staged; do
		message="$message"$'\n'"$(restore_line rm -f -- "$staged")"
	done < <(find "$(dirname "$source_root/$deliverable")" -maxdepth 1 -name "$(basename "$deliverable").review-transfer.*" -print0 2>/dev/null)
	die "$message"
}

# Only called after load_workspace passed, so verification has succeeded.
attest() {
	printf 'sandbox_path=%s\ngit_dir=%s\nbaseline_commit=%s\n' "$sandbox_root" "$git_dir" "$baseline"
	printf 'source_blob=%s\nsource_fingerprint=%s\ndeliverable=%s\n' "$source_blob" "$source_snapshot" "$deliverable"
	printf 'push_guard=%s\n' "$push_guard_url"
	printf 'deliverable_changed=%s\nverify=passed\ntransfer=%s\n' "$deliverable_changed" "$1"
}

# Records the receipt fields before the source is touched, so an unwritable
# state directory refuses the transfer while the source is still unchanged.
write_pending() {
	local transferred="$1"
	local source_target="$2"
	local pending="$state_dir/$pending_name"
	local temporary
	temporary="$(mktemp "$pending.XXXXXX")" ||
		die "cannot record a transfer receipt in $state_dir; the source was not changed"
	if ! {
		attest succeeded
		printf 'transferred_blob=%s\n' "$transferred"
		printf 'source_deliverable=%s\nsandbox_retained=%s\n' "$source_target" "$sandbox_root"
	} >"$temporary" || ! mv "$temporary" "$pending"; then
		rm -f "$temporary"
		die "cannot record a transfer receipt in $state_dir; the source was not changed"
	fi
}

# Publishes a complete receipt built from the pending record for this blob. It
# is linked into place without overwriting, so concurrent publishers of one
# transfer cannot replace each other, and a missing pending record fails rather
# than producing an empty receipt. Returns nonzero rather than dying: the caller
# knows whether delivery happened.
publish_receipt() {
	local transferred="$1"
	local note="$2"
	local pending="$state_dir/$pending_name"
	local receipt="$state_dir/$receipt_name"
	local temporary
	if [[ -f "$pending" && ! -L "$pending" && "$(receipt_value "$pending" transferred_blob || true)" == "$transferred" ]]; then
		temporary="$(mktemp "$receipt.XXXXXX")" || return 1
		if {
			cat "$pending" &&
				printf 'transferred_at=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" &&
				{ [[ -z "$note" ]] || printf '%s\n' "$note"; }
		} >"$temporary" && ln "$temporary" "$receipt" 2>/dev/null; then
			rm -f "$temporary" "$pending"
			return 0
		fi
		rm -f "$temporary"
	fi
	# Another publisher of this same transfer may have won the race.
	[[ -f "$receipt" && "$(receipt_value "$receipt" transferred_blob || true)" == "$transferred" ]] &&
		rm -f "$pending"
}

receipt_value() {
	local file="$1"
	local wanted="$2"
	local key value
	while IFS='=' read -r key value; do
		if [[ "$key" == "$wanted" ]]; then
			printf '%s' "$value"
			return 0
		fi
	done <"$file"
	return 1
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
	local transferred
	transferred="$(git hash-object -- "$sandbox_target")" || die "cannot hash the workspace report"
	write_pending "$transferred" "$source_target"
	# Until the rename, any failure removes the staged copy and the pending record.
	transfer_file=""
	pending_cleanup="$state_dir/$pending_name"
	trap 'rm -f "${transfer_file:-}"; if [[ -n "${pending_cleanup:-}" ]]; then rm -f "$pending_cleanup"; fi' EXIT
	transfer_file="$(mktemp "${source_target}.review-transfer.XXXXXX")" ||
		die "cannot stage the report beside the source deliverable; the source was not changed"
	cp -p "$sandbox_target" "$transfer_file" || die "cannot stage the report; the source was not changed"
	[[ "$(git hash-object -- "$transfer_file")" == "$transferred" ]] ||
		die "workspace report changed during transfer; the source was not changed"
	[[ "$(git -C "$source_root" hash-object -- "$deliverable")" == "$source_blob" ]] ||
		die "source deliverable changed before transfer; preserve the workspace"
	mv "$transfer_file" "$source_target" || die "cannot replace the source deliverable; the source was not changed"
	# From here the pending record is the only evidence of delivery: keep it.
	transfer_file=""
	pending_cleanup=""
	trap - EXIT
	[[ "$(git -C "$source_root" hash-object -- "$deliverable")" == "$transferred" ]] ||
		die_delivered "the source report was replaced but no longer holds the transferred bytes; delivery is unconfirmed: run $0 receipt --sandbox '$sandbox_root'"
	publish_receipt "$transferred" "" ||
		die_delivered "the report was delivered but its receipt was not recorded: run $0 receipt --sandbox '$sandbox_root' to record it"
	cat "$state_dir/$receipt_name"
}

# Delivery evidence for the owner: fails until a transfer has succeeded, and
# whenever the source or workspace no longer holds the transferred bytes. It
# also records an interrupted delivery whose bytes are intact in both places.
show_receipt() {
	parse_sandbox_arg "$@"
	load_state "$sandbox_arg"
	local receipt="$state_dir/$receipt_name"
	local pending="$state_dir/$pending_name"
	local transferred source_now workspace_now
	source_now="$(git -C "$source_root" hash-object -- "$deliverable")"
	workspace_now="$(git -C "$sandbox_root" hash-object -- "$deliverable")"
	if [[ ! -e "$receipt" && -f "$pending" && ! -L "$pending" ]]; then
		transferred="$(receipt_value "$pending" transferred_blob || true)"
		if [[ "$source_now" == "$source_blob" ]]; then
			refuse_unresolved_pending
		elif [[ -n "$transferred" && "$source_now" == "$transferred" && "$workspace_now" == "$transferred" ]]; then
			publish_receipt "$transferred" "recorded_by=receipt after an interrupted transfer" ||
				die "the source holds the transferred report, but its receipt could not be recorded in $state_dir"
		else
			die "delivery uncertain: a transfer replaced the source report but was not recorded, and the bytes no longer agree (transferred ${transferred:-unknown}, source now $source_now, workspace now $workspace_now)"
		fi
	fi
	if [[ ! -f "$receipt" || -L "$receipt" ]]; then
		[[ "$source_now" == "$source_blob" || "$source_now" != "$workspace_now" ]] ||
			die "no transfer receipt recorded, although the source holds this workspace's report; an older helper or a hand copy delivered it, and this helper cannot vouch for it"
		die "no transfer receipt: the guarded transfer has not succeeded for this workspace; a report's transfer=pending attestation is verification evidence, not delivery"
	fi
	[[ "$(receipt_value "$receipt" transfer || true)" == succeeded ]] ||
		die "transfer receipt does not record a successful transfer: $receipt"
	transferred="$(receipt_value "$receipt" transferred_blob || true)"
	[[ -n "$transferred" ]] || die "transfer receipt has no transferred_blob: $receipt"
	[[ "$source_now" == "$transferred" ]] ||
		die "source deliverable no longer matches the transferred report (transferred $transferred, source now $source_now); the receipt records only the earlier transfer"
	[[ "$workspace_now" == "$transferred" ]] ||
		die "workspace deliverable changed after transfer (transferred $transferred, workspace now $workspace_now)"
	cat "$receipt"
	printf 'delivery=confirmed\n'
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
	receipt) show_receipt "$@" ;;
	-h | --help | help) usage ;;
	*) die "unknown command: $command" ;;
	esac
}

main "$@"
