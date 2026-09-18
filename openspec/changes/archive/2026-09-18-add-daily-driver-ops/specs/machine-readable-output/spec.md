## Purpose

Defines the stable machine-readable output contract consumed by external tools — the user's fzf/tmux launcher, fuzzy-finder scripts, and any downstream automation.

## ADDED Requirements

### Requirement: Porcelain output format
`graft list --porcelain` SHALL emit tab-separated records on stdout, one record per line, in these exact record types:

```
stem<TAB><name><TAB><absolute-path>
repo<TAB><stem-name><TAB><repo-name><TAB><absolute-path>
collection<TAB><name><TAB><repo-name,repo-name,...>
source<TAB><name><TAB><absolute-path>
```

Records MAY be emitted in any order. When `--porcelain` is given, no human-oriented output SHALL appear on stdout.

#### Scenario: Launcher consumes stem list
- **WHEN** an external fzf launcher runs `graft list --porcelain` and parses `stem` records
- **THEN** it receives one record per stem with its name and absolute path, with no human-formatted lines to filter out

#### Scenario: Launcher finds a stem's repos
- **WHEN** an external tool parses `repo` records for a given stem name
- **THEN** it receives one record per repo in that stem with the repo name and worktree path

### Requirement: Collections and sources enumeration
`graft list --porcelain` SHALL include one `collection` record per configured collection and one `source` record per discovered source repo, so external fuzzy-finder utilities can drive repo selection without duplicating graft's discovery logic.

#### Scenario: Fuzzy finder lists repos
- **WHEN** an external utility parses `source` records
- **THEN** it receives every discovered source repo's name and path, matching what `graft grow` would graft from

#### Scenario: Launcher offers collections
- **WHEN** an external utility parses `collection` records
- **THEN** it receives each configured collection name and its repo names

### Requirement: Stable record format
Porcelain output SHALL use the record format above; field order and separators SHALL NOT change within a major version. Any format change SHALL require a major version bump.

#### Scenario: Format stability
- **WHEN** a consumer parses porcelain output
- **THEN** record types, field order, and the tab separator are unchanged within a major version
