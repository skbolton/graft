## Purpose

Defines the stable machine-readable output contract consumed by external tools — the user's fzf/tmux launcher, fuzzy-finder scripts, and any downstream automation.

## Requirements

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

### Requirement: Porcelain status framing
`graft status --porcelain` SHALL emit, for each member repo of the stem, one graft header record in the same shape as `graft list --porcelain`'s `repo` record, followed by that repo's `git status --porcelain` lines verbatim:

```
repo<TAB><stem-name><TAB><repo-name><TAB><absolute-path>
<git status --porcelain lines for that repo>
```

The next `repo<TAB>` record or end of output delimits each block. Git status lines SHALL be emitted verbatim, unmodified by graft. A clean repo contributes only its header record. This framing SHALL be stable within a major version; any change SHALL require a major version bump.

#### Scenario: Script attributes status blocks
- **WHEN** an external script parses `graft status --porcelain` output
- **THEN** each block of git status lines is attributable to a member repo via the preceding `repo` header record, and clean repos are identifiable by their empty blocks

#### Scenario: Framing stability
- **WHEN** a consumer parses `graft status --porcelain` output across graft patch releases
- **THEN** the header record shape and the verbatim git lines contract are unchanged within a major version
