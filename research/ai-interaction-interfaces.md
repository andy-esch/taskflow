# Research: AI Interaction Interfaces (MCP vs API)

**Status**: Proposal
**Created**: 2026-01-03
**Context**: The user wants TaskFlow to enable AI agents to "update/revise/create docstrings, inline documentation, and actual documentation." This raises the question: **How does an AI (like Claude, Gemini, or GitHub Copilot) talk to TaskFlow?**

## The Interface Options

### Option 1: TaskFlow as a Library / Tool (Current State)
**Mechanism**: The AI (like me right now) uses `run_shell_command` to execute `./bin/taskflow list` or `read_file` to inspect the index.
- **Pros**: Works with any "Agent" that has shell access.
- **Cons**: "Dumb" text parsing. The AI has to figure out the CLI syntax. High token cost to read big outputs.

### Option 2: Model Context Protocol (MCP) Server ⭐ **Strong Contender**
**Mechanism**: TaskFlow runs an MCP Server (standardized by Anthropic).
- **How it works**:
    - TaskFlow exposes "Resources" (Tasks, Epics) and "Tools" (Create Task, Search Docs) via the MCP protocol.
    - An MCP-compliant client (Claude Desktop, Zed, or a generic Agent) connects to it.
- **Value**: The AI "knows" how to interact with TaskFlow natively. It doesn't need to be taught CLI commands. It just sees `get_tasks()` or `search_vectors()`.
- **Fit**: Perfect for your goal of "offloading from human/AI". The protocol handles the schema.

### Option 3: TaskFlow as a "Context Broker"
**Mechanism**: TaskFlow generates a context bundle.
- **Workflow**: `taskflow context --format=xml` -> Dumps a compressed, prioritized summary of the project state.
- **Usage**: You paste this into ChatGPT/Claude web UI. "Here is my project context, write the docstring for function X."

## The "Updating Documentation" Workflow

You want the AI to update docstrings/docs in the *implementation* repo (`desirelines`).

**The Flow:**
1.  **Context**: The AI needs to know *what* to update.
    - Query TaskFlow: `Use Semantic Engine to find related tasks/docs for 'Auth'`.
    - Result: "Task: Update Auth Docs", "Doc: docs/auth.md", "Code: src/auth.ts".
2.  **Action**: The AI reads the code, reads the doc, and rewrites the content.
3.  **Commit**: The AI (via shell tool or git integration) commits the change.

**TaskFlow's Role**:
TaskFlow is the **Index**. It tells the AI *where* the relevant docs are. It doesn't necessarily "write" the docstring itself (the AI Agent does that), but TaskFlow provides the "Map" so the AI doesn't have to scan 10,000 files.

## Recommendation: Build an MCP Server Endpoint
We should add an **MCP Server** capability to the Python Service (`services/semantic-engine`).

**Why?**
- It standardizes the interface.
- It prepares TaskFlow for the future of AI IDEs (Zed, VS Code) which will likely consume MCP.
- It allows an AI to perform complex queries ("Find all tasks related to this file") structurally, without regex-parsing CLI output.

**Architecture Update**:
The Dockerized Python service exposes:
1.  **HTTP API** (for Go CLI).
2.  **MCP Server** (via stdio or SSE) for AI Agents.
