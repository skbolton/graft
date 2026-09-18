## Why

Graft is a greenfield Go CLI with no buildable code yet. Its core behavior is driving real `git` subprocesses against real repositories, so the test harness convention (isolated temp HOME, real git fixtures, compiled-binary e2e tests) must exist before feature work starts — retrofitting it later would mean rewriting tests instead of extending them.

## What Changes

- Initialize a Go module with a standard `cmd/graft` + `internal/` layout
- Add a minimal Cobra CLI skeleton: `graft` with help and a `--version` flag
- Package the binary via the existing flake (`buildGoModule`) and provide a Go dev shell
- Wire up `go vet` and lint checks
- Establish the e2e test harness convention: each scenario runs the compiled `graft` binary against real git fixture repositories created in an isolated temp HOME
- No feature behavior in this change — the deliverable is a working, tested, packaged binary

Out of scope: CI/release automation, any graft feature behavior.

## Capabilities

### New Capabilities
- `project-toolchain`: build, package, version-reporting, and test-harness requirements for the graft codebase itself

### Modified Capabilities
(none)

## Impact

- `flake.nix` gains a `buildGoModule` package output and dev shell tooling
- New files: `go.mod`, `go.sum`, `cmd/graft/`, `internal/`, test helper packages
- Introduces the Cobra dependency; otherwise stdlib-only
- All later changes build on the harness convention established here
