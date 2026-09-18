## Purpose

Defines the daily-driver operations over existing stems: listing stems and their repos, reporting per-repo health, and safe deletion with dirty-work protection.

## Requirements

### Requirement: List stems
`graft list` SHALL show all stems under the configured stems directory, each with its branch and member repos. Listing SHALL derive from the filesystem and `git worktree list` in the source repos; the `.graft/` manifest is a hint, not the source of truth.

#### Scenario: Listing stems
- **WHEN** two stems exist under the stems directory
- **THEN** `graft list` shows both with their branches and member repos

#### Scenario: Malformed stem
- **WHEN** a stem directory is missing its `.graft/manifest.toml` or the manifest is invalid
- **THEN** `graft list` reports it as malformed rather than crashing

### Requirement: Per-stem status
`graft status <stem>` SHALL report per-repo health: dirty worktrees (uncommitted or untracked changes) and commits ahead of/behind the repo's base ref, where the base ref is `origin/HEAD` as known locally — consistent with what `graft grow` would base a new worktree on.

#### Scenario: Dirty repo surfaced
- **WHEN** a stem contains a repo with uncommitted or untracked changes
- **THEN** `graft status` marks that repo dirty and lists the others cleanly

#### Scenario: Branch drift surfaced
- **WHEN** a stem contains a repo whose branch is ahead of or behind its base ref (`origin/HEAD`)
- **THEN** `graft status` reports the ahead and behind counts for that repo

### Requirement: Safe deletion
`graft delete <stem>` SHALL remove the stem's worktrees, workspace branches, and `.graft/` directory. A repo blocks deletion when it has uncommitted or untracked changes, OR when its branch has commits not available on origin — specifically, commits not reachable from `refs/remotes/origin/<branch>` when that ref exists, or commits beyond the base ref when it does not. Deletion SHALL refuse while any repo blocks, unless `--force` is given. A per-repo summary SHALL be printed before deletion regardless.

#### Scenario: Deleting a clean stem
- **WHEN** `graft delete` runs on a stem where every repo is clean and all branch work is available on origin
- **THEN** all worktrees, the workspace branches, the stem directory, and the `.graft/` manifest are removed

#### Scenario: Dirty stem refused
- **WHEN** `graft delete` runs on a stem with uncommitted or unpushed work
- **THEN** graft exits non-zero, prints the per-repo summary showing where work exists, and removes nothing

#### Scenario: Forced deletion
- **WHEN** `graft delete --force` runs on a blocked stem
- **THEN** deletion proceeds and the summary still reports what was discarded
