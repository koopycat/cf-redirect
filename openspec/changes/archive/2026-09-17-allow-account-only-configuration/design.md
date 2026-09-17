## Context

See `proposal.md` for motivation. The persisted `Config` model already represents account and list IDs independently, and resolution already fills each missing value using flag, environment, then persisted configuration. The restriction is at the persistence boundary: both the Cobra `config set` handler and `config.Save` reject an empty list ID. List-scoped commands correctly remain stricter because Cloudflare API operations require both IDs.

The configuration file is strict JSON, is atomically replaced, and is created with user-only permissions. Tokens are resolved separately from environment or an account-scoped keychain entry and must never enter this file or normal output.

## Goals / Non-Goals

**Goals:**

- Represent account-only persistence as a valid configuration state.
- Keep persistence validation in the configuration package rather than relying only on Cobra.
- Preserve independent precedence resolution so a runtime list ID completes an account-only saved configuration.
- Preserve strict decoding, atomic replacement, file permissions, and credential isolation.

**Non-Goals:**

- Introduce separate partial and complete configuration models.
- Change list-scoped API requirements, mutation execution, or failure ordering.
- Change login verification, token resolution, or keychain storage.
- Merge omitted values with the previously persisted file; `config set` continues to replace saved configuration with exactly the supplied account and optional list identity.

## Decisions

### Make the list ID optional only at the persistence boundary

`config.Save` will trim both fields, require only a non-empty account ID, and serialize the optional list ID using the existing `Config` shape. The `config set` command will mirror that contract by requiring `--account-id` and accepting an omitted `--list-id`.

This keeps one validation rule in the package that owns persistence and prevents future callers from accidentally reintroducing the stricter behavior. An alternative was a separate `SaveAccount` function, but that would duplicate atomic file handling and create two persistence paths for the same file.

### Keep full resolution strict for list-scoped commands

The existing full resolver will continue to require both IDs after independently applying flag, environment, and file precedence. Therefore an account-only file works when `--list-id` or `CLOUDFLARE_LIST_ID` supplies the list and fails with the existing missing-list diagnostic otherwise. Account-only operations continue to use account-only resolution.

An alternative was to let full resolution return a partial `Config` and defer validation to each caller. That would weaken the invariant that every Cloudflare list operation receives explicit account and list IDs and spread validation across CLI paths.

### Replace rather than merge persisted identity

Running `config set --account-id ...` will produce account-only saved state even if the previous file contained a list ID. This makes omission intentional and predictable, and supports switching to a runtime-selected list without stale persisted fallback. Merging with the old file would make it impossible to clear a previously saved list through the requested command form and would make behavior depend on prior state.

### Keep the existing JSON field shape

The account-only file may retain an empty `list_id` string under the existing JSON representation; no schema migration or custom marshaling is needed. Loading already trims empty values, and full resolution treats an empty persisted list as missing. Omitting the JSON field with `omitempty` was considered but adds a serialization-format change with no behavioral benefit.

## Risks / Trade-offs

- [A user omits `--list-id` accidentally and clears a previously persisted list] → Command help and documentation will state that `config set` replaces the saved identity and that a runtime list is then required.
- [Partial persisted configuration could leak into list operations] → Keep the full resolver's final non-empty account-and-list validation unchanged and cover account-only plus runtime-list and missing-runtime-list cases in tests.
- [Persistence validation diverges between CLI and package] → Require account ID in both layers and add focused tests at both boundaries.
- [Credential exposure] → Do not alter token inputs, output, serialization, or keychain code; configuration remains limited to identifiers.

## Migration Plan

No migration is required. Existing files containing both IDs remain valid. After release, users can replace such a file with account-only configuration through `config set`; rollback is achieved by running the prior command form with both IDs. No Cloudflare mutation behavior or partial-failure semantics are affected.
