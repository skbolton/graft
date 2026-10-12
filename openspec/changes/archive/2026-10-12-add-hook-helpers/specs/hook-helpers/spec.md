## Purpose

Defines graft's helper commands that a repo's own `post-checkout` hook may call to bootstrap new worktrees: detecting worktree creation and copying untracked setup files (environment files, build artifacts) from the repo's main worktree into the new worktree. Graft never reads or executes per-repo configuration; these helpers are opt-in tools the repos choose to invoke.

## ADDED Requirements

### Requirement: Creation detection helper

Graft SHALL provide `graft hook on-create <old-sha>` that exits 0 when `<old-sha>` is the all-zeros sha (git's worktree-creation marker in `post-checkout` hooks) and exits non-zero for any other value. A missing argument SHALL be a usage error with a non-zero exit, not a silent false result.

#### Scenario: Called with the all-zeros sha

- **WHEN** the repo's `post-checkout` hook calls `graft hook on-create 0000000000000000000000000000000000000000`
- **THEN** graft exits 0

#### Scenario: Called with a real sha (ordinary branch switch)

- **WHEN** the hook calls `graft hook on-create <some-actual-sha>`
- **THEN** graft exits non-zero

#### Scenario: Called without the old-sha argument

- **WHEN** the hook calls `graft hook on-create` with no argument
- **THEN** graft prints a usage error and exits non-zero

### Requirement: Copy from the main worktree

Graft SHALL provide `graft hook copy <path>...` that copies each given path from the repo's main worktree (the source checkout) into the caller's current directory, preserving each path's relative position. Source discovery SHALL be derived from git itself so the helper works in any worktree of the repo, whether created by graft or by plain git, without graft state.

#### Scenario: Copying an untracked file into a new worktree

- **WHEN** the main worktree contains an untracked `.envrc` and the hook calls `graft hook copy .envrc` from a worktree of the same repo
- **THEN** the worktree contains `.envrc` with the same content as the main worktree's copy

#### Scenario: Missing source entry

- **WHEN** a requested path does not exist in the main worktree
- **THEN** graft skips it quietly and exits 0 (an optional file like `.envrc` is the normal case)

#### Scenario: Source and destination are the same worktree

- **WHEN** `graft hook copy` is invoked from the repo's main worktree itself
- **THEN** graft does not copy onto itself, reports this on stderr, and exits 0

### Requirement: Directory copying

`graft hook copy` SHALL copy directories recursively. When the filesystem supports copy-on-write/reflink graft SHALL use it for efficiency, falling back to a plain recursive copy otherwise. Symbolic links encountered during a directory copy SHALL be preserved as symbolic links, not dereferenced.

#### Scenario: Copying an ignored build-artifact directory

- **WHEN** the main worktree contains a `node_modules` directory and the hook calls `graft hook copy node_modules` from a worktree
- **THEN** the directory tree is copied into the worktree, and the command completes without re-walking the tree through an unconditional full copy where reflink is supported

#### Scenario: Symlink inside a copied directory

- **WHEN** a copied directory contains a symbolic link
- **THEN** the corresponding path in the destination is also a symbolic link pointing to the same target

### Requirement: Copy failure reporting

When a copy operation fails (for example, permission denied), `graft hook copy` SHALL report the failure on stderr and exit non-zero. It SHALL NOT abort remaining copies before reporting: all requested paths are attempted.

#### Scenario: One of several copies fails

- **WHEN** the hook calls `graft hook copy .envrc unreadable-dir` and the directory copy fails
- **THEN** graft attempts the remaining paths, reports the failed one on stderr, and exits non-zero

### Requirement: Helpers independent of graft state

The hook helpers SHALL NOT read graft configuration, manifests, or any graft state. Their behavior SHALL depend only on the git repository containing the caller's working directory.

#### Scenario: Helper run in a worktree created outside graft

- **WHEN** `graft hook copy .envrc` runs in a worktree created by plain `git worktree add`
- **THEN** the copy behaves identically to a worktree created by graft
