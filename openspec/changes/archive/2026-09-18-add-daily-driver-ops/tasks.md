## 1. List

- [x] 1.1 Implement `graft list` human output: iterate stems-directory children, read manifests, verify membership against `git worktree list`; verify e2e scenario with two fixture stems
- [x] 1.2 Implement malformed-stem reporting (missing/invalid manifest) without crashing; verify e2e scenario with a stem whose manifest is deleted
- [x] 1.3 Implement `graft list --porcelain` emitting the four record types (stem, repo, collection, source) in the exact format defined in the machine-readable-output spec; verify e2e scenario parses all record types programmatically

## 2. Status

- [x] 2.1 Implement `graft status <stem>`: per-repo dirty detection (uncommitted + untracked) and ahead/behind counts via `git rev-list --left-right --count` against `origin/HEAD` as known locally; verify e2e scenarios: dirty repo, ahead repo, clean stem
- [x] 2.2 Verify status performs no network operations (no fetch) and reports against locally known refs; verify e2e scenario runs with no network access

## 3. Delete

- [x] 3.1 Implement `graft delete <stem>`: pre-deletion per-repo summary, blocking check (dirty, or commits not on `refs/remotes/origin/<branch>` when it exists, or commits beyond `origin/HEAD` when it does not), `--force` override, then `git worktree remove` → branch deletion → `.graft/` and stem directory removal; verify e2e scenarios: clean delete, dirty refusal removes nothing, forced delete proceeds
- [x] 3.2 Verify the full loop end to end: grow → status (dirty) → delete refused → clean → delete succeeds, as a single e2e scenario
