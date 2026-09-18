## Purpose

Defines the build, packaging, version-reporting, and testing requirements for the graft codebase itself. Every later feature change depends on these foundations, particularly the e2e harness convention for testing behavior that drives real git subprocesses.

## ADDED Requirements

### Requirement: Flake builds a working graft binary
The flake SHALL expose a package that builds the `graft` binary from the Go module.

#### Scenario: Building via nix
- **WHEN** the user runs `nix build` in the project root
- **THEN** the build succeeds and `result/bin/graft` exists and is executable

#### Scenario: Binary runs
- **WHEN** the built binary is invoked with no arguments
- **THEN** it prints help text and exits 0

### Requirement: Dev shell provides Go tooling
The flake dev shell SHALL provide the Go toolchain needed for development, including `go vet`.

#### Scenario: Entering the dev shell
- **WHEN** the user enters `nix develop`
- **THEN** the Go toolchain, including `go vet`, is available on PATH

### Requirement: Version reporting
The `graft` binary SHALL report its version via `graft --version`.

#### Scenario: Version flag
- **WHEN** the user runs `graft --version`
- **THEN** the binary prints a version string and exits 0

### Requirement: End-to-end test harness convention
The test suite SHALL include end-to-end scenarios that execute the compiled `graft` binary against real git fixture repositories created in an isolated temporary HOME. Fixture creation SHALL be provided by shared test helpers, including deterministic git identity configuration.

#### Scenario: Running an e2e scenario
- **WHEN** an e2e test runs
- **THEN** it builds or locates the compiled graft binary, creates an isolated temp HOME, constructs git fixture repositories using shared helpers, and invokes the real binary against them

#### Scenario: Fixtures are self-contained
- **WHEN** fixture repositories are created by the helpers
- **THEN** they require no network access and no host git configuration

### Requirement: Static analysis gate
The project SHALL pass `go vet` on all packages as a check that can be run with a single command.

#### Scenario: Vet passes
- **WHEN** the developer runs the project's check command
- **THEN** `go vet` completes with no findings across all packages
