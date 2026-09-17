## Why

The release-note workflow currently exposes its agent-facing instructions as a Pi prompt while keeping the command in the repository's general `scripts/` directory. Packaging the workflow as an agent skill makes its purpose, trigger conditions, instructions, and deterministic helper tooling discoverable as one unit under `.agents/skills`, while retaining the review gate around public release mutations.

## What Changes

- Add a project-scoped release-notes agent skill to the canonical `my_agents` skill repository and expose it in this project at `.agents/skills/release-notes/` through the managed project-skill linkage, with clear trigger metadata and the release-note writing instructions embedded directly in `SKILL.md`.
- Move the existing deterministic `generate` / `inspect` / `publish` command beside the canonical skill as `scripts/release-notes`; the skill references this bundled helper and the helper no longer loads a separate prompt file.
- Move the workflow's behavioral and structure tests beside the canonical skill so implementation and tests share one owner; keep project verification focused on managed-link integration. Update maintainer documentation to use the skill-owned paths and describe both agent invocation and the deterministic review/publish lifecycle.
- Remove the superseded top-level release-note script and Pi prompt after their behavior and tests have moved into the skill.
- Preserve draft isolation, exact-content inspection records, confirmation requirements, atomic generation behavior, and post-publication verification. This change does not affect Cloudflare mutation, credential, account-scope, or redirect-list safety invariants.
- **Non-goals:** automate releases in CI; let the agent bypass inspection or confirmation; create tags or GitHub releases; change semantic-version selection; commit generated notes; maintain a changelog; alter application behavior; run skill evaluations or publish the skill as a general-purpose package.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-notes-workflow`: Make the release-note workflow available as a repository-local agent skill whose bundled deterministic command retains the existing generation, review, and publication guarantees.

## Impact

- Affected paths include the canonical `my_agents` skills repository (`skills/release-notes/`), this project's `.agents-skills.toml` and generated `.agents/skills/release-notes` link, plus `scripts/release-notes`, `.pi/prompts/release-notes.md`, `scripts/release-notes_test.sh`, `justfile`, and `CONTRIBUTING.md`.
- The documented entry point changes from a general repository script to an agent skill and its bundled helper; no Go APIs, Cloudflare APIs, release Action, or runtime dependencies change.
- Local operation continues to depend on Git, Pi, GitHub CLI authentication, a model-provider configuration for generation, and a pager or editor for review.
