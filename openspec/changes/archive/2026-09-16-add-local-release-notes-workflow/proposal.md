## Why

GitHub's generated release notes contain only a compare link for most cf-redirect releases because development usually lands as direct commits rather than merged pull requests. Maintainers need a simple local workflow that uses Pi to draft useful notes, leaves room for human review, and publishes the reviewed text without committing generated content or adding model credentials to CI.

## What Changes

- Add a local `generate` / `inspect` / `publish` release-notes workflow for stable semantic-version tags.
- Generate a complete Markdown release body from deterministic Git context through a constrained, non-interactive Pi invocation.
- Keep drafts in an OS temporary directory outside the repository so they remain available between workflow commands but cannot be accidentally committed.
- Require an explicit inspection step before publishing, show the proposed remote change, and verify the published release body.
- Add automated script tests, task-runner integration, and maintainer documentation.
- Keep GitHub Actions responsible for building and initially publishing releases; local publication replaces the generated placeholder body afterward.
- Do not add a Pi skill: deterministic release mechanics belong in a testable script, while Pi is limited to summarizing supplied evidence.
- Do not create or maintain a committed `CHANGELOG.md`.
- Do not change Cloudflare redirect mutation, authentication, token storage, planning, or execution behavior; all existing product safety invariants remain unaffected.

## Capabilities

### New Capabilities
- `release-notes-workflow`: Local generation, review, and publication of Pi-authored GitHub release notes.

### Modified Capabilities

None.

## Impact

- Adds a maintainer-facing shell script, a versioned generation prompt, and focused script tests under `scripts/`.
- Extends the `justfile` checks and `CONTRIBUTING.md` release instructions.
- Depends on local `git`, `pi`, and `gh` commands, a writable `${TMPDIR:-/tmp}`, and the maintainer's existing Pi provider and GitHub CLI authentication.
- Does not require new application dependencies, GitHub Actions secrets, public CLI commands, or changes to release artifacts and Homebrew publication.
