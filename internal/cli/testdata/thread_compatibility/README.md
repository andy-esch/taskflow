# Thread preview compatibility fixtures

These files pin artifacts users could have retained from the two Thread preview releases:

- `v0.18.0` is commit `89cfb851d6e11f1530d526dddb447259561555c3`, wire schema 1.58.
- `v0.19.0` is commit `d6bf8dd8841b67ef9f220a84c5d4b8050f2378c7`, wire schema 1.60.

Each release directory contains the exact Thread document and all six Thread JSON goldens committed
at that tag. v0.18.0 and v0.19.0 emitted identical Thread documents and payload shapes; retaining
both copies pins their distinct wire-version labels as well as the shared shape. `planning/`
contains the configuration and three task documents those artifacts referenced; those files are
also byte-identical at both tags, so retaining one labelled copy avoids fake format differences.

The authoring manifests and materialized plan in `artifacts/` reconstruct the exact public structs
and YAML field names present at both tags. The preview dogfood plan files lived only in a throwaway
directory, so no original plan bytes were committed to recover. These fixtures pin the retained
schema-zero/schema-one behavior without claiming otherwise.

Tests must not invoke Git or a historical binary. The committed files are the release evidence; Git
history is provenance for a human refreshing them deliberately. To retain a later release, copy its
Thread fixture and six `internal/cli/testdata/golden/thread_*_json.golden` files into a new labelled
directory and add the release to `threadPreviewReleases`.
