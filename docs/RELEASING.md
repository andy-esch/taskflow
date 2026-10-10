# Releasing taskflow

Run the automated gate from a clean, committed candidate:

```sh
just release-validate
```

The gate checks tool versions, tests/races, formatting, module tidiness, generated docs/schema,
lint, vulnerabilities, planning, and GoReleaser. Snapshot builds use a disposable clone.

For the pinned headless Linux environment, use:

```sh
just release-validate-container
```

Requires Git and Docker; use `TASKFLOW_CONTAINER_ENGINE=podman` for Podman. The runner clones
the candidate, mounts it read-only, and runs unprivileged with reusable external caches.
The [Containerfile](../build/release-validation/Containerfile) pins Go and release tools;
validation covers Linux execution and Darwin/Linux snapshots, not bit-for-bit reproducibility.

## Go and linter version policy

- `go.mod` owns the supported minimum; CI tests its latest patch. Release/CI lint use the
  publication line; the container pins a patch on it. Newer compilers need not raise the minimum.
- CI/release/container use `GOTOOLCHAIN=local` to prevent automatic compiler replacement.
- CI and container linters share a release line; exact stable CI pins are allowed. The linter
  must be built with Go covering the minimum and active compiler line. Qualification requires
  stable toolchains and real lint runs, not matching version numbers.

```sh
just toolchain-check
just renovate-toolchain-check /path/to/node_modules/renovate  # optional, when editing rules
```

The drift check runs in CI and release validation; it does not certify patch freshness.
Renovate gates minimum/line upgrades behind approval; compiler patches bypass schedule/age
delays but still require CI and human merge. Validate rule changes with Renovate's config validator.
The [coordination task](../planning/tasks/6gd68x5gqkk0-reconcile-go-toolchain-versioning-across-go.mod-ci-linters-and-renovate.md)
records detailed policy, limitations, and review evidence.

## Tagging and publication

Record automated validation and bounded CLI/TUI dogfood on one candidate, then tag that commit.
The [release workflow](../.github/workflows/release.yml) publishes it. Verify archives, checksums,
embedded version, and the installed binary before completing the release task.
