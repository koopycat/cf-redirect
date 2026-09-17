## Context

See `proposal.md` for motivation. The `release-notes-workflow` spec's `## Purpose` is descriptive prose that no delta-spec operation can rewrite, so it drifted when `convert-release-notes-to-agent-skill` changed the drafting mechanism. The capability already exists and its requirements are current and validated.

## Goals / Non-Goals

**Goals:**
- Restore agreement between the spec's Purpose and the requirements it introduces.
- Keep the edit confined to the Purpose paragraph.

**Non-Goals:**
- Change any requirement, scenario, or observable behavior.
- Introduce a process that would repeat this drift, such as moving agent-facing instructions into specs.

## Decisions

### Treat the Purpose as a reviewed prose edit, not a spec delta

`## Purpose` states what the capability is and why it exists; requirements carry the behavior contract. Because the sync and archive workflows both treat an existing main-spec Purpose as authoritative and out of scope for a delta, the only honest reconciliation is a separate change that edits the paragraph directly and records why no delta applies.

Recording `skip_specs: true` in `.openspec.yaml` is deliberate rather than a shortcut: `openspec validate` rejects a change with zero deltas otherwise, and inventing a requirement to satisfy the validator would add behavior that does not exist.

Alternatives considered:
- Fold the Purpose rewrite into the original change's delta spec — rejected because the delta format has no Purpose operation for an existing capability.
- Leave the stale Purpose in place — rejected because it describes a Pi subprocess that the requirements no longer use.

## Risks / Trade-offs

- [A reviewer may read `skip_specs: true` as unreviewed spec drift] → The proposal states the capability and the exact section being edited, and the verification task diffs the file to prove only Purpose changed.

## Migration Plan

Edit the paragraph, confirm the requirement sections are unchanged, run `openspec validate --specs`, then archive. Rollback is a single-line revert.