# Terminal Interface Specification

## Purpose

Define consistent, review-first behavior across the Cobra command-line interface and keyboard-driven Bubble Tea interface.

## Requirements

### Requirement: CLI mutation guardrails

The CLI SHALL render every mutation plan before deciding whether to apply it. Interactive application SHALL require affirmative confirmation. Non-interactive application SHALL require `--yes`, while `--dry-run` SHALL stop after the preview.

#### Scenario: Interactive mutation is confirmed

- **WHEN** an interactive user reviews a non-empty plan and answers `y` or `yes`
- **THEN** the system applies the reviewed plan

#### Scenario: Interactive mutation is declined

- **WHEN** the user gives any other confirmation response
- **THEN** the system reports cancellation and submits no mutation

#### Scenario: Non-interactive mutation lacks a guard flag

- **WHEN** a non-interactive invocation produces a non-empty mutation plan without `--yes` or `--dry-run`
- **THEN** the system rejects application after showing the plan

#### Scenario: Dry run is requested

- **WHEN** `--dry-run` is supplied
- **THEN** the system displays the complete plan and submits no mutation

#### Scenario: Empty plan is produced

- **WHEN** planning produces no changes
- **THEN** the system displays the zero-change plan and does not ask for confirmation

### Requirement: CLI apply progress and recovery reporting

The CLI SHALL report meaningful apply stages and phase outcomes without exposing credentials. Interactive terminals MAY update one live progress line; non-terminal output SHALL remain stable and line-oriented.

#### Scenario: Apply succeeds

- **WHEN** all plan phases complete
- **THEN** the CLI reports each applied phase, item count, and operation ID

#### Scenario: Apply fails

- **WHEN** a phase fails
- **THEN** the CLI reports completed and failed phase results
- **AND** it emits applicable partial-apply or rate-limit recovery guidance

### Requirement: TUI launch behavior

Running `cf-redirect` without a subcommand in an interactive terminal or invoking `tui` SHALL open the full-screen redirect editor. A non-interactive invocation without a command or a request to launch the TUI SHALL fail with actionable guidance.

#### Scenario: Root command runs interactively

- **WHEN** the root command has an interactive input and output terminal
- **THEN** it launches the TUI for the configured list

#### Scenario: Root command runs non-interactively

- **WHEN** no subcommand is supplied without an interactive terminal
- **THEN** it rejects the invocation and directs the user to command help

### Requirement: Redirect browsing and filtering

The TUI SHALL display the configured list, support keyboard navigation and pagination, and filter by source, target, or comment. Narrow terminals SHALL use a stacked redirect layout instead of clipping essential URLs.

#### Scenario: Filter is entered

- **WHEN** the user searches in the redirect list
- **THEN** visible items are filtered against source, target, and comment

#### Scenario: Terminal is narrow

- **WHEN** terminal width is below the compact layout threshold
- **THEN** redirects render using the narrow stacked layout

#### Scenario: List reloads

- **WHEN** remote state is reloaded after an operation
- **THEN** the active filter is retained
- **AND** the previously selected item remains selected when its ID is still visible

### Requirement: TUI add, edit, and delete review

The TUI SHALL allow keyboard-driven add, edit, and delete planning. It SHALL show a deterministic review screen and SHALL apply a plan only when the user presses `y`; escape SHALL cancel review.

#### Scenario: Add or edit form is submitted

- **WHEN** valid source and target input is completed
- **THEN** the TUI opens a plan review instead of mutating immediately

#### Scenario: Delete is requested

- **WHEN** the user requests deletion of the selected redirect
- **THEN** the TUI opens an explicit deletion plan for that item's ID

#### Scenario: Review receives an unrelated key

- **WHEN** the plan review is visible and the user presses a key other than `y` or cancel
- **THEN** the plan is not applied

### Requirement: TUI apply interruption is honest

The TUI SHALL show live progress while applying. Pressing quit or escape during apply SHALL cancel only the local wait and SHALL warn that a submitted Cloudflare operation may still complete.

#### Scenario: User interrupts an apply

- **WHEN** the user presses quit or escape while execution is waiting
- **THEN** the local context is cancelled
- **AND** the TUI tells the user that Cloudflare may still finish submitted work
- **AND** it directs the user to exit and verify current state

#### Scenario: Apply finishes or fails

- **WHEN** execution returns success or failure
- **THEN** the TUI reloads remote state before another mutation can be attempted

### Requirement: Persistent readable errors

The TUI SHALL render errors as a dedicated, sanitized view that remains visible until acknowledged.

#### Scenario: An operation returns an error

- **WHEN** loading, planning, or applying produces an error
- **THEN** the full safe error text remains visible until the user presses a key
- **AND** a background reload does not silently erase an unacknowledged apply error
