## MODIFIED Requirements

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
