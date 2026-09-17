## 1. Configuration Persistence

- [x] 1.1 Add focused `internal/config` tests proving `Save` accepts and trims account-only configuration, still rejects a missing account, preserves atomic/private-file behavior, and loads the saved partial identity correctly; verify with `devenv shell -- go test ./internal/config`.
- [x] 1.2 Relax `config.Save` validation so account ID remains required while list ID is optional, without changing strict loading or full list-scoped resolution; verify with `devenv shell -- go test ./internal/config`.
- [x] 1.3 Extend resolution tests for a persisted account combined with a flag or environment list ID and for the missing-list failure case; verify with `devenv shell -- go test ./internal/config`.

## 2. CLI Behavior

- [x] 2.1 Add an end-to-end-style Cobra command test using an isolated user config directory that runs `config set` with only `--account-id`, verifies the saved identity has no list, then verifies a list-scoped resolution succeeds with a runtime list and fails clearly without one; verify with `devenv shell -- go test ./internal/cli`.
- [x] 2.2 Update `config set` to require only the account flag, accept and trim an optional list flag, and clarify in command help that saving replaces any prior list; verify account-only and both-ID command paths with `devenv shell -- go test ./internal/cli`.

## 3. Documentation

- [x] 3.1 Update `README.md` setup examples and `docs/authentication.md` precedence guidance to show account-only persistence plus `--list-id` or `CLOUDFLARE_LIST_ID` at runtime, while retaining the both-ID workflow; verify examples match `cf-redirect config set --help` and `cf-redirect --help`.

## 4. Verification

- [x] 4.1 Run `devenv shell -- just check` and fix all formatting, unit-test, script-test, and vet failures.
- [x] 4.2 Run `devenv shell -- just race` and verify the complete Go test suite passes under the race detector.
- [x] 4.3 Run `devenv shell -- just integration-mock` and verify list-scoped Cloudflare operations still receive explicit account and list IDs.
- [x] 4.4 Run `devenv shell -- just build`, execute the built CLI against an isolated config directory to save only an account ID, and verify `config show` or another list-scoped command succeeds only when a runtime list ID is provided.
