#!/usr/bin/env bash
set -euo pipefail

fail() {
	printf 'container release validation: %s\n' "$*" >&2
	exit 1
}

cleanup() {
	case "${staging:-}" in
		"${TMPDIR:-/tmp}"/taskflow-release-container.*) rm -rf -- "$staging" ;;
		"") ;;
		*) printf 'container release validation: refusing to remove unexpected temporary path %s\n' "$staging" >&2 ;;
	esac
}

engine=${TASKFLOW_CONTAINER_ENGINE:-}
if [[ -z "$engine" ]]; then
	if command -v docker >/dev/null 2>&1; then
		engine=docker
	elif command -v podman >/dev/null 2>&1; then
		engine=podman
	else
		fail "install Docker or Podman, or set TASKFLOW_CONTAINER_ENGINE"
	fi
fi
command -v "$engine" >/dev/null 2>&1 || fail "container engine '$engine' is not installed"
command -v git >/dev/null 2>&1 || fail "Git is required to prepare the isolated candidate"

repo_root=$(git rev-parse --show-toplevel 2>/dev/null) || fail "run from a Git checkout"
state=$(git -C "$repo_root" status --porcelain --untracked-files=all)
if [[ -n "$state" ]]; then
	printf '%s\n' "$state" >&2
	fail "candidate is not clean; commit or remove the listed changes before retrying"
fi
candidate_commit=$(git -C "$repo_root" rev-parse HEAD)
staging=$(mktemp -d "${TMPDIR:-/tmp}/taskflow-release-container.XXXXXX")
trap cleanup EXIT

git clone --quiet --no-hardlinks "$repo_root" "$staging/repo"
git -C "$staging/repo" checkout --quiet --detach "$candidate_commit"

image=taskflow-release-validator:local
"$engine" build --file "$staging/repo/build/release-validation/Containerfile" --tag "$image" "$staging/repo"
"$engine" volume create taskflow-release-go-build >/dev/null
"$engine" volume create taskflow-release-go-mod >/dev/null
"$engine" run --rm --read-only \
	--tmpfs /work:rw,uid=10001,gid=10001,mode=1777 \
	--mount "type=bind,source=$staging/repo,target=/src,readonly" \
	--mount type=volume,source=taskflow-release-go-build,target=/cache/go-build \
	--mount type=volume,source=taskflow-release-go-mod,target=/cache/go-mod \
	"$image"
