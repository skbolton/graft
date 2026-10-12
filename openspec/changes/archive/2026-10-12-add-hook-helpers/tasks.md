## 1. Command scaffold

- [x] 1.1 Add the `graft hook` command tree (cobra) with `on-create` and `copy` subcommands wired to a new `internal/hookhelpers` package; verify `graft hook --help` lists both and `go vet ./...` passes

## 2. on-create helper

- [x] 2.1 Implement `graft hook on-create <old-sha>`: exit 0 on the all-zeros sha, exit 1 on any other sha, usage error (exit non-zero, stderr message) when the argument is missing; verify unit tests cover all three exits
- [x] 2.2 E2E: fixture repo with a `post-checkout` hook calling `graft hook on-create "$1"` and `git worktree add`; assert exit 0 during creation and exit non-zero on an ordinary `git checkout` branch switch in the main worktree

## 3. copy helper

- [x] 3.1 Implement source discovery via `git rev-parse --path-format=absolute --git-common-dir` with the `git worktree list --porcelain` first-entry fallback; error with an actionable message when run outside a git repo; verify unit tests against a scratch repo
- [x] 3.2 Implement file copy: `src/<path>` → `cwd/<path>` with relative-path preservation, symlink recreation for symlinked files, quiet skip for missing source entries, and rejection of absolute or `..`-containing paths; verify unit tests cover each
- [x] 3.3 Implement directory copy via platform `cp` (`--reflink=auto -RP` on Linux, `-cRP` on macOS) with stdlib recursive-copy fallback and symlink preservation; verify a fixture directory containing a symlink copies as a symlink, and that a failed `cp` falls back to the stdlib path
- [x] 3.4 Implement failure semantics: attempt all requested paths, report failures on stderr, exit non-zero if any failed; skip-with-stderr-note (exit 0) when source == cwd; verify unit tests
- [x] 3.5 E2E: fixture repo with untracked `.envrc` and ignored `node_modules/` dir; grow a stem via the compiled binary with a hook calling `graft hook copy`; assert both appear in the new worktree with correct content and the symlink case preserved

## 4. Documentation

- [x] 4.1 Rewrite the README per-repo hook section around the helper idiom: `command -v graft` guard, `if graft hook on-create "$1"` composition, copy examples for files and directories, the manual re-sync escape hatch (`graft hook copy` inside an existing worktree), and the secrets note (`.env.example` vs `.env`); verify the doc renders the full runnable snippet

## 5. Verification

- [x] 5.1 Run `nix run .#check` (vet + tests + flake build) and ensure it passes with no vendorHash drift
