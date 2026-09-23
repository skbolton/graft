## Context

`graft status <stem>` already exists (`cmd/graft/status.go`, `internal/stems/status.go`): it reports per-repo dirty/clean plus ahead/behind against `origin/HEAD`. Two gaps remain, per the proposal: the stem must be named explicitly, and dirty repos are only flagged, not described. Motivation and requirements: see proposal.md and the delta spec.

`gitops.IsDirty` already shells out to `git status --porcelain` and checks for non-empty output — so the file list comes from a call graft already makes.

One user-reported behavior informed this design: `graft status ./` from a stem root fails today because the sole argument is parsed as a stem name. Per the proposal, the argument becomes a path.

## Goals / Non-Goals

**Goals:**

- `graft status` with no argument resolves the stem from the cwd (stem root or any member worktree).
- `graft status <path>` resolves the stem containing the given path; stem names are not accepted as arguments.
- Per-repo summary lines (clean/dirty, ahead/behind) stay, and each dirty repo's full `git status` output appears beneath its summary — whatever git renders for that user's configuration.
- `--porcelain` switches per-repo detail to `git status --porcelain` output under graft repo header records (see the machine-readable-output delta).

**Non-Goals:**

- Any network operation, or fetching to refresh `origin/HEAD`.
- Graft-side reformulation of git status output (colors, short format, custom formats) — git renders; graft loops and frames.

## Decisions

**Stem inference by walking up to the stems directory.** Resolve the argument to an absolute path (defaulting to cwd), resolve symlinks with `filepath.EvalSymlinks`, then walk up parent directories; the first directory that is a direct child of the configured stems dir is the stem. Symlink resolution matters because the shell's logical cwd (e.g., `~/link/subdir` where `~/link` points into a stem) never traverses the stems dir by name. No git calls needed — membership is determined by path position, consistent with graft's filesystem-derived model. Being cd'd into a member worktree (or deeper inside one) is handled naturally: walking up from `stems/<stem>/<repo>/...` lands on the stem, and nested git repos inside a worktree still resolve to the enclosing stem first.

- Alternative considered: `git rev-parse --show-toplevel` then match against `git worktree list`. Rejected: adds git invocations and breaks in non-git contexts (e.g., the stem root itself, malformed stems) where inference should still work.
- Alternative considered: accept a stem name, falling back to a path. Rejected per user decision — arguments are paths, uniformly; path semantics also make `graft status ./` (and `graft status ../<other-stem>` from a sibling) work naturally.

**Dirty detection stays `git status --porcelain`; detail is git's own output.** `Status` captures the porcelain lines per repo and uses non-emptiness for `Dirty`. In default mode, dirty repos additionally get one plain `git status` invocation whose output is shown verbatim — so user customizations of git status rendering (short format, custom display, etc.) carry through, and graft never reformulates git's output. In `--porcelain` mode the captured porcelain lines are reused; no extra invocation.

- Alternative considered: parsing the long-form `git status` output for the dirty check. Rejected: long-form wording is localized and unstable; porcelain is the stable detection surface.
- Alternative considered: graft rendering a short-format file list itself. Rejected per user decision — showing git's own output preserves user customizations and keeps graft a thin loop over repos.

**Porcelain framing reuses `list`'s `repo` record shape.** Each block is headed by `repo<TAB><stem><TAB><repo-name><TAB><path>` — identical to the `repo` record in `graft list --porcelain` — followed by the repo's `git status --porcelain` lines verbatim. Git status lines never start with `repo\t`, so blocks delimit unambiguously without extra separators. Clean repos contribute only the header.

**Inference failure is a hard error.** When the resolved path is not inside any stem — bare invocation outside a stem, a nonexistent path, or a path that is not a regular directory — `status` exits non-zero with the stems directory path and the available stem names (mirroring `graft list`). Empty stems dir says so explicitly. The stems directory itself is ambiguous (it contains many stems) and errors the same way.

**File detail only for dirty repos in default mode; porcelain mode loops all repos.** In default mode clean repos stay one-line, keeping output scannable. Porcelain mode is for scripts, so every repo gets its header record even when clean — absence of changes is information too.

## Risks / Trade-offs

- [cwd deep inside a stem (e.g., a vendored directory) still walks up correctly, but a stem nested inside another stem's path could shadow] → only direct children of the stems dir qualify; stems dir is configured and stems are one level deep by construction.
- [Removing stem-name arguments breaks `graft status <stem>` muscle memory] → pre-release command surface; the error for a non-path argument names the available stems, and bare `graft status` covers the common case.
- [Porcelain paths with unusual characters render raw] → lines are printed verbatim from git; consumers of the human format are users, not scripts.
- [Stem name inferred from directory traversal vs. manifest] → the stem name comes from the directory name under stems dir, which is the same key `Find` uses; manifest content is untouched.

## Migration Plan

None — graft is pre-release and the status command ships with the new argument semantics from the start. Human status output is not a versioned contract; only the porcelain records (`graft list --porcelain`, and the status framing added here) are.

## Open Questions

None.
