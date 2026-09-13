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

The image pins the release compiler independently of the minimum Go version in `go.mod`: release
tools may require a newer toolchain, and GitHub Actions uses the same newer Go release line.

Automated validation does not replace a release task's bounded CLI/TUI dogfood. After both are
recorded on one immutable candidate, tag that exact commit. The tag-triggered
[GitHub Actions workflow](../.github/workflows/release.yml) remains authoritative for publication;
verify its archives, checksums, embedded version, and installed binary before completing the release
task.
