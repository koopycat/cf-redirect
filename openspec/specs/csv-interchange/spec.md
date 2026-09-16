# CSV Interchange Specification

## Purpose

Define the intentionally narrow two-column CSV format used to import redirect upserts and export the configured list without implying full-fidelity backup or synchronization.

## Requirements

### Requirement: Strict two-column CSV input

The system SHALL accept comma-separated or semicolon-separated CSV containing exactly `source` and `target` fields per row. A lowercase `source,target` header is optional, and one delimiter SHALL be used consistently throughout a file.

#### Scenario: Headerless comma-separated input

- **WHEN** valid rows contain two comma-separated fields without a header
- **THEN** every row is parsed as a redirect

#### Scenario: Semicolon-separated input with header

- **WHEN** a file begins with `source;target` and valid semicolon-separated rows follow
- **THEN** the header is omitted from imported data and all data rows are parsed

#### Scenario: Empty or malformed input

- **WHEN** input is empty, mixes an invalid structure, or contains a row without exactly two fields
- **THEN** the import is rejected with row or format context

#### Scenario: Duplicate source appears

- **WHEN** more than one imported row has the same exact source
- **THEN** the import is rejected rather than choosing one row

#### Scenario: Imported URL is invalid

- **WHEN** a row contains a source or target that violates redirect URL constraints
- **THEN** the import is rejected with the row number

### Requirement: Import is upsert-only

The system SHALL match imported rows to current redirects by exact source. It SHALL add missing sources, update only the target of existing sources, report content-equivalent entries as skipped, and SHALL never delete a current redirect merely because it is absent from the import.

#### Scenario: Imported source is new

- **WHEN** an imported source is not in the current list
- **THEN** the plan adds it with query-string preservation enabled

#### Scenario: Imported source already exists with a different target

- **WHEN** an imported source exactly matches a current item but its target differs
- **THEN** the plan updates the target while preserving the item's comment, status code, options, and existing source

#### Scenario: Imported row is unchanged

- **WHEN** an imported source and target already match current content
- **THEN** no mutation is planned for that row
- **AND** it is counted as skipped existing

#### Scenario: Current source is omitted

- **WHEN** a current redirect has no corresponding row in the import
- **THEN** it remains untouched and no delete is planned

### Requirement: CSV export

The system SHALL export a header followed by one `source,target` record per redirect. CSV export SHALL intentionally omit IDs, comments, status codes, and redirect options.

#### Scenario: Export to standard output

- **WHEN** the export destination is `-`
- **THEN** the system writes the two-column CSV to standard output

#### Scenario: Export to a file

- **WHEN** a filesystem destination is selected and the complete CSV is written successfully
- **THEN** the system atomically replaces the destination with the completed export

#### Scenario: Export fails before completion

- **WHEN** the CSV cannot be completely written or synchronized
- **THEN** an existing destination is not replaced by partial output

#### Scenario: CSV is re-imported

- **WHEN** exported CSV is imported into a list
- **THEN** only source and target semantics are available from the file
- **AND** the file is not treated as a full backup of redirect metadata
