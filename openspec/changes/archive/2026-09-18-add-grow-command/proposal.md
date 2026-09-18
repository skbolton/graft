## Why

Working across related repositories today means either juggling manual worktrees per repo or — more often — making feature-adjacent edits directly on a peripheral repo's main branch and cleaning up the mess later. Tools exist for this (grove, gtr), but each makes tradeoffs graft rejects: grove keeps setup commands in committed per-repo config files and records no preset identity in its hook context; neither offers config stored in git config. Graft glues trees together: one command creates a stem — a named directory of worktrees across a chosen collection of source repos, all on the same branch.

## What Changes

- Global config at `~/.config/graft.toml` (XDG-aware) with a concrete schema:

  ```toml
  sources = ["~/c/sources"]   # directories of pristine source checkouts
  stems = "~/c/stems"         # where workstream directories are created

  [collections]
  backend = ["project_a", "project_b"]
  ```

- Source discovery: scan configured sources directories one level deep (skip hidden and vendor dirs); a repo is identified by its directory name
- `graft grow <branch>` creates a stem at `<stems>/<branch>` with one worktree per selected source repo, each checked out to `<branch>`:
  - `-s/--sources <repo,...>` selects repos explicitly
  - `-c/--collection <name>` selects a configured collection
  - cwd-aware mode: run from inside a source repo with no flags → single-repo stem
  - with no selection and no cwd context → exit non-zero with a hint explaining the selection options
- Branch names that cannot be directory names (e.g. containing `/`) are rejected with a clear error
- Each worktree is based on the repo's locally known `origin/HEAD`. Graft performs no network operations in MVP; sources are expected to be kept current by the user
- Per-repo failure policy is warn-and-continue: any single repo failing (git error, hook failure, missing `origin/HEAD`) never aborts the others; all outcomes are reported in a create summary
- Per-repo setup hooks are NOT managed by graft — repos use their own git `post-checkout` hooks (which fire on worktree creation, synchronously); graft is hands-off apart from documenting the creation-guard idiom
- A `postcreate` hook in global config receives context as environment variables (`GRAFT_WORKSPACE`, `GRAFT_PATH`, `GRAFT_BRANCH`, `GRAFT_COLLECTION`, `GRAFT_REPOS`) — no positional argument knowledge required; this is the integration point for generating per-workstream `AGENTS.md` files
- Each stem gets a colocated `.graft/manifest.toml` recording branch, collection, created_at, and the selected repos with their source paths. No central state file — everything else is derived from `git worktree list` and the directory layout
- On success, grow prints the absolute stem path as the final line of stdout (human progress and the summary go to stderr) — the integration point for external fzf/tmux launchers

Out of scope (post-MVP): cross-repo sync/push, idempotent re-grow/repair, provenance metadata, hook helper utilities, delete-time hooks, shell `cd` integration, fetch-on-create, per-repo base branch overrides, slugified branch names, built-in or configurable interactive pickers.

## Capabilities

### New Capabilities
- `graft-config`: the global config file — schema, defaults, validation
- `source-discovery`: scanning sources directories for graftable repos
- `stem-creation`: the `graft grow` command — stem layout, worktree creation, base ref selection, failure policy, stem-path output
- `lifecycle-hooks`: postcreate hook execution and the env-var context contract

### Modified Capabilities
(none)

## Impact

- Depends on the `project-toolchain` scaffolding change (Cobra, module layout, e2e harness)
- New packages under `internal/`: config, discovery, grow, hooks, manifest
- Requires `git` on PATH; no other runtime dependencies; no network access
- Consumers: external fzf/tmux launchers consume grow's stem-path output and hook env vars
