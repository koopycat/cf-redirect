# Redirect Querying Specification

## Purpose

Define read-only discovery of redirects and Cloudflare bulk-operation status for interactive users, scripts, and diagnostics.

## Requirements

### Requirement: Complete redirect listing

The system SHALL list every item from the explicitly configured Cloudflare account and Bulk Redirect List. Human-readable results SHALL be sorted by source, and supported output formats SHALL be table, JSON, and CSV.

#### Scenario: Default list output

- **WHEN** the user runs `list` without selecting a format
- **THEN** the system emits a table containing source, target, effective status code, and comment for every redirect
- **AND** rows are ordered by source

#### Scenario: Machine-readable list output

- **WHEN** the user selects JSON or CSV output
- **THEN** the system emits all redirects in the requested machine-readable format

#### Scenario: Unknown format

- **WHEN** the user selects an unsupported list format
- **THEN** the system rejects the request and identifies the supported formats

### Requirement: Redirect search

The system SHALL search source, target, and comment fields using a case-insensitive substring query and SHALL support the same table, JSON, and CSV formats as listing.

#### Scenario: Query matches any searchable field

- **WHEN** a search query occurs in a redirect's source, target, or comment with any letter case
- **THEN** that redirect is included in the search result

#### Scenario: Query has no matches

- **WHEN** no redirect contains the search query
- **THEN** the system emits a valid empty result in the requested format

### Requirement: Safe terminal and structured output

The system SHALL remove terminal control characters from Cloudflare-provided source, target, comment, operation, and error text before displaying or serializing it.

#### Scenario: Remote text contains terminal controls

- **WHEN** Cloudflare returns a field containing terminal control sequences
- **THEN** output contains the field's safe text without those controls
- **AND** the controls cannot alter the user's terminal state

### Requirement: Bulk operation inspection

The system SHALL retrieve a specified Cloudflare bulk operation in the configured account and emit its normalized status as JSON.

#### Scenario: Operation status is requested

- **WHEN** the user runs `status` with an operation ID
- **THEN** the system retrieves that account-scoped operation and emits its status as JSON

#### Scenario: Operation ID is missing

- **WHEN** the user invokes `status` without exactly one operation ID
- **THEN** the command is rejected before an API request is made
