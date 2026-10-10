# Releasing taskflow

Run the automated gate from a clean, committed candidate:

```sh
just release-validate
```

It checks supported tool versions, focused packages, the full race suite, formatting, module
tidiness, generated CLI/schema material, lint, package vulnerabilities, planning integrity, and
GoReleaser configuration. The snapshot build runs in a disposable clone, so validation leaves the
candidate's tracked files unchanged.

For the pinned headless Linux environment, use:

```sh
just release-validate-container
```

The container runner itself requires only Git and Docker (`just` dispatches it); set
`TASKFLOW_CONTAINER_ENGINE=podman` to use Podman. It clones the exact clean candidate, mounts that
clone read-only, runs as an unprivileged user, and stores reusable Go build/module caches in named
volumes outside the repository. The image
pins Go and the release tools in
[`build/release-validation/Containerfile`](../build/release-validation/Containerfile). It validates
Linux execution plus the same Darwin/Linux snapshot matrix, but is not a bit-for-bit reproducible
build claim. A devcontainer may wrap this image later; it must not become a second release gate.

## Go and linter version policy

`go.mod` owns the minimum supported Go version; the CI test job tracks its newest patch.
The release workflow owns the publication compiler line, which the CI lint job also uses
(`go-version: "1.26"`, `check-latest: true` today). The container pins an exact patch on that
release line, at least as new as the module minimum. A newer release/local compiler does not
automatically raise the supported minimum; a Go-line change must retain minimum-line test coverage.

CI selects a golangci-lint release line; the container pins a patch on that linter line. Go and
linter version numbers are independent. The release preflight checks the actual linter binary's
build Go against the module minimum and active compiler line, not its exact patch; successful
lint remains necessary to prove compatibility.
When changing Go lines, select a supporting linter and run both host and container qualification.
Do not infer support for a future Go release from a linter's version number alone.

```sh
just toolchain-check
```

This offline drift check runs in CI and the shared release gate. It checks the module, CI/release
Go selectors, latest-patch behavior, and container/linter pins. It does **not** check remote patch
freshness or replace `govulncheck`; `just vulncheck` uses the same package scan as the release gate.

## Tagging and publication

Automated validation does not replace a release task's bounded CLI/TUI dogfood. After both are
recorded on one immutable candidate, tag that exact commit. The tag-triggered
[GitHub Actions workflow](../.github/workflows/release.yml) remains authoritative for publication;
verify its archives, checksums, embedded version, and installed binary before completing the release
task.
