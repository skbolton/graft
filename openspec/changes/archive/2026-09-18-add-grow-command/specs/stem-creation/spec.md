## Purpose

Defines the `graft grow` command: creating a stem — a named directory of worktrees across selected source repos, all on the same branch — including repo selection modes, base ref selection, failure policy, and stem-path output.

## ADDED Requirements

### Requirement: Grow creates a stem of worktrees
`graft grow <branch>` SHALL create a directory named exactly after the branch under the configured stems directory, containing one git worktree per selected source repo, each checked out to `<branch>`.

#### Scenario: Growing a stem
- **WHEN** the user runs `graft grow feat-x --sources project_a,project_b`
- **THEN** a stem directory `<stems>/feat-x` is created with `project_a/` and `project_b/` worktrees, both on branch `feat-x`

#### Scenario: Existing stem name
- **WHEN** `graft grow` is invoked with a branch whose stem directory already exists under the stems directory
- **THEN** graft exits non-zero with an error identifying the existing stem, before creating anything

### Requirement: Branch names must be valid directory names
Branch names containing characters invalid for directory names (e.g. `/`) SHALL be rejected with a clear error before any worktree is created.

#### Scenario: Slashed branch rejected
- **WHEN** the user runs `graft grow feat/login`
- **THEN** graft exits non-zero with an error explaining the branch must be usable as a directory name, and no stem or worktrees are created

### Requirement: Repo selection modes
`graft grow` SHALL support selecting repos via `-c/--collection <name>`, via `-s/--sources <repo,...>` (comma-separated repo directory names), and via cwd context (run from inside a source repo → single-repo stem). When no selection context exists, graft SHALL exit non-zero with a hint explaining the selection options.

#### Scenario: Collection selection
- **WHEN** `graft grow feat-x -c backend` runs
- **THEN** exactly the repos in the `backend` collection are grafted

#### Scenario: Explicit sources selection
- **WHEN** `graft grow feat-x -s project_a,project_c` runs
- **THEN** exactly those named repos are grafted

#### Scenario: Cwd-aware single-repo stem
- **WHEN** `graft grow feat-y` runs from inside a source repository, with no collection or `--sources` flags
- **THEN** a single-repo stem is created for that repo

#### Scenario: No selection context
- **WHEN** `graft grow feat-x` runs with no collection, no `--sources`, and outside any source repo
- **THEN** graft exits non-zero with a hint naming the selection options (`-c`, `-s`, or run from inside a source repo)

#### Scenario: Unresolvable repo name
- **WHEN** a selected repo name (via `-s` or a collection) matches no discovered source repo
- **THEN** it is reported as a failure for that name in the create summary and the remaining repos proceed

### Requirement: Worktrees based on locally known origin/HEAD
Each worktree SHALL be based on the repo's `origin/HEAD` as known locally. Graft SHALL perform no network operations during grow; sources are expected to be current. If a repo's `origin/HEAD` cannot be resolved, that repo SHALL fail with an actionable error.

#### Scenario: Worktree based on origin/HEAD
- **WHEN** a source repo has a locally resolvable `origin/HEAD`
- **THEN** its worktree is created from that ref on the new branch

#### Scenario: origin/HEAD unresolvable
- **WHEN** a source repo's `origin/HEAD` cannot be resolved (e.g. never set)
- **THEN** that repo is reported as failed in the create summary with an actionable error (e.g. run `git remote set-head origin -a`), and other repos proceed

### Requirement: Warn-and-continue failure policy
A per-repo failure during stem creation (git error, hook failure, unresolvable base ref, unknown name) SHALL NOT abort creation of the remaining repos. All per-repo outcomes SHALL be reported in the create summary, and the stem SHALL exist with the repos that succeeded. The stem directory and manifest SHALL be created even when every repo fails, so the outcome is inspectable.

#### Scenario: One repo fails mid-create
- **WHEN** `project_a` creates successfully but `project_b`'s worktree creation fails
- **THEN** `project_a`'s worktree remains, `project_b`'s failure is reported in the summary, and the command exits non-zero

#### Scenario: Per-repo post-checkout hook failure
- **WHEN** a source repo's git `post-checkout` hook exits non-zero during worktree creation
- **THEN** the worktree still exists (git behavior), the failure is reported as a warning in the summary, and other repos are unaffected

#### Scenario: Same branch already checked out
- **WHEN** a source repo already has the target branch checked out in another worktree
- **THEN** that repo is reported as failed in the create summary with git's error surfaced, and other repos proceed

### Requirement: Colocated stem manifest
Each grown stem SHALL contain a `.graft/manifest.toml` with exactly this schema:

```toml
branch = "feat-x"
collection = "backend"              # empty string when ad-hoc
created_at = "2026-09-18T12:00:00Z" # RFC 3339

[[repos]]
name = "project_a"
source = "/home/user/c/sources/project_a"

[[repos]]
name = "project_b"
source = "/home/user/c/sources/project_b"
```

Graft SHALL maintain no central state file.

#### Scenario: Manifest contents after grow
- **WHEN** a stem is created
- **THEN** `.graft/manifest.toml` inside the stem matches the schema above, with one `[[repos]]` entry per successfully created worktree

### Requirement: Stem path on stdout
On completion (success or partial success), grow SHALL print the absolute stem directory path as the final line of stdout. Human progress and the create summary SHALL go to stderr.

#### Scenario: Launcher consumes the stem path
- **WHEN** an external tool runs `graft grow feat-x -c backend` and reads the final line of stdout
- **THEN** it receives the absolute path to the stem directory
