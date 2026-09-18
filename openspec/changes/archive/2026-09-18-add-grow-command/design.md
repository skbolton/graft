## Context

Graft is a multi-repo git worktree orchestrator built on two proven models: grove's workspace orchestration (global config, workspaces of worktrees across repos, presets, lifecycle hooks) and gtr's git-config-based configuration and porcelain output conventions. The proposal (`add-grow-command/proposal.md`) captures the WHY; this document records the decisions made during exploration and their rationale.

Constraints from the user's environment and workflow:
- Navigation is external: an fzf launcher creates tmux sessions from templates. Graft never shells out to tmux and provides no `cd` integration.
- The user never uses `/` in branch names; the MVP rejects such names outright rather than mapping them.
- Per-repo setup must remain in the repos themselves; graft must not accumulate per-repo config files.
- MVP must be fast to ship; network operations, format design, and UI work are deferred.

## Goals / Non-Goals

**Goals:**
- Stem creation must be a one-command experience for both multi-repo collections and single-repo "simple project" use
- No central state that can drift from reality
- Hooks that are trivial to write without knowing graft's argument conventions
- Zero network operations in the MVP create path

**Non-Goals:**
- Cross-repo sync/push, repair/idempotent re-grow, provenance, hook helper utilities, delete-time hooks, shell integration — all post-MVP
- Managing or executing per-repo setup hooks
- Fetch-on-create, per-repo base branch overrides, slugified branch names, interactive pickers, porcelain record formats (all explicitly deferred)

## Decisions

### 0. TOML parsing via go-toml/v2

The config file is user-authored TOML and the stem manifest is read back as
TOML, but Go's stdlib has no TOML support. Hand-rolling a parser would put
user-config edge cases on graft's bug surface, so the manifest emit is
hand-rolled (it must match the documented schema byte-for-byte) while all
TOML parsing uses `github.com/pelletier/go-toml/v2`. This extends the
"Cobra + stdlib" dependency goal from the scaffolding design; one small,
widely used dependency is the cheaper side of that tradeoff.

### 1. Stateless core with colocated manifest

No global state file. Stem membership, branches, and paths are derived from the directory layout and `git worktree list` in each source repo. The only state is a colocated `.graft/manifest.toml` per stem, storing what git cannot tell us: branch, collection name, created_at, and the selected repos with source paths (exact schema in the stem-creation spec).

Rationale: grove's central `state.json` requires atomic writes, a `.trash` quarantine, and a `doctor` command *because* it drifts from disk reality. Deriving state eliminates that failure class. The manifest is a hint for listing; git and the filesystem are the source of truth.

Alternative considered: grove-style central state — rejected; its entire troubleshooting surface exists to reconcile state vs. reality, and reality is directly readable here.

### 2. Stem identity is the branch name; invalid names are rejected

`graft grow <branch>` uses the branch as the stem directory name. Names unusable as directories (containing `/`, empty, etc.) are rejected with a clear error before anything is created. Slugification is deliberately deferred: the user's branches never contain `/`, and the mapping machinery (slug↔branch in every downstream command) costs more than it buys today. Re-adding it later is additive.

### 3. Base ref is locally known origin/HEAD; no network

Worktrees are based on each repo's `origin/HEAD` as already known locally. Graft never fetches in MVP — the sources are rootstock the user keeps current, and removing network ops eliminates auth handling, timeout handling, hang prevention (`GIT_TERMINAL_PROMPT=0`), and network e2e complexity from the create path. An unresolvable `origin/HEAD` is a per-repo failure with an actionable error (`git remote set-head origin -a`).

Alternatives considered and deferred: fetch-on-create with `--no-fetch` escape (network in the hot path), per-repo `graft.baseBranch` git-config override (no repos need it yet).

### 4. Hands-off per-repo hooks

Graft never reads, executes, or validates per-repo configuration. Repos that want worktree setup use git's own `post-checkout` hook, which fires on `git worktree add` with cwd set to the new worktree (verified experimentally). The creation guard is documented by graft, not provided by it:

```sh
[ "$1" = "0000000000000000000000000000000000000000" ] || exit 0
```

Git runs post-checkout synchronously within `worktree add`, so graft simply drives the checkout and waits. Hook utilities (a gate command, copy helper, stem lookup) are deliberately deferred until real hook-writing creates demand.

Alternative considered: graft-owned per-repo config keys executed by graft (gtr-style `gtr.hook.postCreate`) — rejected for MVP because git-native hooks work today, work outside graft, and keep graft's trust surface empty. May be revisited later.

### 5. Postcreate hook with env-var context

The MVP has one lifecycle hook: `postcreate`, defined in global config, executed after all repos are attempted, with cwd set to the stem directory. Context arrives as environment variables:

| Variable | Value |
|---|---|
| `GRAFT_WORKSPACE` | Stem directory name |
| `GRAFT_PATH` | Absolute path to stem root |
| `GRAFT_BRANCH` | Branch name |
| `GRAFT_COLLECTION` | Collection name, empty if ad-hoc |
| `GRAFT_REPOS` | Newline-separated names of successfully created worktrees |

No positional placeholders, no shell interpolation of user data — env vars are not subject to the injection concerns grove documents for its `{placeholder}` expansion. A `postcreate` failure is a warning; the stem stays intact. `precreate` and its abort policy are deferred until a use case exists.

### 6. Warn-and-continue failure policy

A repo failing to create (git error, hook failure, unresolvable `origin/HEAD`, unknown name) never aborts the stem. Each repo's outcome is collected and reported in a create summary; the stem exists with whatever repos succeeded, and the manifest records them. Note git's post-checkout failure makes `worktree add` exit 1 while still producing a checked-out worktree — the summary reports it as a warning, not a missing worktree. The stem directory and manifest are created even when every repo fails, so partial outcomes are inspectable. Same-branch collisions (git refuses to check out one branch in two worktrees of a repo) surface through this same path with git's error text.

### 7. Stem path as the grow output contract

On completion, grow prints the absolute stem path as the final line of stdout; all progress and the summary go to stderr. This is the integration point for the user's fzf/tmux launcher, which templates sessions from the stem path. Full porcelain record formats are designed once, in the daily-driver-ops change, where there is real listing data to format.

### 8. No-selection fallback: hint and exit

When no collection, no `--sources`, and no cwd context is given, graft prints a hint naming the selection options and exits non-zero. No built-in picker in MVP. The anticipated future extension is a config option naming an external picker command (fzf, gum, ...) that graft invokes with the candidate list and reads the selection from — keeping graft dependency-free while delegating UI taste to the user. That extension is post-MVP and intentionally unspecced until needed.

### 9. Discovery: one level deep, name-keyed

Sources directories are scanned exactly one level deep, matching the documented layout (`~/c/sources/<project>`), skipping hidden dirs and common vendor directories. A repo is identified by its directory name; duplicate names within one selection collapse to a single repo. Remote-URL deduplication across differently named checkouts is deferred — it requires per-repo `git config` inspection and only matters for layouts graft's sources convention doesn't produce.

## Risks / Trade-offs

- [No fetch means stems can be cut from stale wood if the user forgets to update sources] → Accepted for MVP; the user's rootstock discipline covers it. Fetch-on-create is the first post-MVP candidate if this bites
- [origin/HEAD is unset in some clones] → Per-repo failure with an actionable error rather than a silent fallback; git sets origin/HEAD on normal clones, so this is rare in practice
- [Manifest can be hand-edited or deleted] → The manifest is a hint for listing; git worktree paths still resolve. `graft list` treats a missing/invalid manifest as a malformed stem and reports it
- [Stem with zero successful repos still exists on disk] → Accepted: inspectable via manifest and summary; cleanup via later `graft delete`

## Open Questions

None blocking. Deferred items (fetch-on-create, base branch overrides, precreate hook, slugification, picker config, porcelain records, hook helpers, repair, provenance) are recorded in the proposal's out-of-scope list and do not affect the specs or task breakdown.
