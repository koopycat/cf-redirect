## Why

Users may want to persist their Cloudflare account identity once while selecting a Bulk Redirect List per invocation. Today, `config set` rejects account-only configuration, forcing a list ID into persistent state even when the intended workflow supplies `--list-id` or `CLOUDFLARE_LIST_ID` at runtime.

## What Changes

- Allow `config set` to persist a non-empty `--account-id` without requiring `--list-id`.
- Keep `--account-id` mandatory for `config set`; a supplied list ID remains optional and is persisted when present.
- Allow persisted configuration files whose list ID is absent or empty while preserving strict JSON decoding, atomic replacement, and user-only file permissions.
- Continue requiring both account and list IDs when executing list-scoped commands; a runtime flag or environment variable can provide the missing list ID according to existing precedence rules.
- Update command help and user documentation to describe account-only persistence and runtime list selection.
- No safety invariants for redirect mutation planning, explicit deletion IDs, asynchronous operation ordering, or token storage change. Account and list IDs remain explicit on every Cloudflare list operation.
- Non-goals: supporting a list ID without an account ID, storing multiple named account/list profiles, changing credential precedence or keychain behavior, and allowing list-scoped operations without a resolved list ID.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `configuration-and-authentication`: Permit secure account-only persisted configuration while retaining account-and-list requirements for list-scoped operations and independent runtime resolution.

## Impact

- Configuration persistence and validation in `internal/config`.
- Cobra behavior, help text, and focused command tests for `config set` in `internal/cli`.
- Setup and configuration documentation in `README.md` and `docs/authentication.md`.
- Existing configuration files with both IDs remain compatible; account-only files become newly valid. No new dependencies or Cloudflare API changes are required.
