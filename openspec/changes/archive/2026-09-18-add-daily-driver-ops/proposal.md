## Why

Once stems can be created, the daily loop needs support: seeing what stems exist, checking their health across every repo, and cleaning them up safely. The cleanup step matters most — the recurring failure mode this tool exists to prevent is rushed deletion of peripheral-repo work. Additionally, the primary navigation workflow is external (an fzf launcher that starts tmux sessions from a template), so graft's listings must be machine-readable as a first-class contract, not a screen-scraping afterthought.

## What Changes

- `graft list` — show all stems with their branches and member repos (human-readable); `--porcelain` emits the machine-readable contract
- `graft list --porcelain` is the single machine-readable surface: it emits four record types — stems, per-stem repos, configured collections, and discovered source repos — giving fuzzy-finder utilities a stable data source without duplicating discovery logic
- `graft status <stem>` — per-repo health: dirty worktrees (uncommitted or untracked changes) and ahead/behind counts against the repo's base ref (`origin/HEAD`)
- `graft delete <stem>` — remove the stem's worktrees, branches, and `.graft/` directory; **refuse when any repo has uncommitted work or commits not pushed to origin unless `--force` is given**, printing a per-repo summary first
- Listing derives from the filesystem and `git worktree list`; the `.graft/` manifest is a hint, with git as the source of truth

Out of scope (post-MVP): cross-repo sync, `graft go`/shell integration (external tmux launcher covers navigation), prune, repair, delete-time hooks, grow-side porcelain records (grow's output contract is the stem path on stdout).

## Capabilities

### New Capabilities
- `stem-operations`: listing, status, and deletion requirements for stems
- `machine-readable-output`: the porcelain record contract consumed by external launchers and scripts

### Modified Capabilities
(none)

## Impact

- Builds on `add-grow-command`'s discovery, manifest, and config capabilities
- `graft delete` performs destructive operations; the dirty-work refusal is the safety contract
- External consumers: fzf/tmux launcher tooling is the primary porcelain consumer
