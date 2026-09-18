## Context

Change 1 (`add-grow-command`) established stems, discovery, the colocated `.graft/` manifest, and grow's output contract (stem path as the final stdout line). This change completes the daily loop. Navigation is external (fzf launcher + tmux), so there is no `go`/shell-init command and no interactive navigation UX; the porcelain records defined here are the primary interface for external tooling, and they debut here — grow deliberately has no record format.

## Goals / Non-Goals

**Goals:**
- One-command answers to "what stems exist," "is anything dirty or unpushed," and "can I safely delete this stem"
- A porcelain contract stable enough that an external launcher can be built against it

**Non-Goals:**
- Cross-repo sync, prune, repair, delete-time hooks, shell integration — all post-MVP
- Interactive pickers inside list/status (external tooling's job)
- Network operations (no fetch in status)

## Decisions

**Listing derives from disk + git, with the manifest as a hint.** Iterate stems-directory children; for each, read `.graft/manifest.toml` for the branch/collection identity, then verify against `git worktree list` in each recorded source repo. A missing/invalid manifest marks the stem malformed and it is reported, not silently skipped. Rationale: the manifest is authoritative only for branch identity of the stem (the directory name equals the branch in MVP, so the manifest is genuinely advisory here); for existence and membership, git and the filesystem win.

**Status semantics: dirty = uncommitted + untracked; ahead/behind relative to `origin/HEAD`.** The base ref is `origin/HEAD` as known locally — the same ref grow bases worktrees on — so status answers "is my branch base current?" consistently. Ahead/behind uses `git rev-list --left-right --count`. Status never fetches; staleness shows up as behind counts and fixing it is an explicit user action.

**Delete refuses on uncommitted work or unpushed commits.** The dirty check is shared with status. "Unpushed" is defined concretely per repo: commits not reachable from `refs/remotes/origin/<branch>` when that remote-tracking ref exists; otherwise (branch never pushed), commits beyond the base ref (`origin/HEAD`) — so a freshly grown stem with zero commits deletes freely, while real unpushed work blocks. Deletion order: remove worktrees via `git worktree remove`, delete branches, remove the `.graft/` directory and stem directory last. No grove-style trash quarantine in MVP — the refusal gate is the safety mechanism; background-unlink complexity is unwarranted at this scale.

**One porcelain format, defined here.** Tab-separated records with four record types (stem, repo, collection, source), field order fixed by the machine-readable-output spec. Consumers parse by record type; `collection` records carry repo names as a comma-separated list since they're the only multi-value field.

## Risks / Trade-offs

- [No trash quarantine means `delete --force` is truly destructive] → The per-repo summary always prints before removal (even with `--force`), so what was discarded is recorded in output; quarantine can be added later if real data-loss incidents occur
- [Ahead/behind needs an up-to-date `origin/HEAD`] → Status does not fetch; it reports against locally known refs. Fetching is an explicit user action, consistent with grow's no-network MVP stance
- [Malformed stems could hide work] → Malformed stems are always surfaced in `list` output, never dropped silently

## Open Questions

None blocking.
