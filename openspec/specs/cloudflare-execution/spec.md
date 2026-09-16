# Cloudflare Execution Specification

## Purpose

Define safe interaction with Cloudflare's Bulk Redirect List API, including pagination, revalidation, explicit mutation ordering, asynchronous completion, bounded retries, and partial-failure reporting.

## Requirements

### Requirement: Focused and explicit API scope

The system SHALL include the configured account ID and list ID on every list-item operation. It SHALL use GET to list items, POST to create or upsert items, and DELETE with explicit item IDs to remove items. It SHALL NOT use Cloudflare's replace-all `PUT /items` endpoint.

#### Scenario: Items are listed

- **WHEN** current redirects are requested
- **THEN** the request addresses the explicitly configured account and list

#### Scenario: Items are deleted

- **WHEN** a delete mutation is submitted
- **THEN** its body contains `items` with a non-empty `id` for every requested deletion

#### Scenario: Empty delete is attempted

- **WHEN** a delete request has no item IDs or contains an empty ID
- **THEN** the client rejects it before sending the request

### Requirement: Cursor pagination

The system SHALL retrieve list items with Cloudflare cursor pagination using `per_page=500` until no next cursor remains. It SHALL reject repeated or reused cursors.

#### Scenario: Multiple pages are available

- **WHEN** a list response includes a new `after` cursor
- **THEN** the system requests the next page with that cursor and combines its items with prior pages

#### Scenario: Cursor repeats

- **WHEN** Cloudflare repeats the current cursor or returns a previously observed cursor
- **THEN** listing fails instead of entering an unbounded pagination loop

### Requirement: Plans are revalidated before apply

The system SHALL fetch fresh list state immediately before applying a non-empty plan and SHALL reject stale, malformed, duplicate, or colliding changes before submitting a mutation.

#### Scenario: Planned item changed remotely

- **WHEN** an update or deletion's prior item content no longer matches the current item with that ID
- **THEN** apply fails as stale before any mutation is submitted

#### Scenario: Planned source now exists

- **WHEN** an add or replacement source collides with fresh list state
- **THEN** apply fails before mutation

#### Scenario: Empty plan is applied

- **WHEN** a plan has no changes
- **THEN** execution returns successfully without contacting a mutation endpoint

### Requirement: Availability-aware mutation ordering

The system SHALL apply source-changing updates and explicit deletes before their dependent creates. An update whose source is unchanged SHALL use Cloudflare's POST upsert without deleting first, avoiding a redirect availability gap. The system SHALL clear item IDs from create payloads.

#### Scenario: Target-only update is applied

- **WHEN** an update retains its source
- **THEN** the system submits it through POST upsert without a preceding delete

#### Scenario: Source-changing update is applied

- **WHEN** an update changes its source
- **THEN** the system deletes the old explicit item ID and waits for completion before creating the replacement

#### Scenario: Delete phase fails

- **WHEN** a required delete does not complete successfully
- **THEN** dependent creates are not submitted

### Requirement: Bounded sequential batches

The system SHALL submit at most 500 deletes or creates in one mutation request and SHALL wait for each asynchronous operation to complete before submitting the next dependent batch.

#### Scenario: Mutation exceeds one batch

- **WHEN** a phase contains more than 500 items
- **THEN** it is split into batches of at most 500
- **AND** each batch operation reaches completion before the next batch is sent

### Requirement: Asynchronous operation polling

The system SHALL poll each submitted bulk operation until success, reported failure, cancellation, or a bounded timeout. Polling SHALL begin immediately, back off to reduce API pressure, and provide the operation ID for recovery when completion cannot be confirmed.

#### Scenario: Operation completes

- **WHEN** polling observes a successful terminal status with no reported errors
- **THEN** the phase is recorded as completed

#### Scenario: Operation reports an error

- **WHEN** Cloudflare returns a failure status, error text, or non-zero error count
- **THEN** the operation is treated as failed even if another status field appears non-terminal

#### Scenario: Operation times out

- **WHEN** no terminal status is observed within the bounded wait
- **THEN** the system returns an error containing the operation ID and a status-check recovery command

#### Scenario: Local wait is cancelled

- **WHEN** the caller cancels while waiting
- **THEN** polling stops locally
- **AND** the system does not claim that an already submitted Cloudflare operation was cancelled remotely

### Requirement: Conservative retry behavior

The system SHALL automatically retry transient GET responses with bounded backoff. It SHALL NOT automatically replay POST or DELETE after ambiguous transport or server failures. It MAY retry a mutation only when Cloudflare explicitly rejects it with HTTP 429, using bounded attempts and delays.

#### Scenario: Read receives a transient response

- **WHEN** a GET receives HTTP 429 or a 5xx response within the retry budget
- **THEN** the system retries after bounded `Retry-After` or exponential jittered backoff

#### Scenario: Mutation receives an ambiguous failure

- **WHEN** POST or DELETE receives a network error or non-429 server failure
- **THEN** the request is not automatically replayed

#### Scenario: Mutation is rate limited

- **WHEN** POST or DELETE receives HTTP 429
- **THEN** the executor waits according to bounded server guidance or exponential delay and retries within its attempt budget

### Requirement: Partial application is observable

The system SHALL retain the result of every completed or failed phase and SHALL identify an apply as partial when an earlier phase completed but a later phase failed.

#### Scenario: Later phase fails

- **WHEN** deletion completes and the dependent create phase fails
- **THEN** the error reports partial application
- **AND** the user is instructed to inspect current list state before retrying
