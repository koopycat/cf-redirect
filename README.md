# cf-redirect

`cf-redirect` manages one Cloudflare Bulk Redirect List from the terminal. It includes a keyboard-driven TUI and commands for scripts and CI.

Every change is shown before it runs. CSV imports only add or update redirects, so a missing row never deletes an existing redirect.

## Quick start

You need a Cloudflare Bulk Redirect List and an API token with account-level `Account Filter Lists: Edit` permission. If you do not have a list yet, [create a Bulk Redirect List and rule in Cloudflare](https://developers.cloudflare.com/rules/url-forwarding/bulk-redirects/create-dashboard/).

### 1. Install

```sh
brew install koopycat/tap/cf-redirect
```

Homebrew supports macOS and Linux. See [other installation options](#other-installation-options) if you do not use Homebrew.

### 2. Configure the list

Copy the account and list IDs from the Cloudflare dashboard URL:

```text
https://dash.cloudflare.com/ACCOUNT_ID/example.com/rules/settings/bulk-redirects/redirect-list/LIST_ID/add-redirects
```

`ACCOUNT_ID` is the first path segment. `LIST_ID` follows `redirect-list`. The zone between them is only dashboard navigation context because Bulk Redirect Lists belong to the account, not the zone.

Save them locally:

```sh
cf-redirect config set \
  --account-id YOUR_ACCOUNT_ID \
  --list-id YOUR_LIST_ID
```

### 3. Store the API token

If you need a token, [create a custom Cloudflare API token](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/) with account-level `Account Filter Lists: Edit` permission. Then run:

```sh
cf-redirect auth login
```

The command checks that the token can read the configured list, then stores it in the OS keychain. See [Authentication](docs/authentication.md) for token setup, headless Linux, WSL, containers, and CI.

### 4. Check the connection

```sh
cf-redirect list
```

Run `cf-redirect` with no command to open the TUI:

```sh
cf-redirect
```

Preview a change without applying it:

```sh
cf-redirect add example.com/old/ https://www.example.com/new/ --dry-run
```

Remove `--dry-run` to review the same plan and confirm it interactively.

## Common commands

| Task | Command |
| --- | --- |
| Open the TUI | `cf-redirect` |
| List redirects | `cf-redirect list` |
| Search sources, targets, and comments | `cf-redirect search example.com` |
| Add a redirect | `cf-redirect add <source> <target>` |
| Change a target | `cf-redirect edit <source> <new-target>` |
| Change a source and target | `cf-redirect edit <source> <new-target> <new-source>` |
| Delete by exact source | `cf-redirect delete <source>` |
| Preview deletion of every redirect | `cf-redirect clear --dry-run` |
| Import CSV upserts | `cf-redirect import redirects.csv` |
| Export CSV | `cf-redirect export redirects.csv` |

Run `cf-redirect --help` to see every command, or `cf-redirect <command> --help` for its arguments and flags. Shell completion is available through `cf-redirect completion --help`.

Machine-readable output is available from `list` and `search` with `--format json` or `--format csv`. Other useful commands include `cf-redirect config show`, `cf-redirect auth logout`, and `cf-redirect status <operation-id>`.

## Import and export

Export the configured list to a file:

```sh
cf-redirect export redirects.csv
```

Use `-` for standard input or output:

```sh
cf-redirect export - > redirects.csv
cf-redirect import - --dry-run < redirects.csv
```

Import accepts commas or semicolons. The header is optional:

```csv
source,target
example.com/old/,https://www.example.com/new/
```

```csv
example.com/old/;https://www.example.com/new/
```

A file must use one separator throughout and have exactly two fields per row. Import matches rows by source, then adds or updates them. New redirects always enable Cloudflare's **Preserve query string** option. Existing redirects retain their current options when updated. Import reports unchanged rows as `skipped existing`, and redirects missing from the file are left alone.

CSV contains only `source` and `target`. It does not preserve comments, status codes, or redirect options, so it is not a full backup. Export replaces an existing destination only after the complete file has been written successfully.

## Scripts and CI

A mutating command without an interactive terminal must use one of these flags:

- `--dry-run` prints the plan without applying it.
- `--yes` applies the plan without asking for confirmation.

Pass credentials and list IDs through the environment:

```sh
export CLOUDFLARE_API_TOKEN="$(your-secret-manager read cloudflare-api-token)"
export CLOUDFLARE_ACCOUNT_ID="YOUR_ACCOUNT_ID"
export CLOUDFLARE_LIST_ID="YOUR_LIST_ID"

cf-redirect import redirects.csv --dry-run
cf-redirect import redirects.csv --yes
```

`CLOUDFLARE_API_TOKEN` takes precedence over the OS keychain. Flags take precedence over account and list ID environment variables, which take precedence over the saved config file. See [Authentication](docs/authentication.md) for keychain setup, WSL instructions, and loading credentials from an env file.

## URL behavior

A source may omit the scheme, as in `example.com/blog/`. Cloudflare then matches both HTTP and HTTPS. A target must be an absolute `http://` or `https://` URL. Sources and targets cannot contain fragments, user information, or control characters.

Source matching for edit, delete, and import is exact. Copy the source shown by `cf-redirect list` and use trailing slashes consistently.

## Safety behavior

- Every add, edit, import, delete, and clear operation produces a plan before it runs.
- CSV import never deletes redirects that are absent from the file. There is no sync mode.
- New redirects preserve the original request's query string by default.
- Edits preserve comments and redirect options, including options explicitly set to `false`.
- Updates delete the old item, wait for Cloudflare to finish, and then add the replacement. The client never uses Cloudflare's replace-all endpoint.
- Large mutations run in batches and show live batch/operation progress. If Cloudflare returns HTTP 429, the rejected batch waits and retries automatically; the status includes the retry countdown and attempt number.
- If an update fails after deletion, the old redirect may already be gone. The error reports a partial apply. Run `cf-redirect list`, then retry or add the redirect again.
- The TUI requires `y` before applying a plan.
- Pressing `q` or `Esc` while an operation is running stops the local wait. An operation already sent to Cloudflare may still finish. Reopen the TUI or run `cf-redirect list` to check the result.

## Other installation options

Download a prebuilt archive for Linux or macOS on `amd64` or `arm64` from [GitHub Releases](https://github.com/koopycat/cf-redirect/releases). Each release includes a `SHA256SUMS` file.

To install with Go 1.25 or newer:

```sh
go install github.com/koopycat/cf-redirect/cmd/cf-redirect@latest
```

## Documentation

- [Authentication, headless Linux, and WSL](docs/authentication.md)
- [Contributing, development, and releases](CONTRIBUTING.md)
- [License](LICENSE)
