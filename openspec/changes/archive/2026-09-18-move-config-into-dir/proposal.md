# Proposal: move-config-into-dir

## Why

The global config currently sits alone at `~/.config/graft.toml`. Users writing hook
scripts want to collocate those scripts with the config — one directory to back up, one
directory to point at — but a bare `graft.toml` file leaves nowhere to put them. Moving
the config into `~/.config/graft/` turns it into a graft-owned directory where hook
scripts can live alongside it:

```
~/.config/graft/
  config.toml
  hooks/
    postcreate.sh
```

## What Changes

- **BREAKING**: The global config path moves from `~/.config/graft.toml` to
  `~/.config/graft/config.toml` (XDG-aware: `$XDG_CONFIG_HOME/graft/config.toml`).
- The missing-config error message identifies the new path.
- No fallback to the legacy location: graft is unreleased, so the old path is simply no
  longer read. The error message is the migration guide — users move their file.
- README and spec prose updated to show the new layout, with the hooks/ subdirectory
  presented as the collocation idiom (user-managed; graft never reads it).
- Config schema (sources, stems, collections, postcreate) is unchanged; only the location
  moves.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `graft-config`: the config location requirement changes from `~/.config/graft.toml` to
  `~/.config/graft/config.toml` (XDG-aware), and the missing-config error SHALL identify
  the new path.

## Impact

- `internal/config/config.go` — `Path()` joins the new directory/name; error text updated.
- `internal/config/config_test.go` — path and error-message assertions updated.
- `internal/e2e/harness.go` — `WriteConfig` writes to the new location.
- `README.md` — documented path and example layout.
- `openspec/specs/graft-config/spec.md` — location requirement.
- No change to `internal/hooks` or the `lifecycle-hooks` spec: `postcreate` remains a
  shell command whose scripts may reference `~/.config/graft/hooks/` via ordinary shell
  paths; graft does not resolve hook paths against the config directory.
