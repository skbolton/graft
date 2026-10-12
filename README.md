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

Graft reads `~/.config/graft/config.toml` (or
`$XDG_CONFIG_HOME/graft/config.toml`). Create it manually:

```toml
sources = ["~/repos"]       # directories of pristine source checkouts
stems = "~/workspaces"      # where workstream directories are created

[collections]
backend = ["project_a", "project_b"]
```

The config directory is also the natural home for your hook scripts:

```
~/.config/graft/
  config.toml
  hooks/
    postcreate.sh
```

Graft reads only `config.toml`; it never scans or manages `hooks/` itself.
Upgrading from an earlier version? Move the old file:
`mv ~/.config/graft.toml ~/.config/graft/config.toml`.

`sources` directories are scanned one level deep for git repos; a repo is
identified by its directory name.

## Lifecycle hooks

Graft runs one global `postcreate` hook after growing a stem. Define it in
`config.toml`:

```toml
postcreate = "echo working on $GRAFT_BRANCH > $GRAFT_PATH/NOTES.md"
```

To keep scripts in the config directory, point the command at them, e.g.
`postcreate = "~/.config/graft/hooks/postcreate.sh"` (the hook runs through
the shell, so `~` expands).

The hook runs with the stem directory as cwd and receives context via
environment variables: `GRAFT_WORKSPACE` (stem name), `GRAFT_PATH` (absolute
stem path), `GRAFT_BRANCH`, `GRAFT_COLLECTION` (empty when ad-hoc), and
`GRAFT_REPOS` (newline-separated names of successfully created worktrees).
Pass `--no-hooks` to skip it for one invocation.

Per-repo setup hooks are not managed by graft. Repos bootstrap their own
worktrees through git's `post-checkout` hook, which fires on `git worktree
add` with cwd set to the new worktree. Graft ships helpers so the hook stays
short:

```sh
#!/bin/sh
command -v graft >/dev/null 2>&1 || exit 0

if graft hook on-create "$1"; then
  graft hook copy .envrc .env
  graft hook copy node_modules
fi
```

- `graft hook on-create "$1"` succeeds only when the hook was fired by
  worktree creation: git passes the all-zeros sha as `$1` there, while an
  ordinary branch switch passes a real sha. Forward `$1` unchanged.
- `graft hook copy <path>...` copies files or directories from the repo's
  main worktree (the source checkout) into the current worktree, preserving
  paths. Missing source entries are skipped, so optional files are safe to
  list. Directories copy recursively, using copy-on-write where the
  filesystem supports it and preserving symlinks.

Both helpers derive everything from git, so they work in any worktree of the
repo, however it was created. Since `graft hook copy` is not gated on
creation, it can also be re-run inside an existing worktree to refresh a
stale copy:

```sh
cd ~/workspaces/my-stem/my-repo && graft hook copy .envrc
```

Note git runs the hook synchronously inside `git worktree add`; a failing
hook makes `worktree add` exit 1 while still producing a fully checked-out
worktree, which graft reports as a warning.

Secrets: copying `.env` into every stem widens its on-disk exposure. Consider
copying `.env.example` and templating per stem, or keep secret files out of
the copy list.

## Current work

See `openspec/` for specifications and active work.
