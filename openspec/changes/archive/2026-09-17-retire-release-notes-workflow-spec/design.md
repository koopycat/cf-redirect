## Context

See `proposal.md` for motivation. `release-notes-workflow` is the only capability in `openspec/specs/` that describes work performed outside this repository. Its implementation, helper, and tests now live in the canonical `my_agents` skills repository, which has no OpenSpec root, so retiring the capability here leaves the contract with the skill's own instructions and tests.

The main spec is well-formed for retirement: it has a title, a `## Purpose`, and a `## Requirements` section whose only content is the five canonical requirements and their scenarios. No fenced examples or unrelated sections would be lost.

## Goals / Non-Goals

**Goals:**
- Remove the spec that describes a capability this repository no longer owns.
- Leave no empty `## Requirements` section or stray capability directory behind.

**Non-Goals:**
- Create an OpenSpec root in `my_agents` or move the spec there.
- Change the skill, its helper, its tests, the project manifest, or maintainer documentation.
- Change any Cloudflare, credential, account-scope, redirect-list, or release-automation behavior.

## Decisions

### Retire the capability instead of trimming it to a project-level contract

A trimmed spec could keep a tool-agnostic statement that drafts are reviewed before publication. It was rejected because the enforcement, the evidence preparation, and the review record all live in the skill. A spec without an implementation in the same repository cannot be validated here and would drift, as the `## Purpose` already did during the migration.

### Delete the spec file rather than leave an empty Requirements section

The capability is retired, so the file and its directory are removed. `openspec validate --specs` treats the absence of a capability as valid, while an empty `## Requirements` section is malformed. `.openspec.yaml` declares `retire_capabilities: true` so the removal is an explicit, reviewable decision rather than an accidental deletion.

Alternatives considered:
- Keep the spec and mark it deprecated — rejected because OpenSpec has no deprecation state, so the file would keep claiming behavior this repository does not implement.
- Move the spec into `my_agents` — rejected as out of scope; that repository has no OpenSpec root, and creating one is a separate decision.

## Risks / Trade-offs

- [Release-note publication loses a repository-level safety constraint] → The skill's `SKILL.md` and its offline test suite remain the enforcing contract in `my_agents`; `CONTRIBUTING.md` still documents the reviewed-before-publish process for maintainers.
- [A future capability with the same name reappears without history] → The delta, this design, and the proposal record why it was retired, and the deletion is a reviewable diff.

## Migration Plan

Apply the REMOVED delta so the main spec is deleted, confirm `openspec validate --specs` passes with six capabilities, then archive this change. Rollback restores the spec file from git.