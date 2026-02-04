# Research: Configuration & Onboarding Flow

**Status**: Proposal
**Created**: 2026-01-03
**Context**: TaskFlow needs to know *where* things are. A user might run it in `my-app/` where docs are in `planning/`, or `docs/project-management/`. The CLI also needs to know where the API "Brain" lives (default: `localhost:8000`).

## The Concept: `.taskflow/config.yaml`
Just like `.git/config` or `.vscode/settings.json`, every TaskFlow-enabled project needs a configuration file.

### Proposed Location
`./.taskflow/config.yaml` (inside the project root).

### Schema Draft
```yaml
# .taskflow/config.yaml

project:
  name: "Desirelines"
  
paths:
  # The root of the planning documents (usually current repo)
  planning: "." 
  
  # Optional: The root of the implementation code (for split-repo setups)
  # TaskFlow can index docs found here too.
  implementation: "../desirelines"

  # Sub-paths (relative to planning root)
  epics: "epics"
  tasks: "tasks"
  research: "research"
  
brain:
  # How to talk to the intelligence layer
  api_url: "http://localhost:8000"
  
  # For local docker setup (optional, if we manage the stack)
  docker_compose_file: ".taskflow/docker-compose.yml"
```

## The Onboarding Flow (`taskflow init`)

**Scenario**: You have a repo `my-cool-app`. You want to add TaskFlow.

**1. Run Init**
```bash
$ cd my-cool-app
$ taskflow init
```

**2. Interactive Wizard**
> 🤖 **Welcome to TaskFlow!**
> 
> ? **Where should we store your planning docs?**
> [ ] ./planning (Recommended)
> [ ] ./docs
> [ ] (Custom)
>
> ? **Initialize Docker Intelligence Layer?** (Requires Docker)
> [Y/n] Y
>
> ⚙️  Creating .taskflow/config.yaml...
> 📂 Scaffolding directory structure...
> 🐳 Generating docker-compose.yml...

**3. Result**
The tool creates:
- `.taskflow/config.yaml`
- `.taskflow/docker-compose.yml` (The Brain stack)
- `.gitignore` entry for `.taskflow/data` (Postgres volume)
- `planning/tasks/.gitkeep`

## The "Brain" Connection
**Critical Decision**: Does every project have its own Brain (Docker stack), or is there one global Brain?

### Option A: Per-Project Brain (Isolated)
- **Pros**: Clean separation. Data lives in the repo folder (`.taskflow/data`).
- **Cons**: You have to run `docker-compose up` for *every* project you work on. Heavy resource usage if multiple projects are open.

### Option B: Global Brain (Centralized)
- **Pros**: One Docker stack running in background serves all your projects.
- **Cons**: API needs to be multi-tenant ("Which project are these vectors for?").
- **Constraint**: You wanted a "local-first" tool.

**Recommendation**: **Option A (Per-Project)** starts simpler. It feels like `npm install` - everything is local to the repo. If you switch projects, you spin up that project's environment.

## Handling "Where to send vectors"
In Option A, the config is simple:
`api_url: http://localhost:8000`

The CLI reads this config. When you run `taskflow list`, it:
1.  Reads `paths.tasks` from config.
2.  Parses markdown files found there.

When you run `taskflow related`:
1.  Reads `api_url`.
2.  Sends content to `localhost:8000/embed`.

## Blockers / Risks
1.  **Docker Friction**: Users *must* have Docker installed. The `init` flow needs to check for this.
2.  **Port Conflicts**: If running two projects, both can't bind port 8000.
    - *Fix*: `taskflow init` could auto-assign a random port or check availability? Or rely on Docker networking.
