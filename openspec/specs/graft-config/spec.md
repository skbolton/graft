## Purpose

Defines the global configuration file for graft: where source repositories live, where stems are created, and named collections of source repos.

## Requirements

### Requirement: Global config location and loading
Graft SHALL load its global configuration from `~/.config/graft/config.toml`
(XDG-aware: `$XDG_CONFIG_HOME/graft/config.toml`). If the file does not exist, commands
that require configuration SHALL fail with a message identifying the expected path. The
file uses this concrete schema:

```toml
sources = ["~/repos"]       # list of directories containing source repo checkouts
stems = "~/workspaces"      # directory where stem directories are created

[collections]
backend = ["project_a", "project_b"]
```

Path values SHALL support `~` expansion.

#### Scenario: Config file present
- **WHEN** a valid `config.toml` exists at the config location
- **THEN** all graft commands read their settings from it

#### Scenario: Config file missing
- **WHEN** graft runs and no config file exists
- **THEN** graft exits non-zero with an error identifying the expected config path

### Requirement: Sources configuration
The config SHALL define one or more `sources` directories that contain pristine source repository checkouts. `sources` MUST be a list of directory paths.

#### Scenario: Configuring sources
- **WHEN** the config defines `sources = ["~/repos"]`
- **THEN** graft treats repositories found under that directory as graftable source repos

### Requirement: Stems directory configuration
The config SHALL define a `stems` directory where workstream directories are created.

#### Scenario: Configuring stems directory
- **WHEN** the config defines `stems = "~/workspaces"`
- **THEN** newly grown stems are created under that directory

### Requirement: Collections configuration
The config SHALL support named collections under a `[collections]` table, each mapping a collection name to a list of source repo names (by directory name).

#### Scenario: Defining a collection
- **WHEN** the config defines `[collections] backend = ["project_a", "project_b"]`
- **THEN** `graft grow --collection backend` selects exactly those repos

#### Scenario: Unknown collection
- **WHEN** `graft grow --collection <name>` names a collection not present in `[collections]`
- **THEN** graft exits non-zero with an error listing the available collection names

#### Scenario: Collection references unknown repo
- **WHEN** a collection lists a repo name that discovery does not find under any sources directory
- **THEN** that repo is reported as a per-repo failure during grow (see stem-creation's warn-and-continue policy) and other repos proceed
