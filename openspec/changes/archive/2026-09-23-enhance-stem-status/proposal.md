## Why

Working on a feature that spans several repos means cd-ing around the stem to find out where changes live. `graft status <stem>` already reports per-repo health, but it requires naming the stem — and users reach for path-like arguments (`graft status ./`), which fail — and "dirty" alone does not tell you *what* changed: the user still has to open each repo and run `git status` to see the actual files.

## What Changes

- The status argument becomes an optional **path**, not a stem name: `graft status` (no argument) infers the stem from the current working directory; `graft status <path>` infers it from the given path (stem root, a member worktree, or anywhere inside the stem).
- **BREAKING** (pre-release): stem names are no longer accepted as the status argument. To check a different stem, pass its path.
- The per-repo report keeps a one-line summary (clean/dirty, ahead/behind) and shows each dirty repo's full `git status` output beneath it — whatever the user's git configuration renders, not a graft re-implementation.
- A `--porcelain` flag switches the per-repo detail to `git status --porcelain` output, framed by graft repo header records so scripts can attribute blocks to repos.
- Running `graft status` with a path outside any stem (or with no argument outside any stem) SHALL fail with a clear error listing available stems.

Assumptions (recorded, not blocking):

- Dirty detection still uses `git status --porcelain` internally (stable, locale-independent); the default-mode detail is git's normal long output, displayed verbatim.
- No network operations are added; ahead/behind stays based on locally known refs.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `stem-operations`: the Per-stem status requirement changes to take an optional path argument (defaulting to cwd) instead of a stem name, to show each dirty repo's actual `git status` output, and to add a `--porcelain` flag proxying to `git status --porcelain` with graft repo framing.
- `machine-readable-output`: adds the porcelain status framing — a `repo` header record per repo followed by that repo's `git status --porcelain` lines — as a versioned contract for script consumers.

## Impact

- `cmd/graft/status.go`: argument validation moves from `ExactArgs(1)` (stem name) to an optional path argument defaulting to cwd; adds `--porcelain`.
- `internal/stems`: new stem-from-path inference; `Status`/`WriteStatus` extended with full git status detail and porcelain framing.
- `internal/gitops`: helpers for capturing `git status --porcelain` (detection + porcelain mode) and full `git status` output.
- `internal/e2e`: extended coverage for cwd and path inference, status detail, and porcelain framing.
