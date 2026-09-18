## Context

Greenfield Go project. An existing flake-parts `flake.nix` is scaffolded with empty outputs. No Go module, no CLI, no tests exist. The tool's core behavior is spawning `git` subprocesses against real repositories, which shapes the testing approach: mocking git would test nothing useful, so e2e tests must drive real git against fixture repos.

Reference precedents exist in grove (Go + Cobra, e2e tests with isolated temp HOME and real git fixtures) and gtr (bash, `--porcelain` output convention).

## Goals / Non-Goals

**Goals:**
- A one-command build (`nix build`) producing a working binary
- A test harness convention that later feature changes can extend without redesign
- Minimal dependency surface (Cobra + stdlib)

**Non-Goals:**
- CI/CD, release automation, goreleaser — deferred
- Any graft feature behavior
- Cross-platform story beyond what Go gives for free (Linux + macOS via flake systems)

## Decisions

**Cobra for the CLI.** Grove uses it, the command family (grow, list, status, delete) fits subcommand structure, and it provides help/completions for free. Alternative considered: stdlib `flag` with manual dispatch — less code, but re-implementing help/flag plumbing for a multi-command CLI isn't worth it.

**Module layout: `cmd/graft/main.go` plus `internal/` packages.** Expected package split (created as needed by later changes, not preemptively): `config`, `discovery`, `grow`, `hooks`, `manifest`, `gitops`. Scaffolding establishes only `cmd/` and the convention.

**Packaging via `buildGoModule` in the existing flake.** No goreleaser until there's something to release. The dev shell gets the Go toolchain and `go vet`; `go vet ./...` is the required static analysis gate.

**E2E harness: Go test package that drives the compiled binary.** Shared helpers (`t.Helper()`) create fixture git repos in an isolated temp HOME with deterministic `user.name`/`user.email` config; tests invoke the real `graft` binary against them. This mirrors grove's e2e approach and directly matches graft's git-subprocess reality. Unit tests stay per-package for pure logic (config parsing, slugification).

**Version via build-time variable.** A package-level `var version = "dev"` overridden by ldflags at build time in the flake. Simple, no version-pinning infrastructure.

## Risks / Trade-offs

- [Git fixtures need identity config or commits fail in fresh environments] → Fixture helpers always set `user.name`/`user.email` and disable any host config leakage via the isolated HOME
- [E2E tests cannot run inside the nix build sandbox reliably (git + network-free fixtures are fine, but build-time test invocation of the binary is awkward)] → e2e tests are a devshell/CI concern; `nix build` only builds, it does not test
- [Choosing Cobra early locks the CLI UX] → Low risk: Cobra is the dominant Go CLI framework and the command set is already sketched

## Open Questions

- Exact lint tool (vet-only vs golangci-lint) — vet is the required gate; adding golangci-lint to the dev shell later is additive
