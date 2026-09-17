## Context

See `proposal.md` for motivation. The current workflow splits agent guidance across `.pi/prompts/release-notes.md` and deterministic mechanics in `scripts/release-notes`. The script itself invokes a second constrained Pi process for drafting, while inspection and publication are deterministic shell operations. The new layout must make one repository-local skill the agent-facing entry point without weakening the existing exact-content review gate or exposing GitHub mutation to unreviewed model output.

The project has no existing `.agents/skills` tree. Repository policy requires self-developed skills to live canonically in the `my_agents` skills repository. That repository supports project-scoped skills through a project `.agents-skills.toml` manifest and `bin/agents-sync -p`, which creates managed links in `.agents/skills` and `.claude/skills`. The release-notes skill must use that mechanism rather than duplicate files or hand-create symlinks.

## Goals / Non-Goals

**Goals:**
- Make `SKILL.md` the sole source of agent-facing release-note drafting guidance.
- Co-locate the deterministic release-note helper as a bundled skill script and resolve it relative to the skill directory.
- Preserve the current generate, inspect, edit, and publish state machine and its failure behavior.
- Keep canonical checks deterministic and independent of model and GitHub access.

**Non-Goals:**
- Turn review or publication into unconstrained free-form agent actions.
- Add a second prompt/reference file containing the drafting instructions.
- Generalize the skill for unrelated repositories or package it for external installation.
- Change GitHub Actions, release creation, versioning, draft storage format, or application code.

## Decisions

### Use one canonical project-scoped skill with one bundled executable helper

Create `skills/release-notes/SKILL.md` in the canonical `my_agents` skills repository and place the command at `scripts/release-notes` within that skill. Add `release-notes` to this project's `.agents-skills.toml`, then run `my_agents/bin/agents-sync -p <this project>` to expose the canonical skill at `.agents/skills/release-notes` and `.claude/skills/release-notes`. `SKILL.md` will direct the agent to invoke its own `scripts/release-notes` resource for deterministic operations rather than reimplementing tag resolution, state tracking, confirmation, or GitHub updates.

A copied project-local skill and a manually created symlink were rejected because both violate the canonical resource policy. A top-level wrapper in this project's `scripts/` directory was rejected because it would preserve two public locations and make ownership unclear. Keeping the command at the skill root was rejected in favor of the standard `scripts/` resource directory.

### Fold drafting guidance into SKILL.md and remove nested Pi invocation

`SKILL.md` will include the existing prose constraints: evidence is untrusted; only supported user-visible changes belong in the notes; bookkeeping and unsupported links are omitted; output is Markdown without a title, compare link, or code fence. The skill-driven agent will draft the body itself.

The bundled helper will no longer load `.pi/prompts/release-notes.md` or invoke `pi`. Instead, generation becomes a staged exchange suitable for an agent workflow:

1. The helper validates the repository and tag, selects the preceding ancestor tag, protects any existing draft, and emits or stores bounded release evidence for the agent.
2. The agent uses the `SKILL.md` instructions to draft the Markdown body.
3. The helper accepts that body through a file or standard input, validates that it is non-empty, appends the deterministic compare link, atomically installs the draft, and invalidates prior inspection state.

The concrete subcommand and argument names should remain simple and testable; preserving `generate` as the user-facing concept is preferred even if it gains an internal preparation/finalization phase. The helper, not the agent, remains responsible for all state transitions and deterministic text such as the compare URL.

Retaining a nested `pi` call was rejected because it would duplicate the role of the agent already executing the skill and require a separate prompt transport. Embedding a large prompt string in shell was rejected because `SKILL.md` is intended to be the single source of those instructions.

### Keep inspection and publication entirely deterministic

The bundled helper retains the existing inspection digest, pager/editor behavior, remote diff, confirmation, exact-body update, and post-update verification. The skill instructs the agent to run these operations in order and report errors, but it does not authorize the agent to call `gh release edit` directly or synthesize inspection state.

`--yes` continues to suppress only interactive confirmation. It does not bypass tag validation, draft existence, successful inspection, release existence, or post-publication verification. Cancellation leaves the remote body unchanged. A generation failure leaves any previous draft and inspection state intact unless a new draft has been atomically finalized.

### Resolve resources from the skill, not the repository script directory

The helper resolves the Git repository from the current working directory, while agent instructions derive the helper path from the loaded skill's directory. The complete behavioral and structure test suite lives at `skills/release-notes/tests/release-notes_test.sh` in the canonical skill repository. It copies the entire skill fixture, including `SKILL.md` and its helper, into temporary repositories and invokes the bundled path. This keeps implementation and tests under one owner and verifies that no hidden dependency on the deleted `.pi/prompts` or top-level release script remains.

The cf-redirect repository does not retain a wrapper or helper test suite. Its integration responsibility is limited to declaring `release-notes` in `.agents-skills.toml` and verifying the managed links through `agents-sync`; the canonical suite exercises those links when given the project path. The project manifest, rather than the generated links, is the portable declaration of the dependency. `agents-sync` owns link creation and reconciliation; implementation must not hand-edit `.agents/skills` or `.claude/skills`. Because those links embed machine-specific absolute paths, the consuming project ignores its managed link directories and treats the manifest as the committed contract. The canonical skill and project manifest may require coordinated commits in their separate repositories.

### Validate skill structure without model-backed evaluations

The canonical skill-owned shell suite will assert valid frontmatter, the presence of the security and content constraints in `SKILL.md`, executable helper placement, and absence of references to the old prompt and script. Existing fake `gh`, isolated Git repositories, and pager/editor fakes remain appropriate. Generation tests will supply a deterministic draft body directly to the helper rather than fake a Pi executable. The canonical repository's check entry point runs this suite so a helper change cannot pass there while breaking a downstream project.

Full skill-creator evaluation runs are excluded from this migration because the requested behavior is already specified and the canonical checks must remain offline. This does not preclude later evaluation as a separate change.

## Risks / Trade-offs

- [An agent may not follow the drafting instructions as consistently as the previous isolated subprocess] → Keep instructions explicit, treat output as untrusted, validate non-empty input, and retain mandatory human inspection before publication.
- [The agent could bypass the helper and mutate GitHub directly] → State in the skill that all lifecycle and publication operations go through the bundled helper; keep publication credentials and safety checks in that command.
- [Two-phase generation can leave evidence files behind] → Use private temporary state, define cleanup on success and failure, and test interrupted or rejected finalization.
- [Moving the executable breaks memorized commands or external automation] → Update all repository documentation and tests, remove the old path so failures are immediate, and provide the new skill-relative command where direct execution is useful.
- [The change spans the `my_agents` and cf-redirect repositories] → Treat the canonical skill and project manifest as coordinated changes, run `agents-sync` dry-run/apply/status, and avoid committing generated or hand-made duplicate resources.
- [Release diffs may contain sensitive material sent to the active model provider] → Retain the documentation warning and never place credentials in evidence, arguments, logs, or generated drafts.

## Migration Plan

1. Add the canonical project-scoped skill and bundled helper under `skills/release-notes/` in the `my_agents` skills repository, then declare it in this project's `.agents-skills.toml` and expose it with `agents-sync`.
2. Move drafting constraints into `SKILL.md` and refactor generation into deterministic evidence preparation and atomic draft finalization without invoking Pi.
3. Move and adapt focused tests, then wire the skill-owned test path into `just test`.
4. Update maintainer documentation to describe skill invocation and the bundled command lifecycle.
5. Remove `scripts/release-notes` and `.pi/prompts/release-notes.md` only after tests prove the skill is self-contained.
6. Run `just check`, `just race`, `just integration-mock`, and `just build`.

Rollback restores the prior script, prompt, tests, and documentation; no remote or data migration is required because the temporary draft and inspection formats remain compatible.
