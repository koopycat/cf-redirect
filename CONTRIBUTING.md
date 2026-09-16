# Contributing

## Development environment

The project uses Go 1.25, `devenv`, and `just`.

Enter the development shell with direnv:

```sh
direnv allow
```

Or run a command directly through `devenv`:

```sh
devenv shell -- just check
```

## Checks

```sh
just check              # formatting, tests, and vet
just race               # race detector
just integration-mock   # redirect lifecycle against an in-process HTTP API
just build
```

`just integration-mock` exercises the planner, executor, operation polling, and HTTP client through Go's `httptest.Server`. It needs no credentials, external network access, container runtime, or separate mock server.

## Live integration test

The live test runs a redirect lifecycle against Cloudflare. Use a disposable account and list because it adds, edits, and deletes test redirects. It leaves unrelated redirects alone.

Give the token account-level `Account Filter Lists: Edit` permission. Cloudflare cannot scope this permission to one list. Supply the token through `CLOUDFLARE_API_TOKEN` or save it in the account-specific OS keychain with `cf-redirect auth login`, then run:

```sh
CF_REDIRECT_INTEGRATION=1 \
CLOUDFLARE_ACCOUNT_ID=... \
CLOUDFLARE_LIST_ID=... \
just integration-live
```

The live test will not run unless `CF_REDIRECT_INTEGRATION=1` is set.

## Publishing a release

1. Update `internal/version/VERSION`.
2. Commit the change on `main`.
3. Push `main`.
4. Create and push the matching stable semantic version tag, such as `v1.2.3`.

The release workflow checks that the tag matches the embedded version and points to a commit on `main`. It runs the test suite, builds Linux and macOS archives for `amd64` and `arm64`, creates checksums, publishes the GitHub release, and updates `koopycat/homebrew-tap`.

### Homebrew credentials

Homebrew publishing uses a GitHub App installed only on `homebrew-tap`. Add its credentials as the `HOMEBREW_APP_ID` and `HOMEBREW_APP_PRIVATE_KEY` Actions secrets. Give the app read and write access to repository contents and no other optional repository or organization permissions.

The workflow passes the numeric app ID through the action's `client-id` input, which also accepts an OAuth-style client ID. It pins each action to a reviewed commit and limits the generated installation token to contents access on `homebrew-tap`.

Protect `v*` tags so that only release maintainers can create, update, or delete them. Protect `main` in both repositories from force pushes and deletion.
