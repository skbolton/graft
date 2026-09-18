# graft

Multi-repo git worktree workspaces. One command creates a **stem**: a named
directory containing one git worktree per selected source repository, all
checked out to the same branch — so a single feature that touches several
repos gets isolated, disposable workspace everywhere it needs it, without
ever working directly on a source checkout.

## Install

Build from source with Nix:

```sh
nix build
result/bin/graft --help
```

Or enter the dev shell for development:

```sh
nix develop
```

## Configuration

Graft reads `~/.config/graft.toml` (or `$XDG_CONFIG_HOME/graft.toml`). Create
it manually:

```toml
sources = ["~/repos"]       # directories of pristine source checkouts
stems = "~/workspaces"      # where workstream directories are created

[collections]
backend = ["project_a", "project_b"]
```

`sources` directories are scanned one level deep for git repos; a repo is
identified by its directory name.

## Lifecycle hooks

Graft runs one global `postcreate` hook after growing a stem. Define it in
`graft.toml`:

```toml
postcreate = "echo working on $GRAFT_BRANCH > $GRAFT_PATH/NOTES.md"
```

The hook runs with the stem directory as cwd and receives context via
environment variables: `GRAFT_WORKSPACE` (stem name), `GRAFT_PATH` (absolute
stem path), `GRAFT_BRANCH`, `GRAFT_COLLECTION` (empty when ad-hoc), and
`GRAFT_REPOS` (newline-separated names of successfully created worktrees).
Pass `--no-hooks` to skip it for one invocation.

Per-repo setup hooks are not managed by graft. Use git's own `post-checkout`
hook, which fires on `git worktree add` with cwd set to the new worktree, and
guard it so it only acts on creation:

```sh
[ "$1" = "0000000000000000000000000000000000000000" ] || exit 0
```

The guard works because the old-HEAD argument is all zeros only on worktree
creation. Note git runs the hook synchronously inside `git worktree add`; a
failing hook makes `worktree add` exit 1 while still producing a fully
checked-out worktree, which graft reports as a warning.

## Current work

See `openspec/` for specifications and active work.
