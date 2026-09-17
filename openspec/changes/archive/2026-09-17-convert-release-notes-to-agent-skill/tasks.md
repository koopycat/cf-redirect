## 1. Canonical Skill Setup

- [x] 1.1 In the canonical `my_agents` skills repository, add `skills/release-notes/SKILL.md` with matching `name` frontmatter, trigger guidance, the complete release-note drafting rules formerly held in `.pi/prompts/release-notes.md`, and instructions that reference the skill-relative helper; verify the skill contains no dependency on a separate prompt file.
- [x] 1.2 Add `skills/release-notes/scripts/release-notes` as an executable bundled resource and verify its `--help` output describes the staged generation, inspection, and publication interface from both the canonical path and a managed project link.
- [x] 1.3 Add `release-notes` to the cf-redirect `.agents-skills.toml`, run `my_agents/bin/agents-sync -p <this project>` through dry-run and apply modes, and verify `agents-sync status` reports valid managed `.agents/skills/release-notes` and `.claude/skills/release-notes` links with no duplicated skill files.

## 2. Agent-Driven Draft Generation

- [x] 2.1 Refactor the bundled helper's generation flow to prepare validated release evidence and accept an agent-authored Markdown body without invoking Pi or reading a prompt file; verify focused tests cover default and explicit tags, ancestry-aware predecessor selection, stable-tag validation, evidence content, and successful draft finalization with the deterministic compare link.
- [x] 2.2 Preserve private repository-specific temporary state, protected overwrite behavior, atomic draft installation, and inspection invalidation; verify tests cover separate repositories, `--force`, empty or failed finalization, stale intermediate state, restrictive permissions, and preservation of an existing draft after every generation failure.
- [x] 2.3 Under `skills/release-notes/tests/` in the canonical skill repository, add skill-structure tests that parse the `SKILL.md` frontmatter, confirm the helper is executable and referenced by a skill-relative path, confirm untrusted-evidence and user-visible-change constraints are embedded in `SKILL.md`, and confirm neither the skill nor helper references `.pi/prompts/release-notes.md`, a top-level release-note script, or a nested Pi invocation.

## 3. Review and Publication Safety

- [x] 3.1 Move the existing inspection behavior into the bundled helper unchanged and verify tests cover pager review, editor changes, cancelled or failed inspection, empty edits, digest recording, and mandatory reinspection after draft changes.
- [x] 3.2 Move the existing publication behavior into the bundled helper unchanged and verify tests cover current-body diffs, interactive cancellation, explicit `--yes` publication, missing releases, changed drafts, rejected updates, exact notes-file use, and post-update body mismatch without contacting GitHub.
- [x] 3.3 Verify the skill directs the agent to use the bundled helper for every state transition and never to invoke `gh release edit` directly, while preserving the warning that release evidence is sent to the configured model provider.

## 4. Project Migration and Documentation

- [x] 4.1 Move the complete release-note behavioral and structure suite to `skills/release-notes/tests/release-notes_test.sh` in the canonical skill repository, remove the project-owned `scripts/release-notes_test.sh` and its `justfile` invocation, and add the suite to the canonical repository's check entry point; verify it runs without model credentials, GitHub authentication, or network access.
- [x] 4.2 Update `CONTRIBUTING.md` to document agent-trigger phrases, skill-driven generation, direct skill-relative helper commands where useful, review/edit/publication ordering, temporary draft storage, provider exposure, cancellation, and non-interactive safeguards; verify every documented path and option matches `--help`.
- [x] 4.3 Remove `scripts/release-notes` and `.pi/prompts/release-notes.md` after migration, then verify repository search finds no stale references and the skill remains usable only through its canonical managed links.

## 5. Verification

- [x] 5.1 Run the canonical skill repository's relevant checks, including the skill-owned release-note suite, plus `agents-sync --dry-run`, project sync, and `agents-sync status`; verify the skill source is canonical, project-scoped, and linked without drift.
- [x] 5.2 Run `devenv shell -- just check` and verify formatting, Go tests, migrated release-note tests, token-script tests, and vet all pass.
- [x] 5.3 Run `devenv shell -- just race`, `devenv shell -- just integration-mock`, and `devenv shell -- just build`; verify no application regression and a successful `bin/cf-redirect` build.
