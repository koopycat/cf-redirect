## Why

`convert-release-notes-to-agent-skill` moved release-note drafting out of a constrained Pi subprocess and into the agent-facing skill, but the `release-notes-workflow` spec's `## Purpose` still describes "a constrained local Pi process". The delta-spec format cannot carry a Purpose change, so the drift survived that sync. The Purpose now contradicts the requirements it introduces.

## What Changes

- Replace the stale `## Purpose` paragraph in `openspec/specs/release-notes-workflow/spec.md` so it describes the repository-local agent skill preparing release evidence and the exact-content review gate before publication.
- No requirement, scenario, or application behavior changes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. This is descriptive prose only; no requirement text or scenario changes, so the change intentionally declares `skip_specs: true`.

## Impact

- `openspec/specs/release-notes-workflow/spec.md` (`## Purpose` only)
- No source, test, documentation, or runtime impact.