# AGENTS.md — graft

> **Status note:** This document describes graft as designed. Implementation is
> in progress; not every capability described here exists yet. Check
> `openspec list` for what is active, and the change artifacts under
> `openspec/changes/` for what is planned or done. When this document and a
> change artifact disagree, the change artifact wins.

## What graft is

Graft is a Go CLI that manages multi-repo git worktree workspaces. One command
creates a **stem**: a named directory containing one git worktree per selected
source repository, all checked out to the same branch — so a single feature
that touches several repos gets isolated, disposable workspace everywhere it
needs it, without ever working directly on a source checkout.

The name is the metaphor: sources are the **rootstock** (pristine checkouts,
never worked in directly), stems are the new growth grafted onto them.

## Vocabulary

| Term | Meaning |
|------|---------|
| source | A pristine git checkout under a configured `sources` directory; identified by its directory name; never worked in directly |
| stem | A workstream directory under the configured `stems` directory containing one worktree per selected source repo, all on one branch; named after that branch |
| collection | A named group of source repo names in the global config, used for one-command stem creation |
| manifest | The colocated `.graft/manifest.toml` inside each stem recording branch, collection, created_at, and member repos |
| postcreate | The single global lifecycle hook, run after a stem's repos are created, receiving context via `GRAFT_*` environment variables |

## Design principles

These are settled decisions; do not relitigate them without strong cause:

1. **Stateless core, colocated state.** No central state file. Membership and
   paths are derived from the filesystem and `git worktree list`. The only
   state is each stem's `.graft/manifest.toml`, and it is advisory except for
   branch identity.
2. **Hands-off per-repo hooks.** Graft never reads or executes per-repo
   configuration. Repos that want worktree setup use git's own `post-checkout`
   hook (see below). Graft documents the idiom; it does not own it.
3. **Env-var hook context.** Lifecycle hooks receive context exclusively as
   `GRAFT_*` environment variables — no positional placeholders, no shell
   interpolation of user data.
4. **No network in the create path.** Worktrees are based on locally known
   `origin/HEAD`. Graft does not fetch in MVP.
5. **Porcelain is a versioned contract.** Machine-readable output formats
   (tab-separated records) must not change within a major version; external
   launchers are built against them.

## Verified git behavior (treat as facts)

`git worktree add` fires the repo's `post-checkout` hook synchronously, with
cwd set to the new worktree. The old-HEAD argument is all zeros on worktree
creation, which is how per-repo setup hooks distinguish creation from ordinary
branch switches:

```sh
[ "$1" = "0000000000000000000000000000000000000000" ] || exit 0
```

A failing `post-checkout` hook makes `git worktree add` exit 1 while still
producing a fully checked-out worktree — treat hook failures as warnings, not
as missing worktrees. If git behavior ever seems to contradict this, re-verify
with a scratch repo before changing code.

## Work order

Changes are implemented in dependency order — check `openspec list` for
current status

## Verification

Before considering any work done, run the project check and make sure it
passes:

```sh
nix run .#check
```

This runs `go vet ./...` and `go test ./...`, then builds the `graft` package
through the flake (`nix build .#graft`) against your working tree. The build
gate catches packaging drift — dependency changes that need a `vendorHash`
update, sandbox inputs the tests depend on — that vet and tests alone miss.
Do not declare a task complete, summarize results, or hand work back to the
user without a passing check. Run it again after any follow-up edits.

## Implementation conventions

- Packages live under `internal/` (`config`, `discovery`, `grow`, `hooks`,
  `manifest`, `gitops` — created as needed, not preemptively); the binary
  entry point is `cmd/graft/main.go`
- CLI framework is Cobra
- E2E tests drive the compiled binary against real git fixture repos in an
  isolated temp HOME; shared fixture helpers live with the harness (see
  `add-project-scaffolding` design). No git mocking.
- Static analysis and tests run through the `check` flake app (see
  Verification); its contents are the gate — do not loosen them
- Deferred/post-MVP items are listed in each change's proposal — do not
  implement them opportunistically
