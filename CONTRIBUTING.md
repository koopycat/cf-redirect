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

### Publishing reviewed release notes

After the release workflow creates the GitHub release, ask an agent to use the repository's `release-notes` skill. For example:

- "Generate release notes for v1.2.3."
- "Draft the release notes for the current version, then let me review them."
- "Inspect and publish the reviewed release notes for v1.2.3."

The skill is exposed through the managed `.agents/skills/release-notes` link. It prepares deterministic release evidence, drafts only user-visible changes, and uses its bundled helper for every state transition. Generation sends commit messages, the changed-file summary, and the complete diff since the preceding stable tag to the model provider configured for the active agent. Do not generate notes for changes you are unwilling to send to that provider.

The agent follows this order:

1. Prepare validated evidence with `.agents/skills/release-notes/scripts/release-notes generate [TAG] [--force]`.
2. Draft a Markdown body from that untrusted evidence and finalize it with `generate [TAG] --body-file FILE` or `generate [TAG] --stdin`.
3. Review with `inspect [TAG]`, or review and edit with `inspect [TAG] --edit`.
4. Publish the exact inspected body with `publish [TAG]`.

`TAG` is optional. When omitted, it defaults to `v` followed by `internal/version/VERSION`. The helper requires local stable tags and chooses the immediately preceding stable tag that is an ancestor of the selected release. Direct helper use is useful for resuming review or publication:

```sh
RELEASE_NOTES=.agents/skills/release-notes/scripts/release-notes
"$RELEASE_NOTES" inspect v1.2.3 --edit
"$RELEASE_NOTES" publish v1.2.3
```

Drafts and prepared evidence are never committed. They use private permissions in a repository-specific directory under `${TMPDIR:-/tmp}/cf-redirect-release-notes/`. The operating system may clean this directory. Preparation refuses to proceed when a draft exists unless `--force` is supplied; the existing draft remains untouched until a non-empty agent-authored body is successfully finalized. Successful finalization invalidates any earlier inspection.

`inspect` opens the draft with `$PAGER`; `inspect --edit` uses `$VISUAL`, then `$EDITOR`, then `vi`. A successful inspection records the exact reviewed content. Changing the draft afterward makes reinspection mandatory.

`publish` requires an existing GitHub release, displays the current-body difference, and asks before replacement. Cancellation leaves the release unchanged. Use `--yes` only for deliberate non-interactive publication; it bypasses the prompt but not validation, inspection, exact-file publication, or post-update verification. The skill never authorizes direct `gh release edit` use.

Run `.agents/skills/release-notes/scripts/release-notes --help` for the complete interface.

The skill and its helper are owned by the canonical `my_agents` repository. Their structure and behavioral tests live beside the helper at `skills/release-notes/tests/release-notes_test.sh` in that repository and run offline through its `bin/test` check. This project does not keep a copy of those tests; it only declares the skill in `.agents-skills.toml`. To verify that this project's links resolve to the canonical skill, run the canonical check against the project root:

```sh
"$MY_AGENTS"/bin/test --project "$(pwd)"
```

### Homebrew credentials

Homebrew publishing uses a GitHub App installed only on `homebrew-tap`. Add its credentials as the `HOMEBREW_APP_ID` and `HOMEBREW_APP_PRIVATE_KEY` Actions secrets. Give the app read and write access to repository contents and no other optional repository or organization permissions.

The workflow passes the numeric app ID through the action's `client-id` input, which also accepts an OAuth-style client ID. It pins each action to a reviewed commit and limits the generated installation token to contents access on `homebrew-tap`.

Protect `v*` tags so that only release maintainers can create, update, or delete them. Protect `main` in both repositories from force pushes and deletion.
