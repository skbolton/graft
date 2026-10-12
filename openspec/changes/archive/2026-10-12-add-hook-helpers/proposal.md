## Why

Graft's design keeps per-repo setup in the repos' own git `post-checkout` hooks, documented as a raw idiom (all-zeros creation guard, manual source discovery). Writing that idiom correctly per repo — discovering the source worktree, guarding, copying optional files — is real work repeated in every repo that wants `.envrc`-style bootstrapping. Gribbed-by-gtr-style tools get one config line; graft users get boilerplate. Hook helper subcommands keep graft hands-off (repos opt in by calling graft) while carrying the mechanics.

## What Changes

- New `graft hook` command namespace with two helpers a repo's `post-checkout` hook can call:
  - `graft hook on-create <old-sha>`: exits 0 iff the sha is the all-zeros creation marker (worktree creation, not an ordinary branch switch); exits non-zero otherwise; usage error on a missing argument.
  - `graft hook copy <path>...`: copies files or directories from the repo's main worktree (the rootstock source, discovered via git plumbing — no graft state) into the current directory, preserving relative paths. Directories copy recursively with copy-on-write/reflink where the filesystem supports it, falling back to a plain recursive copy. Missing source entries are skipped quietly.
- README documentation updated: the per-repo hook section shows the helper-based idiom (guard + copy composed with `if`), notes the degradation pattern for hosts without graft installed, and adds a secrets note (`.env.example` vs `.env` exposure when copying into multiple stems).
- Stem-awareness helper (`graft hook stem`) is deliberately not included: the manifest is written after worktree creation, so creation-time stem context would require graft exporting env into `git worktree add`; deferred until real hook-writing demands it.

## Capabilities

### New Capabilities

- `hook-helpers`: graft-provided helper commands executed by repos' own `post-checkout` hooks: creation detection (`on-create`) and source-worktree copying (`copy`), including source discovery, copy semantics for files and directories, and failure behavior.

### Modified Capabilities

None. README documentation updates are not spec-level behavior changes.

## Impact

- New package or extension of `internal/` (likely `internal/hookcmds` or a `hook` subcommand tree under `cmd/graft`); no changes to grow, config, manifest, or lifecycle hooks.
- No porcelain output is affected: helpers print to stderr for humans (and nothing machine-parsed in MVP).
- No new dependencies; copy uses stdlib `os`/`io/fs` plus optional reflink via `cp` invocation or direct syscall fallback decision made in design.
- Repos that adopt helpers gain a soft dependency on `graft` being on PATH; documented degradation keeps hooks functional without it.
