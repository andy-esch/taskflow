#!/usr/bin/env bash
set -euo pipefail

fail() {
	printf 'release validation: %s\n' "$*" >&2
	exit 1
}

require_command() {
	command -v "$1" >/dev/null 2>&1 || fail "missing required command '$1' ($2)"
}

version_at_least() {
	local actual=${1#go}
	local required=${2#go}
	local actual_major actual_minor actual_patch required_major required_minor required_patch
	IFS=. read -r actual_major actual_minor actual_patch <<<"$actual"
	IFS=. read -r required_major required_minor required_patch <<<"$required"
	actual_patch=${actual_patch%%[^0-9]*}
	required_patch=${required_patch%%[^0-9]*}
	actual_patch=${actual_patch:-0}
	required_patch=${required_patch:-0}
	(( actual_major > required_major )) ||
		(( actual_major == required_major && actual_minor > required_minor )) ||
		(( actual_major == required_major && actual_minor == required_minor && actual_patch >= required_patch ))
}

run_phase() {
	local label=$1
	shift
	printf '\n==> %s\n' "$label"
	if ! "$@"; then
		fail "phase failed: $label"
	fi
}

check_clean() {
	local state
	state=$(git status --porcelain --untracked-files=all)
	if [[ -n "$state" ]]; then
		printf '%s\n' "$state" >&2
		fail "candidate is not clean; commit or remove the listed changes before retrying"
	fi
}

check_tools() {
	local required_go actual_go lint_version release_version
	required_go=$(awk '$1 == "go" { print $2; exit }' go.mod)
	[[ -n "$required_go" ]] || fail "go.mod does not declare a Go version"
	actual_go=$(go env GOVERSION)
	version_at_least "$actual_go" "$required_go" ||
		fail "Go $required_go or newer is required; found ${actual_go#go}"

	lint_version=$(golangci-lint version 2>&1)
	[[ "$lint_version" =~ version[[:space:]]+2\. ]] ||
		fail "golangci-lint v2 is required; got: ${lint_version%%$'\n'*}"
	if [[ "$lint_version" =~ built[[:space:]]+with[[:space:]]+go([0-9]+\.[0-9]+(\.[0-9]+)?) ]]; then
		version_at_least "${BASH_REMATCH[1]}" "$required_go" ||
			fail "golangci-lint must be built with Go $required_go or newer"
	fi

	release_version=$(goreleaser --version 2>&1)
	[[ "$release_version" =~ GitVersion:[[:space:]]+v?2\. ]] ||
		fail "GoReleaser v2 is required"
}

check_formatting() {
	local files
	files=$(gofmt -l .)
	if [[ -n "$files" ]]; then
		printf 'files requiring gofmt:\n%s\n' "$files" >&2
		return 1
	fi
}

check_generated_docs() {
	local output=$validation_tmp/docs
	go run ./internal/tools/docgen -out "$output"
	diff -ruN docs/cli "$output"
}

check_generated_schema_comments() {
	local output=$validation_tmp/schema_comments.json
	go run ./internal/tools/schemacomments -out "$output"
	diff -u internal/wire/schema_comments.json "$output"
}

check_planning() {
	go run ./cmd/tskflwctl --no-color lint
}

build_snapshot_in_clone() {
	local snapshot=$validation_tmp/snapshot
	git clone --quiet --no-hardlinks "$repo_root" "$snapshot"
	git -C "$snapshot" checkout --quiet --detach "$candidate_commit"
	(
		cd "$snapshot"
		goreleaser release --snapshot --clean
	)
}

cleanup() {
	case "${validation_tmp:-}" in
		"$release_tmp_root"/taskflow-release-validate.*) rm -rf -- "$validation_tmp" ;;
		"") ;;
		*) printf 'release validation: refusing to remove unexpected temporary path %s\n' "$validation_tmp" >&2 ;;
	esac
}

for dependency in git go gofmt golangci-lint govulncheck goreleaser awk diff mktemp; do
	case "$dependency" in
		git) remedy="install Git" ;;
		go|gofmt) remedy="install the Go toolchain declared by go.mod" ;;
		golangci-lint) remedy="install golangci-lint v2" ;;
		govulncheck) remedy="go install golang.org/x/vuln/cmd/govulncheck@latest" ;;
		goreleaser) remedy="install GoReleaser v2" ;;
		*) remedy="install the standard command-line utility" ;;
	esac
	require_command "$dependency" "$remedy"
done

repo_root=$(git rev-parse --show-toplevel 2>/dev/null) || fail "run from a Git checkout"
cd "$repo_root"
check_clean
candidate_commit=$(git rev-parse HEAD)
release_tmp_root=${TASKFLOW_RELEASE_TMP_ROOT:-${TMPDIR:-/tmp}}
validation_tmp=$(mktemp -d "$release_tmp_root/taskflow-release-validate.XXXXXX")
trap cleanup EXIT
export GOCACHE="$validation_tmp/go-build-cache"
export GOLANGCI_LINT_CACHE="$validation_tmp/golangci-lint-cache"
mkdir -p "$GOCACHE" "$GOLANGCI_LINT_CACHE"

run_phase "toolchain compatibility" check_tools
run_phase "focused package tests" go test ./internal/core ./internal/store ./internal/cli ./internal/tui ./internal/wire
run_phase "full race suite" go test -race ./...
run_phase "Go formatting" check_formatting
run_phase "Go module tidiness" go mod tidy -diff
run_phase "generated CLI documentation" check_generated_docs
run_phase "generated schema comments" check_generated_schema_comments
run_phase "golangci-lint" golangci-lint run ./...
run_phase "package vulnerability scan" govulncheck -scan package ./...
run_phase "planning integrity" check_planning
run_phase "GoReleaser configuration" goreleaser check
run_phase "isolated release snapshot" build_snapshot_in_clone
run_phase "candidate remained clean" check_clean

printf '\nrelease validation passed for %s\n' "$candidate_commit"
