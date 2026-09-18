## 1. Config & Discovery

- [x] 1.1 Implement `~/.config/graft.toml` loading (XDG-aware) with the schema from the graft-config spec (`sources` list, `stems` path, `[collections]` table), `~` expansion, missing-file error identifying the expected path, and unknown-collection error listing available names; verify unit tests cover parsing, `~` expansion, missing file, and unknown collection
- [x] 1.2 Implement source discovery: scan each sources directory exactly one level deep for git repos, skipping hidden and vendor directories (e.g. `node_modules`); verify unit tests cover a fixture sources layout including a hidden dir and a nested-repo case that must be ignored
- [x] 1.3 Document manual config creation with the concrete example from the graft-config spec in `README.md` and in `graft grow`'s error/hint output; verify the docs show the example and the hint text matches the spec

## 2. Manifest

- [x] 2.1 Implement `.graft/manifest.toml` read/write matching the exact schema in the stem-creation spec (branch, collection, created_at RFC 3339, `[[repos]]` entries with name and source path); verify unit tests cover round-trip write/read and graceful handling of missing/invalid manifests

## 3. Grow Command

- [x] 3.1 Implement `graft grow <branch>` core: reject branch names unusable as directory names with a clear error; reject existing stem directories; create the stem directory; per-repo `git worktree add` based on locally known `origin/HEAD` with an actionable per-repo error when unresolvable; verify e2e scenarios: two-repo stem grows, slashed branch rejected before anything is created, existing stem refused, unresolvable origin/HEAD fails one repo without blocking the other
- [x] 3.2 Implement repo selection: `-c/--collection <name>`, `-s/--sources <repo,...>`, cwd-aware single-repo mode, and the no-selection hint fallback; dedupe overlapping selections by repo name; report unknown names as per-repo failures; verify e2e scenarios for each mode plus the hint fallback
- [x] 3.3 Implement warn-and-continue with create summary: collect per-repo outcomes (success, git error, hook failure, unknown name), report summary on stderr, exit non-zero on any failure while preserving successful repos and creating the manifest even on total failure; verify e2e scenarios: one repo's post-checkout hook fails and the other survives; same-branch collision reports as a per-repo failure; all-repos-fail still leaves stem + manifest
- [x] 3.4 Implement the stem-path output contract: absolute stem path as the final stdout line, progress and summary on stderr; verify e2e scenario asserts the final stdout line equals the stem path

## 4. Hooks

- [x] 4.1 Implement `postcreate` execution from global config with cwd set to the stem directory and env-var context (`GRAFT_WORKSPACE`, `GRAFT_PATH`, `GRAFT_BRANCH`, `GRAFT_COLLECTION`, `GRAFT_REPOS` — repos limited to successfully created worktrees), failure-is-warning policy, and `--no-hooks`; verify e2e scenarios: hook writes AGENTS.md from env vars, failing hook warns and stem remains, `--no-hooks` skips
- [x] 4.2 Document the per-repo `post-checkout` creation-guard idiom (all-zeros old SHA) in `README.md`; verify the doc includes the guard snippet and the verified behavior notes (runs synchronously inside `git worktree add`, cwd = new worktree)

## 5. Output Contract

- [x] 5.1 Verify the full integration path an external launcher will use: grow a stem, capture the stem path from stdout, confirm `GRAFT_*` env vars a postcreate hook sees match the spec table; verify one e2e scenario exercises grow + postcreate together and parses the stem path programmatically
