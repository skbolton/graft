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

## Current work

See `openspec/` for specifications and active work.
