# Configuration and Authentication Specification

## Purpose

Define how cf-redirect identifies its single Cloudflare account and Bulk Redirect List while obtaining API credentials without persisting or exposing tokens.

## Requirements

### Requirement: Account and list configuration resolution

The system SHALL require a non-empty Cloudflare account ID and Bulk Redirect List ID for list-scoped operations. It SHALL resolve each value independently in this precedence order: command flag, environment variable, persisted configuration.

#### Scenario: Flags override all other configuration

- **WHEN** account and list IDs are supplied through `--account-id` and `--list-id`
- **THEN** the system uses those IDs even when environment variables or persisted values differ

#### Scenario: Environment fills missing flags

- **WHEN** an ID is not supplied by flag and its corresponding `CLOUDFLARE_ACCOUNT_ID` or `CLOUDFLARE_LIST_ID` variable is set
- **THEN** the system uses the environment value before consulting persisted configuration

#### Scenario: Persisted configuration is the fallback

- **WHEN** a required ID is absent from flags and the environment
- **THEN** the system uses the corresponding value from the user configuration file

#### Scenario: Required identity is missing

- **WHEN** an operation cannot resolve a non-empty required account or list ID
- **THEN** the system rejects the operation with an error identifying the missing value

### Requirement: Secure persisted configuration

The system SHALL persist only the account ID and list ID in the user configuration directory. It SHALL atomically replace the configuration file and make newly created configuration data readable only by the current user.

#### Scenario: Configuration is saved

- **WHEN** the user runs `config set` with both required ID flags
- **THEN** the system atomically saves the trimmed account and list IDs in the user configuration location
- **AND** it does not store an API token in that file

#### Scenario: Configuration input is incomplete

- **WHEN** the user runs `config set` without either required ID flag
- **THEN** the system rejects the command without replacing the persisted configuration

#### Scenario: Persisted configuration is malformed

- **WHEN** the configuration file contains unknown fields, malformed JSON, or trailing data
- **THEN** the system rejects the file rather than silently accepting ambiguous configuration

### Requirement: Token resolution and isolation

The system SHALL resolve a Cloudflare API token from `CLOUDFLARE_API_TOKEN` first and otherwise from an OS keychain entry scoped to the configured account. It SHALL NOT write the token to the application configuration file or include it in normal output.

#### Scenario: Environment token is present

- **WHEN** `CLOUDFLARE_API_TOKEN` contains a non-empty value
- **THEN** the system uses its trimmed value without reading the keychain

#### Scenario: Environment token is absent

- **WHEN** `CLOUDFLARE_API_TOKEN` is absent or empty
- **THEN** the system reads the account-specific token from the OS keychain

#### Scenario: Token cannot be found

- **WHEN** neither the environment nor the account-specific keychain contains a token
- **THEN** the operation fails with guidance to set the environment variable or authenticate

#### Scenario: Accounts use separate credentials

- **WHEN** tokens are stored for two Cloudflare accounts
- **THEN** each account resolves only its own keychain entry

### Requirement: Authenticated login and logout

The system SHALL verify a candidate token by listing the configured redirect list before storing it in the OS keychain. It SHALL support hidden interactive input and explicit standard-input token ingestion, and SHALL remove only the configured account's keychain entry on logout.

#### Scenario: Interactive login succeeds

- **WHEN** a user enters a non-empty token in an interactive terminal and it can read the configured list
- **THEN** the system stores the token in the configured account's keychain entry
- **AND** it does not echo the token

#### Scenario: Login from standard input succeeds

- **WHEN** a token is supplied to `auth login --token-stdin` and verification succeeds
- **THEN** the system stores the trimmed token without printing it

#### Scenario: Token verification fails

- **WHEN** the candidate token cannot list the configured redirect list
- **THEN** the system reports the verification failure and does not store the token

#### Scenario: Logout has only account configuration

- **WHEN** the user logs out with a resolvable account ID but no list ID
- **THEN** the system removes that account's stored token without requiring list configuration

#### Scenario: Keychain is unavailable

- **WHEN** the operating system keychain cannot be accessed
- **THEN** the system returns platform-appropriate recovery guidance without exposing credential material
