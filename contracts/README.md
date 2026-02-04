# Data Contracts

This directory contains the **Single Source of Truth** for the TaskFlow data model.

## 🏗 Structure

- `proto/`: The source definitions (Protocol Buffers).
    - `taskflow/v1/`: Versioned schema definitions.
- `gen/`: Auto-generated code artifacts. **Do not edit these manually.**
    - `go/`: Go structs for the CLI.
    - `python/`: Python Pydantic models for the Semantic Engine.
    - `jsonschema/`: JSON Schemas for validating Markdown frontmatter.

## 🛠 Workflow

To update the data model:

1.  Edit the `.proto` files in `proto/taskflow/v1/`.
2.  Run the generation command from the repo root:
    ```bash
    just proto
    ```
3.  Commit the changes to `proto/` and `gen/`.

## 🔍 Validation

The generated JSON Schema (`gen/jsonschema/taskflow.v1.Task.jsonschema.json`) is used to:
1.  Validate task file frontmatter.
2.  Provide strict output constraints for AI Agents generating tasks.
