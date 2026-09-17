## Why

The release-notes workflow is no longer implemented in this repository. `convert-release-notes-to-agent-skill` moved the skill, its bundled helper, and its tests into the canonical `my_agents` skills repository, leaving this repository with only a manifest declaration and maintainer documentation. The surviving `release-notes-workflow` spec still describes a "repository-local" skill, its `SKILL.md`, its bundled helper layout, and the canonical repository's checks — all of which now live elsewhere. A spec in this repository therefore defines behavior this repository does not own, and will drift again the next time the skill changes, exactly as its `## Purpose` already did.

## What Changes

- Retire the `release-notes-workflow` capability and delete `openspec/specs/release-notes-workflow/spec.md`.
- Remove all five of its requirements: draft generation, AI-assistance constraints, pre-publication inspection and editing, reviewed-only publication, and maintainer-release-process integration.
- **BREAKING** for this repository's spec surface only: no Cloudflare, credential, account-scope, or redirect-list behavior changes.
- The release-note contract now lives with the skill that implements it, in `my_agents`. Maintainer documentation in `CONTRIBUTING.md` and the project-level `.agents-skills.toml` declaration are unchanged and remain non-normative.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-notes-workflow`: Retire the capability. All requirements are removed and the spec file is deleted; nothing replaces it in this repository.

## Impact

- `openspec/specs/release-notes-workflow/spec.md` is deleted
- No source, test, documentation, or runtime change
- Release-note publication is subsequently constrained only by the skill's own instructions and tests in `my_agents`, which has no OpenSpec root