## Context

See proposal.md for motivation. Current state: graft documents a raw per-repo `post-checkout` idiom (README) and owns nothing hook-related beyond the global `postcreate` lifecycle hook. Two git behaviors are verified with a scratch repo and are the foundation of this design:

- A `post-checkout` hook fired by `git worktree add` runs with cwd = the new worktree, `$1` = all-zeros sha, and can resolve the main worktree via `git rev-parse --path-format=absolute --git-common-dir` (requires git ≥ 2.31, released 2021).
- Environment variables exported by graft when it invokes `git worktree add` reach the hook — available as a future stem-context lever, not used here.

Constraint from the project: graft never reads or executes per-repo configuration; helpers must be opt-in tools repos call, not mechanisms graft drives.

## Goals / Non-Goals

**Goals:**
- Make the per-repo hook a copy-paste of a few readable lines instead of hand-rolled shell plumbing
- Helpers that work in any worktree of a repo, created by graft or not
- Copying that handles both config files (`.envrc`) and ignored artifact directories (`node_modules`) with copy-on-write efficiency where available

**Non-Goals:**
- Stem awareness inside per-repo hooks (`graft hook stem`): the manifest is written after all `worktree add` calls, so creation-time context requires exporting `GRAFT_*` env into git — a contract to own deliberately later
- Pattern/glob-based copying with includes/excludes config (gtr's 600-line copy.sh surface): literal paths only until usage proves the need
- Graft-driven copying during grow (would require graft to read per-repo config)
- Re-sync command for refreshing copies in existing stems; manual `graft hook copy` re-run covers the escape hatch

## Decisions

### 1. `graft hook` command namespace with `on-create` and `copy`

New cobra command tree `graft hook` with subcommands `on-create` and `copy`, implemented in a new `internal/hookhelpers` package (keeping `internal/hooks` for graft's own lifecycle hook). The guard is a *predicate* the hook composes under `if`, per the user's mental model:

```sh
if graft hook on-create "$1"; then
  graft hook copy .envrc .env
fi
```

Alternatives considered: wrapper form (`graft hook create -- cmd...` — hides the git convention but doubles line length and breaks the ungated-copy re-run); raw-shell guard only (rejected: user wants graft to carry the idiom); gtr-style graft-owned copy config (rejected: contradicts the hands-off per-repo design principle).

### 2. Source discovery: `git-common-dir`, no graft state

`graft hook copy` resolves the source worktree as the parent of `git rev-parse --path-format=absolute --git-common-dir` run in the caller's cwd. Verified: from inside a worktree this resolves to the main checkout's `.git`. This makes helpers graft-state-free per the spec's independence requirement, and means the helper behaves identically in worktrees created by graft, gtr, grove, or plain git. Fallback if `--path-format=absolute` is unsupported (git < 2.31): derive from the first entry of `git worktree list --porcelain`, which is always the main worktree.

### 3. Copy mechanics: stdlib for files, `cp` for directories

Single files copy through stdlib (`os`/`io`). Directories shell out to the platform `cp` — `cp --reflink=auto -RP` on Linux, `cp -cRP` on macOS — giving CoW on Btrfs/XFS/APFS (the `node_modules` case) with automatic fallback, and symlink preservation via `-P`. If the subprocess fails, graft falls back to a recursive stdlib copy so behavior never depends on cp's presence quirks. This mirrors the gtr-proven approach at a fraction of its code: gtr's 617-line copy.sh exists mostly for glob resolution, excludes, and dry-run — all out of scope here.

Destination mapping is `src/<path>` → `cwd/<path>`; the caller's cwd is the new worktree because git sets it. Paths are rejected if absolute or containing `..` segments (same traversal guard gtr applies); the helper copies out of the main worktree only, never into arbitrary destinations.

### 4. Failure and reporting semantics

- `on-create` prints nothing on either exit path (the `if` consumes the status); usage errors print to stderr.
- `copy` prints one line per copied/skipped path to stderr — stderr, not stdout, so the command tree never competes with porcelain output (machine-readable formats are a versioned contract). Missing source entries are silent skips (optional files are the normal case); actual copy failures are reported and cause exit 1, with remaining paths still attempted.
- Running `copy` from the main worktree itself (source == cwd) notes it on stderr and exits 0 — likely a misuse, not an error worth failing a checkout over.

### 5. Documented degradation for graft-less hosts

The README hook example guards with `command -v graft >/dev/null 2>&1 || exit 0` before the `if`, so repos adopting helpers keep working wherever graft is absent. A failing hook is already a warning in graft's grow summary, so worst case degrades gracefully; the guard makes it a no-op instead.

### 6. Secrets documentation, not enforcement

Copying `.env` into N stems widens secret exposure on disk. Graft documents the `.env.example`-vs-`.env` consideration in the hook section rather than adding allow/deny machinery — enforcement would require graft to know about repo file semantics, which is out of scope.

## Risks / Trade-offs

- [Repos soft-depend on graft being on PATH] → Documented `command -v` guard keeps hooks functional without graft
- [`cp` flag differences across platforms] → Two-platform case map (linux/darwin) with stdlib fallback on any failure; Windows is out of scope (graft targets the user's Linux/Nix environment)
- [Helpers invoked outside a git repo fail confusingly] → `copy` checks for repository context and errors with an actionable message ("run inside a worktree of the repo")
- [Symlinked `.envrc` tricks (link into secrets)] → Out of scope; the secrets doc note covers the threat model at MVP depth
- [Manifest-less stem detection deferred means hooks can't branch on graft stems] → Accepted; env-export lever is verified and additive when demand appears

## Migration Plan

Additive commands; no config schema, porcelain, or spec behavior changes for existing users. Rollback is trivial (unused subcommands).

## Open Questions

None. Remaining choices (package naming, exact stderr wording) are implementation-level.
