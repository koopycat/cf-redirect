## MODIFIED Requirements

### Requirement: Secure persisted configuration

The system SHALL require a non-empty account ID when persisting configuration and SHALL permit the list ID to be omitted. It SHALL persist only the account ID and an optional list ID in the user configuration directory. It SHALL atomically replace the configuration file and make newly created configuration data readable only by the current user.

#### Scenario: Configuration is saved

- **WHEN** the user runs `config set` with a non-empty account ID and list ID
- **THEN** the system atomically saves the trimmed account and list IDs in the user configuration location
- **AND** it does not store an API token in that file

#### Scenario: Account-only configuration is saved

- **WHEN** the user runs `config set` with a non-empty account ID and no list ID
- **THEN** the system atomically saves the trimmed account ID without a list ID
- **AND** it does not store an API token in that file

#### Scenario: Configuration input is incomplete

- **WHEN** the user runs `config set` without a non-empty account ID
- **THEN** the system rejects the command without replacing the persisted configuration

#### Scenario: Runtime list completes account-only configuration

- **WHEN** persisted configuration contains only an account ID
- **AND** a list-scoped command receives a non-empty list ID from its command flag or environment variable
- **THEN** the system combines the persisted account ID with the higher-precedence runtime list ID
- **AND** executes the command with both IDs explicit

#### Scenario: Account-only configuration lacks a runtime list

- **WHEN** persisted configuration contains only an account ID
- **AND** a list-scoped command cannot resolve a list ID from a command flag or environment variable
- **THEN** the system rejects the operation with an error identifying the missing list ID

#### Scenario: Persisted configuration is malformed

- **WHEN** the configuration file contains unknown fields, malformed JSON, or trailing data
- **THEN** the system rejects the file rather than silently accepting ambiguous configuration
