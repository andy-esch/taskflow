# Research: AI File Editing Strategy (Diffs vs. Direct)

**Status**: Proposal
**Created**: 2026-01-03
**Context**: TaskFlow will have AI agents editing files (updating tasks, generating docs). We need to decide the most efficient and reliable mechanism for the AI to apply changes.

## Option 1: Direct File Overwrite
**Mechanism**: AI generates the *entire* new file content and overwrites the old one.
- **Pros**: 100% reliable (no merge conflicts). Simple for the AI.
- **Cons**: **Expensive**. If you change 1 line in a 500-line file, you pay for 500 lines of output tokens. Slow.

## Option 2: Search & Replace (Tool Use)
**Mechanism**: AI calls a tool `replace(file, old_string, new_string)`.
- **Pros**: Token efficient. Precision editing.
- **Cons**: **Brittle**. If `old_string` isn't unique or whitespace doesn't match exactly, it fails. AI often struggles with indentation or context in search strings.

## Option 3: Unified Diff (`git apply`)
**Mechanism**: AI writes a standard `.diff` or `.patch` file. System runs `git apply`.
- **Pros**:
    - **Token Efficient**: Only outputs changed lines + context.
    - **Standard**: Uses `git`'s robust merging logic (fuzzy matching context).
    - **Reviewable**: Humans can read the diff before applying.
- **Cons**: AI is notoriously bad at calculating line numbers correctly (which standard diffs often require). However, `git apply` with context (hunks) is more forgiving than strict line numbers.

## Option 4: Aider/Cursor Style "Search/Replace Block"
**Mechanism**: A simplified diff format optimized for LLMs.
```
<<<<<<< SEARCH
function foo() {
  return 1;
}
=======
function foo() {
  return 2;
}
>>>>>>> REPLACE
```
- **Pros**: AI is *very* good at this. No line numbers needed. Less brittle than strict string replacement because it matches blocks.
- **Cons**: Requires custom parsing logic (can't just use `git apply`).

## Recommendation: Option 4 (Custom "Block Diff")

**Why?**
1.  **Reliability**: Standard `git diff` requires header metadata (file paths, timestamps, line numbers) that LLMs hallucinate or mess up.
2.  **Efficiency**: Much cheaper than full rewrites.
3.  **Simplicity**: It's easier to implement a "Block Replacer" in Python/Go than to make an LLM output valid unified diffs consistently.

### Implementation in TaskFlow
- **Tool**: `update_file(path, search_block, replace_block)`
- **Logic**:
    1.  Read file.
    2.  Find `search_block` (ignoring minor whitespace diffs?).
    3.  Replace with `replace_block`.
    4.  If unique match not found, return error to AI ("Context ambiguous, please provide more surrounding lines").

**Backup**: If the file is small (< 50 lines), just overwrite it. It's safer.
