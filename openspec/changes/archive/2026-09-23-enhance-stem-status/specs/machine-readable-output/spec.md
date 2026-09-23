## ADDED Requirements

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
