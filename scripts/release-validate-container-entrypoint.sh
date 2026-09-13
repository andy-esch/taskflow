#!/usr/bin/env bash
set -euo pipefail

[[ -d /src/.git ]] || {
	printf 'release validator: /src must be a standalone Git clone\n' >&2
	exit 1
}

checkout=$(mktemp -d /work/taskflow-release.XXXXXX)
cp -R /src/. "$checkout"
cd "$checkout"
exec ./scripts/release-validate.sh
