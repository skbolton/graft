## 1. Change plumbing

- [x] 1.1 In `cmd/graft/status.go`, change arg validation from `ExactArgs(1)` to at most one argument treated as a path, defaulting to cwd (stem names are not accepted). Verify `go build ./...` succeeds
- [x] 1.2 Add `stems.FindByPath(stemsDir, path)` in `internal/stems`: resolve the path to absolute, resolve symlinks (`filepath.EvalSymlinks`, so a symlinked path into a stem still resolves), walk it and its parents, and return the first directory that is a direct child of the stems dir; error otherwise, listing available stems. Verify with a unit test covering: stem root, member worktree, deep inside a stem, outside any stem, nonexistent path, the stems directory itself, and a symlinked path into a stem

## 2. Status detail and porcelain framing

- [x] 2.1 Replace the `IsDirty` call in `stems.Status` with capturing `git status --porcelain` output (new `gitops` helper or extended `IsDirty`); set `Dirty` from non-empty output and store the porcelain lines on `RepoStatus`. Verify existing `internal/stems` tests still pass
- [x] 2.2 Extend `WriteStatus` default mode to run plain `git status` for each dirty repo and print its output verbatim beneath the repo's summary line; clean repos stay one-line. Verify with a unit test covering dirty, clean, and problem repos
- [x] 2.3 Add the `--porcelain` flag in `cmd/graft/status.go` and porcelain rendering: one `repo<TAB><stem><TAB><name><TAB><path>` header record per member repo (same shape as `list --porcelain`) followed by that repo's captured `git status --porcelain` lines verbatim; clean repos emit only the header. Verify with a unit test asserting the record shape and block delimiting

## 3. E2E and gate

- [x] 3.1 Extend the e2e status suite: run the binary bare and with `./` from inside a stem root and from inside a member worktree and assert output (summary lines plus full git status for the dirty repo); run bare and with a path from outside any stem and assert non-zero exit with available stem names in the error; run `--porcelain` and assert header records precede each repo's git status lines. Verify with `go test ./internal/e2e/ -run TestStatus`
- [x] 3.2 Run `nix run .#check` and confirm vet, tests, and the flake build all pass
