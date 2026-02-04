# Research: TaskFlow's Role in Documentation

**Status**: Proposal
**Created**: 2026-01-03
**Goal**: Define how TaskFlow interacts with, generates, or enhances project documentation (`docs/`, `README.md`, etc.).

## The Spectrum of Integration

### Level 1: Passive Observer (Current Plan)
**Concept**: TaskFlow indexes `docs/` alongside `tasks/` to make Semantic Search smarter.
- **Workflow**: `taskflow search "auth architecture"` -> returns `docs/auth-design.md` AND `tasks/fix-auth-bug.md`.
- **Value**: Connects "Plan" (Task) with "Theory" (Docs).
- **Implication**: Documentation remains manual Markdown files written by humans.

### Level 2: Active Generator (The "Living Docs")
**Concept**: TaskFlow auto-generates high-level documentation from the task state.
- **Feature**: `taskflow generate roadmap` -> Reads all Epics/Tasks -> Overwrites `docs/ROADMAP.md`.
- **Feature**: `taskflow generate changelog` -> Reads completed tasks -> Generates `CHANGELOG.md`.
- **Value**: Documentation never goes stale because it is derived from the work.

### Level 3: The "Meta-Doc" Layer
**Concept**: TaskFlow *is* the documentation viewer.
- **Feature**: `taskflow doc auth` -> Renders a TUI view combining `docs/auth.md` + related open tasks + recent history.
- **Value**: "Context-aware reading". You don't just read the doc; you read the doc *in the context of current work*.

## Recommendation: Start at Level 1, Aim for Level 2

**Step 1 (MVP)**:
- TaskFlow simply indexes the `paths.implementation_root/docs` folder.
- Search results show: "Found in Documentation: ..." vs "Found in Tasks: ...".

**Step 2 (Fast Follow)**:
- Add a `generate` command.
- **Why**: Maintaining a manual `ROADMAP.md` is painful. TaskFlow knows exactly what's next. Let it write that file.

**Step 3 (Future)**:
- Context-aware TUI reader. (Cool, but not critical for V1).

## Impact on "How to Write Docs"
- **Standards**: We should encourage writing docs in a way that is "TaskFlow Friendly" (clear headers, meaningful filenames).
- **Linking**: We can implement a syntax like `[[task:123]]` in documentation that TaskFlow resolves/validates.

## Configuration
Add to `.taskflow/config.yaml`:
```yaml
docs:
  generation:
    roadmap_path: "docs/ROADMAP.md"
    changelog_path: "CHANGELOG.md"
```
