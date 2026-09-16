# Redirect Mutations Specification

## Purpose

Define validation and deterministic planning for adding, editing, bulk-editing, and explicitly deleting redirects while preserving unrelated list data.

## Requirements

### Requirement: Redirect URL validation

The system SHALL accept HTTP or HTTPS source URLs with a hostname, including schemeless sources, and SHALL require targets to be absolute HTTP or HTTPS URLs with a hostname. Sources and targets SHALL be non-empty, free of surrounding whitespace, fragments, user information, and control characters.

#### Scenario: Schemeless source is valid

- **WHEN** a source such as `example.com/old/` and an absolute HTTPS target are supplied
- **THEN** the redirect passes local URL validation

#### Scenario: Invalid source or target is rejected

- **WHEN** either URL violates a local URL constraint
- **THEN** planning fails before any mutation is submitted

#### Scenario: Unsupported status code is present

- **WHEN** a redirect has a non-zero status code other than 301, 302, 307, or 308
- **THEN** validation rejects it

### Requirement: Add planning

The system SHALL plan an add only when its exact source is absent. New redirects SHALL default to status code 301 and preservation of the original query string.

#### Scenario: New source is added

- **WHEN** the user supplies a valid source that does not exist and a valid target
- **THEN** the plan contains one add with query-string preservation enabled

#### Scenario: Source already exists

- **WHEN** the current list already contains the exact source
- **THEN** add planning fails without changing that redirect

### Requirement: Edit planning preserves redirect data

The system SHALL identify edits by the current redirect's explicit item ID and SHALL change only the requested source and target. It SHALL preserve the status code, comment, and every redirect option.

#### Scenario: Target-only edit

- **WHEN** an existing redirect receives a different target and keeps its source
- **THEN** the plan contains one update preserving its ID-derived identity, options, status code, and comment

#### Scenario: Source rename

- **WHEN** an existing redirect receives a new source that is not used by another item
- **THEN** the plan contains one update from the old source to the new source

#### Scenario: Replacement source collides

- **WHEN** an edit's new source belongs to another current item
- **THEN** planning rejects the edit

#### Scenario: Edit makes no content change

- **WHEN** the requested source and target equal the existing values
- **THEN** the plan contains no mutation

### Requirement: Bulk query-string option editing

The system SHALL support setting `preserve query string` to an explicit true or false value across all redirects while preserving every other field. It SHALL plan updates only for entries whose current value differs and SHALL report matching entries as skipped.

#### Scenario: Bulk option is enabled

- **WHEN** bulk editing requests query-string preservation and some entries have it disabled
- **THEN** only those entries receive update changes
- **AND** already enabled entries are counted as skipped existing

#### Scenario: Bulk option value is omitted

- **WHEN** `edit-all` is invoked without explicitly supplying `--preserve-query-string`
- **THEN** the command is rejected without planning a mutation

### Requirement: Explicit deletion planning

The system SHALL plan individual deletion by exact source and explicit Cloudflare item ID. Clearing a list SHALL produce one explicit-ID deletion for every current item and SHALL NOT infer deletion from absence in another dataset.

#### Scenario: Exact source is deleted

- **WHEN** the requested source exactly matches one item with a non-empty ID
- **THEN** the plan contains one deletion carrying that item's ID

#### Scenario: Delete source is absent

- **WHEN** no item has the exact requested source
- **THEN** deletion is rejected without selecting a partial or normalized match

#### Scenario: Entire list is cleared

- **WHEN** the user requests `clear`
- **THEN** the plan contains an explicit-ID deletion for every current item

#### Scenario: Current item lacks an ID

- **WHEN** a mutation plan depends on a current item with no Cloudflare item ID
- **THEN** planning fails rather than constructing an unsafe deletion

### Requirement: Deterministic plans

The system SHALL represent mutations as add, update, and delete changes with both sides present where applicable. Plans SHALL provide deterministic counts and ordering so users can review the intended effects.

#### Scenario: Plan is rendered

- **WHEN** a mutation is planned
- **THEN** the preview reports add, update, delete, and applicable skipped counts
- **AND** each change is marked with `+`, `~`, or `-`

#### Scenario: Plan output fails

- **WHEN** the complete plan cannot be written to the selected output
- **THEN** the system does not proceed to confirmation or application
