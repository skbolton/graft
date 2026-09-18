## Purpose

Defines how graft discovers graftable source repositories within the configured sources directories. Discovery is one level deep by design; a repo is identified by its directory name.

## Requirements

### Requirement: Scan sources directories
Graft SHALL scan each configured sources directory for git repositories exactly one level deep (direct children), skipping hidden directories and common vendor directories (e.g. `node_modules`).

#### Scenario: Flat sources layout
- **WHEN** a sources directory contains `project_a/` and `project_b/`, both git repositories
- **THEN** discovery reports both as available source repos by directory name

#### Scenario: Hidden and vendor directories excluded
- **WHEN** a sources directory contains hidden directories or `node_modules`
- **THEN** those are not reported as source repos

#### Scenario: Nested repos are out of scope
- **WHEN** a sources directory contains a directory whose child is itself a git repository
- **THEN** only the direct child is evaluated; deeper nesting is not scanned in MVP

### Requirement: Repo names are directory names
A discovered source repo SHALL be identified by its directory name, which is the value used in collections, `--sources` flags, and the stem manifest.

#### Scenario: Selecting a repo by name
- **WHEN** a user passes a repo directory name via `--sources` or a collection references it
- **THEN** graft resolves it to the discovered source repo path

### Requirement: Duplicate repo names in a selection
Graft SHALL tolerate a repo name appearing more than once within a single selection (e.g. a collection plus `--sources` overlap) by treating it as one repo. Remote-URL-level deduplication across differently named checkouts is out of scope for MVP.

#### Scenario: Overlapping selections
- **WHEN** a selection resolves to the same repo name twice
- **THEN** the repo is grafted exactly once
