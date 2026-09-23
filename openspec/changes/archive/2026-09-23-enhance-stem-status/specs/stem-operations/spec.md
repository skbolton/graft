## MODIFIED Requirements

### Requirement: Per-stem status
`graft status` SHALL report per-repo health of a stem: dirty worktrees (uncommitted or untracked changes) and commits ahead of/behind the repo's base ref, where the base ref is `origin/HEAD` as known locally — consistent with what `graft grow` would base a new worktree on. The stem SHALL be resolved from an optional path argument, defaulting to the current working directory: the path may be the stem root, a member worktree, or any directory inside the stem. Stem names SHALL NOT be accepted as the argument. The report SHALL show a one-line summary per repo (clean/dirty, ahead/behind counts) and, beneath it, the full `git status` output for dirty repos — as git renders it under the user's git configuration. With `--porcelain`, graft SHALL emit one `repo` header record per member repo followed by that repo's `git status --porcelain` output, in the framing defined by the machine-readable-output capability. No network operation SHALL be performed.

#### Scenario: Dirty repo surfaced
- **WHEN** a stem contains a repo with uncommitted or untracked changes
- **THEN** `graft status` marks that repo dirty in the summary and shows that repo's `git status` output beneath it, and lists the others cleanly

#### Scenario: User git customizations honored
- **WHEN** the user has customized their git status rendering in git config
- **THEN** the per-repo `git status` output shown by `graft status` reflects that customization, since graft displays git's output rather than reformulating it

#### Scenario: Branch drift surfaced
- **WHEN** a stem contains a repo whose branch is ahead of or behind its base ref (`origin/HEAD`)
- **THEN** `graft status` reports the ahead and behind counts for that repo

#### Scenario: Status from inside a stem
- **WHEN** `graft status` is run with no argument, or with any path inside a stem (stem root or member worktree), from that location
- **THEN** graft reports the status of that stem's repos

#### Scenario: Status outside a stem
- **WHEN** `graft status` is run with no argument, or with a path that is not inside any stem
- **THEN** graft exits non-zero with an error that names the available stems

#### Scenario: Porcelain output attributed to repos
- **WHEN** an external script parses `graft status --porcelain` output
- **THEN** it can attribute each block of git status lines to a member repo via the graft `repo` header record that precedes it
