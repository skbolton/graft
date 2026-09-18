## Purpose

Defines graft's lifecycle hook: a `postcreate` command executed after stem creation, with context delivered via environment variables.

## Requirements

### Requirement: Postcreate hook
Graft SHALL execute a `postcreate` hook command defined in the global config after all repos in a stem have been attempted. The hook runs with the stem directory as its working directory.

#### Scenario: Hook fires after creation
- **WHEN** a stem is created and the config defines a `postcreate` hook
- **THEN** the hook runs after all repos have been attempted, with cwd set to the stem directory

### Requirement: Hook context via environment variables
The `postcreate` hook SHALL receive workspace context as environment variables: `GRAFT_WORKSPACE` (stem directory name), `GRAFT_PATH` (absolute stem path), `GRAFT_BRANCH` (branch name), `GRAFT_COLLECTION` (collection name, empty if ad-hoc), and `GRAFT_REPOS` (newline-separated repo names of successfully created worktrees). Hook commands SHALL NOT rely on positional placeholder interpolation of user data.

#### Scenario: Postcreate hook generates workspace notes
- **WHEN** a `postcreate` hook reads `GRAFT_COLLECTION` and `GRAFT_PATH`
- **THEN** the hook can act on the collection identity and stem location (e.g. generating an AGENTS.md next to the projects) without parsing positional arguments

### Requirement: Postcreate failure is a warning
If the `postcreate` hook fails, graft SHALL report the failure as a warning; the created stem SHALL remain intact.

#### Scenario: Postcreate exits non-zero
- **WHEN** the `postcreate` hook exits non-zero after a successful creation
- **THEN** graft reports the failure as a warning and the stem remains fully intact

### Requirement: Skip-hooks escape hatch
Graft SHALL provide a `--no-hooks` flag to skip the `postcreate` hook for a single invocation.

#### Scenario: Creating without hooks
- **WHEN** the user runs `graft grow feat-x ... --no-hooks`
- **THEN** the `postcreate` hook does not run
