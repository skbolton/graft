# Design: move-config-into-dir

## Context

`config.Path()` currently returns `~/.config/graft.toml` (or `$XDG_CONFIG_HOME/graft.toml`)
and every consumer — `Load()`, the e2e harness's `WriteConfig`, tests — builds on that
return value. The `lifecycle-hooks` spec defines `postcreate` as a shell command; script
collocation in the config directory needs no graft support beyond a directory to live in.

## Goals / Non-Goals

**Goals:**

- One-line relocation of the config path, XDG-aware.
- Error messages point users at the new location.
- Document the `hooks/` subdirectory as the collocation idiom.

**Non-Goals:**

- No legacy-location fallback or migration command: graft is unreleased, so there are no
  users with an old path to carry forward.
- No relative hook-path resolution: `postcreate` stays a shell command; the shell's own
  `~` expansion already lets users write `~/.config/graft/hooks/postcreate.sh`.
- No `GRAFT_CONFIG_DIR` env var for hooks; add one if a use case appears.
- No schema changes to config.toml itself.

## Decisions

1. **Compute a config directory, then join the file name.** `Path()` resolves the XDG
   base directory as today, then joins `graft/config.toml`. Tests and the e2e harness
   update to write to the same location. Alternative considered: a `Dir()` helper
   alongside `Path()` — rejected as unnecessary until another caller needs the directory.

2. **Hard cut, no fallback.** Alternative considered: read `~/.config/graft.toml` if the
   new path is missing, with a deprecation warning. Rejected — graft has no released
   users, and a fallback would leave two code paths and confusing dual-write semantics
   for zero real benefit. The missing-file error identifying the new path is the
   migration message.

3. **Documented, not implemented, hooks directory.** Graft does not create
   `~/.config/graft/hooks/`, list it, or resolve against it. Per design principle 2
   (hands-off per-repo/user hooks), the directory is purely a user convention that
   README documents. This keeps the change to a path constant.

## Risks / Trade-offs

- [Users with an existing `graft.toml` get a missing-config error after upgrade] →
  Acceptable pre-release; error names the new path so the fix is one `mv`.
- [`hooks/` collocation is convention only, discoverable only via README] → Deliberate;
  formalizing it would require hook-path resolution rules that no use case yet demands.
