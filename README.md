# cf-redirect

`cf-redirect` is a Go CLI and terminal UI for managing one Cloudflare Bulk Redirect List. It can list and search redirects, preview additions and edits, import CSV upserts, and delete redirects by item ID.

## Safety model

- Every mutation is validated and shown before it runs. Scripts must pass `--yes` or `--dry-run`.
- CSV import only adds or updates redirects. It never deletes redirects that are missing from the file, and there is no `--sync` mode.
- An update deletes the old item, waits for Cloudflare to finish, and then posts the replacement.
- The client never calls Cloudflare's replace-all `PUT /items` endpoint.
- Edits preserve comments and redirect options, including boolean options explicitly set to `false`.
- The TUI requires `y` before it applies a plan. Pressing `q` or `Esc` during an apply stops the local wait, but an operation already submitted to Cloudflare may continue. Reopen the TUI or run `cf-redirect list` to check the result.

## Install

Homebrew works on macOS and Linux:

```sh
brew install koopycat/tap/cf-redirect
```

Prebuilt archives for Linux and macOS on `amd64` and `arm64` are available from [GitHub Releases](https://github.com/koopycat/cf-redirect/releases). Each release includes a `SHA256SUMS` file.

Building from source requires Go 1.25 or newer:

```sh
go install github.com/koopycat/cf-redirect/cmd/cf-redirect@latest
# or, from this checkout
just build
```

### Publishing a release

Update `internal/version/VERSION`, commit it, and push the matching stable semantic version tag, such as `v1.2.3`. The tagged commit must be on `main`. The release workflow checks the tag against the embedded version, runs the test suite, builds all supported archives, creates checksums, publishes the GitHub release, and updates `koopycat/homebrew-tap`.

Homebrew publishing uses a GitHub App installed only on `homebrew-tap`. Add its credentials as the `HOMEBREW_APP_ID` and `HOMEBREW_APP_PRIVATE_KEY` Actions secrets. Give the app read and write access to repository contents and no other optional repository or organization permissions. The workflow passes the numeric app ID through the action's `client-id` input, which also accepts an OAuth-style client ID.

Protect `v*` tags so that only release maintainers can create, update, or delete them. Protect `main` in both repositories from force pushes and deletion. The workflow pins each action to a reviewed commit and limits the generated installation token to contents access on `homebrew-tap`.

## Configuration

A Bulk Redirects dashboard URL looks like this:

```text
https://dash.cloudflare.com/ACCOUNT_ID/example.com/rules/settings/bulk-redirects/redirect-list/LIST_ID/add-redirects
```

`ACCOUNT_ID` is the first path segment. `LIST_ID` follows `redirect-list`. The zone between them is only dashboard navigation context because Bulk Redirect Lists belong to the account, not the zone.

Save the IDs and inspect the resolved configuration with:

```sh
cf-redirect config set \
  --account-id YOUR_ACCOUNT_ID \
  --list-id YOUR_LIST_ID
cf-redirect config show
```

The config file has mode `0600` and is replaced atomically. IDs are resolved in this order:

1. `--account-id` and `--list-id`
2. `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_LIST_ID`
3. The saved config file

API tokens are resolved from `CLOUDFLARE_API_TOKEN` first, then from the OS keychain. Keychain entries are scoped to the account ID. Login checks that the token can list the configured redirect list before saving it:

```sh
cf-redirect auth login

# Short aliases
cf-redirect login
cf-redirect logout
```

To read a token from standard input instead of a command argument:

```sh
printf '%s' "$TOKEN" | cf-redirect auth login --token-stdin
```

### Headless Linux and WSL

On Linux, keychain storage uses the Secret Service D-Bus API. It does not require a desktop. On Debian, Ubuntu, or WSL, install a user-session bus, GNOME Keyring, and the optional diagnostic tool:

```sh
sudo apt update
sudo apt install dbus-user-session gnome-keyring libsecret-tools
```

Start a new login session after installation. If the environment does not start a D-Bus user session, open a login shell inside one:

```sh
test -n "${DBUS_SESSION_BUS_ADDRESS:-}" || exec dbus-run-session -- "$SHELL" -l
```

A headless session may also need to create or unlock the login keyring. In Bash, this reads the keyring password without echoing it or adding it to shell history:

```bash
read -rsp 'Keyring password: ' KEYRING_PASSWORD; printf '\n'
printf '%s' "$KEYRING_PASSWORD" | gnome-keyring-daemon --unlock
unset KEYRING_PASSWORD
```

Now run `cf-redirect auth login`. `cf-redirect` does not need `libsecret-tools`, but `secret-tool` can confirm that the token is present without printing it:

```sh
secret-tool lookup service cf-redirect username cloudflare-api-token:YOUR_ACCOUNT_ID >/dev/null \
  && echo 'cf-redirect token found'
```

For CI, containers, and other short-lived sessions, an environment variable is usually simpler than a keyring daemon. Load it from the platform's secret manager rather than putting it in a repository or command argument:

```sh
export CLOUDFLARE_API_TOKEN="$(your-secret-manager read cloudflare-api-token)"
cf-redirect list
```

`config.json` contains only the account and list IDs. `cf-redirect` never writes an API token there. If you cannot inject an environment variable safely, configure a Secret Service provider instead of saving the token as plaintext.

The Cloudflare token needs account-level `Account Filter Lists: Edit` permission. `Read` permission is not enough because add, edit, import, and delete change list items.

The repository includes a helper that opens Cloudflare's token form with the account, token name, and required permission filled in:

```sh
scripts/create-cloudflare-token \
  --account-id YOUR_ACCOUNT_ID
```

Review the form before creating the token, then paste it into the hidden terminal prompt. The helper checks access to the configured list and stores the token in the OS keychain. It does not print the token or pass it as a process argument. Use `--verify-list-id LIST_ID` to check a different list, or `--no-open` to print the form URL. Cloudflare cannot limit this permission to one list, so the token can edit every filter list and Bulk Redirect List in the account.

## Usage

Running `cf-redirect` in an interactive terminal opens the TUI. The same operations are available as commands:

```sh
cf-redirect list
cf-redirect list --format json
cf-redirect export redirects.csv
cf-redirect export - > redirects.csv
cf-redirect search example.com
cf-redirect add example.com/blog/ https://www.example.com/new-blog/
cf-redirect edit example.com/blog/ --target https://www.example.com/new-blog/
cf-redirect import redirects.csv
cf-redirect delete ITEM_ID
cf-redirect clear
cf-redirect tui
```

`delete` accepts only a Cloudflare item ID. Use `list --format json` or the TUI to find one. `clear` plans explicit deletions for all current item IDs. CSV import has no sync mode.

A mutating command run without a terminal must use `--yes` or `--dry-run`. Interactive commands show the plan and ask for confirmation. `--dry-run` prints the same plan without applying it.

### Sources and targets

Sources and targets without a path are normalized to `/`:

```text
https://example.com       -> https://example.com/
https://www.example.com   -> https://www.example.com/
```

A source may omit the scheme, as in `example.com/blog/`, which matches both HTTP and HTTPS. A target must be an absolute `http://` or `https://` URL. Sources and targets cannot contain fragments, user information, or control characters.

### CSV export and import

`cf-redirect export redirects.csv` writes every redirect in the configured list to a comma-separated file with a `source,target` header. If the destination already exists, the command replaces it only after the complete export has been written. Use `cf-redirect export -` to write CSV to standard output.

Import accepts commas or semicolons. The `source,target` or `source;target` header is optional:

```csv
source,target
example.com/old/,https://www.example.com/new/
```

```csv
example.com/old/;https://www.example.com/new/
```

A file must use one separator throughout and contain exactly two fields per row. Import matches rows by source and adds or updates them as needed. Rows whose source and target already match are reported as `skipped existing`. Redirects that are absent from the file are left alone and are not counted as skipped.

## Development

The development shell provides Go 1.25 and `just`:

```sh
direnv allow
# or, without direnv
devenv shell -- just check
```

Available checks:

```sh
just check              # formatting, tests, and vet
just race               # race detector
just integration-mock   # redirect lifecycle against an in-process HTTP API
just build
```

### Integration tests

`just integration-mock` exercises the planner, executor, operation polling, and HTTP client through Go's `httptest.Server`. It needs no credentials, external network access, container runtime, or separate mock server.

The live test runs the same redirect lifecycle against Cloudflare. Use a disposable account and list because it adds, edits, imports, and deletes test redirects. The test leaves unrelated redirects alone.

Give the token read and edit permission for the configured list. Supply it through `CLOUDFLARE_API_TOKEN` or save it in the account-specific OS keychain with `cf-redirect auth login`, then run:

```sh
CF_REDIRECT_INTEGRATION=1 \
CLOUDFLARE_ACCOUNT_ID=... \
CLOUDFLARE_LIST_ID=... \
just integration-live
```

The live test will not run unless `CF_REDIRECT_INTEGRATION=1` is set.
