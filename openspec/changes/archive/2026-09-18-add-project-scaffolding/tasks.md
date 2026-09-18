## 1. Go Module & CLI Skeleton

- [x] 1.1 Initialize Go module (`go.mod`) with module path for graft; verify `go build ./...` succeeds
- [x] 1.2 Add Cobra dependency and create `cmd/graft/main.go` with root command, help output, and `--version` flag (build-time ldflags variable, default `dev`); verify `graft --version` prints a version and `graft` with no args prints help and exits 0
- [x] 1.3 Create `internal/` package directory convention (empty for now, documented in a short README or doc.go); verify `go vet ./...` passes on all packages
- [x] 1.4 Create a minimal `README.md`: project name, one-line description ("multi-repo git worktree workspaces"), install note pointing at the flake, and a pointer to `openspec/` for current work; verify the file exists and renders sensibly

## 2. Flake Packaging & Dev Shell

- [x] 2.1 Add `buildGoModule` package output to `flake.nix` wiring `graft` and inject the version via ldflags; verify `nix build` produces `result/bin/graft` and that the binary runs
- [x] 2.2 Add dev shell with the Go toolchain and `go vet`; verify `nix develop -c go version` and `nix develop -c go vet ./...` both succeed

## 3. Test Harness

- [x] 3.1 Create shared e2e test helpers that build/locate the compiled graft binary, create an isolated temp HOME, and construct git fixture repos with deterministic `user.name`/`user.email`; verify helpers need no network and no host git config
- [x] 3.2 Write a first e2e scenario running the real binary (version or help) through the harness; verify `go test ./...` passes
- [x] 3.3 Add a single check command (Makefile/Justfile/script) running `go vet` + `go test ./...`; verify the command passes cleanly on a fresh checkout
