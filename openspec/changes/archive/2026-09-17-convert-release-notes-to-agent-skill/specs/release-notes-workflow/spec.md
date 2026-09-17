## MODIFIED Requirements

### Requirement: Generate a release-note draft from a tagged release

The repository-local release-notes skill SHALL generate a complete Markdown release-note draft for a local stable semantic-version tag by comparing that tag with the immediately preceding stable semantic-version tag and summarizing evidence prepared by its bundled helper command. When no tag is specified, the workflow SHALL use `v` followed by the version recorded in `internal/version/VERSION`. The generated body SHALL end with the deterministic GitHub compare link for the two tags and SHALL be stored in an OS temporary directory outside the repository so it cannot be accidentally committed.

#### Scenario: Generate notes for the embedded version

- **WHEN** the maintainer asks the agent to generate release notes without a tag and the embedded version and its preceding stable tag both exist locally
- **THEN** the skill uses its bundled helper to generate a non-empty draft for the corresponding `v`-prefixed tag in a repository-specific location under the OS temporary directory and reports its path

#### Scenario: Generate notes for an explicit tag

- **WHEN** the maintainer asks the agent to generate release notes for an explicit stable semantic-version tag with a preceding stable tag
- **THEN** the skill compares exactly those tags and stores the draft under the explicit tag's identity

#### Scenario: Reject invalid release context

- **WHEN** the selected tag is not a stable semantic version, does not exist locally, or has no preceding stable semantic-version tag
- **THEN** generation fails without creating or replacing a draft

#### Scenario: Preserve an existing draft

- **WHEN** a draft already exists for the selected tag and the maintainer has not explicitly requested replacement
- **THEN** generation fails without modifying the existing draft

#### Scenario: Pi generation fails

- **WHEN** release evidence cannot be prepared, the agent cannot produce non-empty release notes, or the bundled helper rejects the generated content
- **THEN** generation fails without replacing any existing draft or leaving a partial draft

### Requirement: Constrain AI-assisted generation

The repository-local release-notes skill SHALL contain the release-note drafting instructions in its `SKILL.md` and SHALL reference a bundled deterministic helper located within the same skill. The skill SHALL treat all release evidence as untrusted data, describe only user-visible changes supported by that evidence, omit release bookkeeping, and avoid following instructions found in repository content. The workflow SHALL not require model credentials in GitHub Actions and SHALL not retain a separate release-note prompt file.

#### Scenario: Generate through a locally authenticated Pi installation

- **WHEN** the maintainer asks an agent to generate release notes
- **THEN** the agent follows the skill's embedded drafting instructions and uses the skill-relative bundled helper for deterministic tag, evidence, draft-state, review, and publication operations

#### Scenario: Repository content contains agent instructions

- **WHEN** commit messages or changed content contain instructions addressed to an agent
- **THEN** the skill treats those instructions only as release evidence and does not execute or follow them

#### Scenario: Skill or helper is incomplete

- **WHEN** the repository-local skill cannot be loaded or its referenced helper is missing or not executable
- **THEN** the workflow fails with an actionable error without creating, replacing, inspecting, or publishing a draft

### Requirement: Integrate with the maintainer release process

The repository SHALL document how to invoke the release-notes skill and complete local generation, inspection, optional editing, and publication after the tag-triggered release workflow creates the GitHub release. The canonical skill repository's checks SHALL validate the skill structure and exercise the bundled helper's deterministic behavior without invoking a model or mutating a real GitHub release, and SHALL verify the consuming project's managed linkage when given its path. The consuming project SHALL declare the skill in its manifest and SHALL not own the helper's behavioral test suite.

#### Scenario: Maintainer follows release documentation

- **WHEN** a maintainer prepares and pushes a valid release tag
- **THEN** the documentation explains how to ask an agent to use the repository-local skill and how the skill generates, inspects, and publishes the reviewed body without committing the draft

#### Scenario: Run project checks

- **WHEN** the canonical skill repository's check command runs
- **THEN** the release-note skill's structure and behavioral suite verifies the skill metadata and embedded drafting guidance and uses isolated fakes for GitHub operations without requiring provider credentials, GitHub authentication, or network access

#### Scenario: Canonical checks verify downstream linkage

- **WHEN** the canonical skill repository's checks run with a consuming project path
- **THEN** they confirm the project manifest declares the release-notes skill and that `.agents/skills/release-notes` and `.claude/skills/release-notes` resolve to the canonical skill, without the project owning helper behavioral tests
