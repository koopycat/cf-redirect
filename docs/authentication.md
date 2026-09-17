# Authentication

`cf-redirect` needs a Cloudflare API token with account-level `Account Filter Lists: Edit` permission. `Read` permission is not enough because add, edit, import, and delete change list items.

Cloudflare cannot limit this permission to one list. The token can edit every filter list and Bulk Redirect List in the account.

## Token lookup order

`cf-redirect` checks for a token in this order:

1. `CLOUDFLARE_API_TOKEN`
2. The account-specific OS keychain entry

The config file contains only the account and list IDs. `cf-redirect` never writes an API token to it.

## Store a token in the OS keychain

Configure the account and list first, then run:

```sh
cf-redirect auth login
```

The command asks for the token without echoing it, checks that it can list the configured redirect list, and saves it in the OS keychain.

To pipe a token from another program:

```sh
printf '%s' "$TOKEN" | cf-redirect auth login --token-stdin
```

Remove the token with:

```sh
cf-redirect auth logout
```

Keychain entries include the account ID, so separate Cloudflare accounts do not share one stored token.

## Create a Cloudflare token

You can [create a custom token in Cloudflare](https://developers.cloudflare.com/fundamentals/api/get-started/create-token/) with account-level `Account Filter Lists: Edit` permission.

A repository checkout also includes a helper that opens the same form with the account, token name, and permission filled in. It requires `jq` and either `cf-redirect` on `PATH` or a binary at `bin/cf-redirect`:

```sh
scripts/create-cloudflare-token \
  --account-id YOUR_ACCOUNT_ID
```

Review the form before creating the token, then paste it into the hidden terminal prompt. The helper checks access to the configured list and stores the token in the OS keychain. It does not print the token or pass it as a process argument.

Use `--verify-list-id LIST_ID` to check a different list, or `--no-open` to print the form URL.

## Headless Linux and WSL

Linux keychain storage uses the Secret Service D-Bus API. It does not require a desktop.

On Debian, Ubuntu, or WSL, install a user-session bus, GNOME Keyring, and the optional diagnostic tool:

```sh
sudo apt update
sudo apt install dbus-user-session gnome-keyring libsecret-tools
```

Start a new login session after installation.

### Start a D-Bus session

Check whether the shell already has a D-Bus user session:

```sh
printf '%s\n' "${DBUS_SESSION_BUS_ADDRESS:-not running}"
```

If it is not running, replace the current shell with a login shell inside a D-Bus session:

```sh
exec dbus-run-session -- "$SHELL" -l
```

Keep that shell open. Its D-Bus session ends when the shell exits.

### Create or unlock the login keyring

In Bash, read the keyring password without echoing it or adding it to shell history:

```bash
read -rsp 'Keyring password: ' KEYRING_PASSWORD; printf '\n'
printf '%s' "$KEYRING_PASSWORD" | gnome-keyring-daemon --unlock
unset KEYRING_PASSWORD
```

The first run creates the login keyring. Later runs must use the same password to unlock it.

Now store the Cloudflare token:

```sh
cf-redirect auth login
```

`cf-redirect` does not need `libsecret-tools`, but `secret-tool` can confirm that the token is present without printing it:

```sh
secret-tool lookup service cf-redirect username cloudflare-api-token:YOUR_ACCOUNT_ID >/dev/null \
  && echo 'cf-redirect token found'
```

An encrypted keyring cannot unlock itself without getting its password from somewhere else. For unattended WSL sessions, servers, containers, and CI, inject `CLOUDFLARE_API_TOKEN` instead.

## Environment variables

Set the token for the current process environment:

```sh
export CLOUDFLARE_API_TOKEN="$(your-secret-manager read cloudflare-api-token)"
cf-redirect list
```

Do not put the token directly in shell history, a command argument, or the repository.

### Load an env file

`cf-redirect` does not parse `.env` files itself. Load the file into a subshell before starting it:

```bash
(
  set -a
  source ~/.config/cf-redirect/env
  set +a
  exec cf-redirect list
)
```

`set -a` exports variables assigned while the file is loaded. `set +a` turns that behavior off. The subshell keeps the variables out of the parent shell.

An env file can contain:

```dotenv
CLOUDFLARE_API_TOKEN=YOUR_TOKEN
CLOUDFLARE_ACCOUNT_ID=YOUR_ACCOUNT_ID
CLOUDFLARE_LIST_ID=YOUR_LIST_ID
```

Protect it because it contains a plaintext token:

```sh
chmod 600 ~/.config/cf-redirect/env
```

`source` executes the file as shell code. Only load a file you control.

## Precedence for account and list IDs

Account and list IDs are resolved in this order:

1. `--account-id` and `--list-id`
2. `CLOUDFLARE_ACCOUNT_ID` and `CLOUDFLARE_LIST_ID`
3. The saved config file

Persist both IDs when one list is always used:

```sh
cf-redirect config set --account-id YOUR_ACCOUNT_ID --list-id YOUR_LIST_ID
```

Or persist only the account and select the list at runtime:

```sh
cf-redirect config set --account-id YOUR_ACCOUNT_ID
cf-redirect --list-id YOUR_LIST_ID list
# Alternatively:
CLOUDFLARE_LIST_ID=YOUR_LIST_ID cf-redirect list
```

`config set` replaces the saved configuration. Omitting `--list-id` clears any previously saved list ID. List-scoped commands still require a list ID from a flag, environment variable, or saved configuration.

Inspect the resolved IDs and config path with:

```sh
cf-redirect config show
```

With account-only saved configuration, `config show` also requires a runtime list ID because it displays the fully resolved list-scoped identity.
