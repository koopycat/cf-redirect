## ADDED Requirements

### Requirement: Generate a release-note draft from a tagged release

The maintainer workflow SHALL generate a complete Markdown release-note draft for a local stable semantic-version tag by comparing that tag with the immediately preceding stable semantic-version tag and asking Pi to summarize the supplied Git evidence. When no tag is specified, the workflow SHALL use `v` followed by the version recorded in `internal/version/VERSION`. The generated body SHALL end with the deterministic GitHub compare link for the two tags and SHALL be stored in an OS temporary directory outside the repository so it cannot be accidentally committed.

#### Scenario: Generate notes for the embedded version
- **WHEN** the maintainer requests generation without a tag and the embedded version and its preceding stable tag both exist locally
- **THEN** the workflow generates a non-empty draft for the corresponding `v`-prefixed tag in a repository-specific location under the OS temporary directory and reports its path

#### Scenario: Generate notes for an explicit tag
- **WHEN** the maintainer requests generation for an explicit stable semantic-version tag with a preceding stable tag
- **THEN** the workflow compares exactly those tags and stores the draft under the explicit tag's identity

#### Scenario: Reject invalid release context
- **WHEN** the selected tag is not a stable semantic version, does not exist locally, or has no preceding stable semantic-version tag
- **THEN** generation fails without creating or replacing a draft

#### Scenario: Preserve an existing draft
- **WHEN** a draft already exists for the selected tag and the maintainer has not explicitly requested replacement
- **THEN** generation fails without modifying the existing draft

#### Scenario: Pi generation fails
- **WHEN** Pi is unavailable, returns an error, or produces an empty response
- **THEN** generation fails without replacing any existing draft or leaving a partial draft

### Requirement: Constrain AI-assisted generation

The workflow SHALL run Pi non-interactively with sessions, tools, discovered context files, extensions, skills, and prompt templates disabled. Pi SHALL receive only the versioned release-note instructions and release evidence assembled from the selected Git range. The workflow SHALL not require model credentials in GitHub Actions.

#### Scenario: Generate through a locally authenticated Pi installation
- **WHEN** the maintainer generates a draft
- **THEN** Pi uses the maintainer's local provider configuration and cannot execute tools or load project-local agent resources during generation

#### Scenario: Repository content contains agent instructions
- **WHEN** commit messages or changed content contain instructions addressed to an agent
- **THEN** those instructions are treated only as release evidence and cannot grant Pi tool access

### Requirement: Inspect and edit a draft before publication

The workflow SHALL provide an inspection command that displays the complete local draft through the maintainer's pager or, when requested, opens it in the maintainer's editor. Successful completion of inspection SHALL record the exact draft content that was reviewed. Generating or modifying the draft afterward SHALL require another inspection before publication.

#### Scenario: Inspect a generated draft
- **WHEN** the maintainer inspects an existing draft and the pager exits successfully
- **THEN** the workflow records that exact draft as inspected and reports whether it differs from the existing GitHub release body when that release exists

#### Scenario: Edit during inspection
- **WHEN** the maintainer requests editable inspection and the editor exits successfully with a non-empty draft
- **THEN** the workflow records the resulting exact content as inspected

#### Scenario: Cancel or fail inspection
- **WHEN** the pager or editor exits unsuccessfully
- **THEN** the workflow does not mark the draft as inspected

### Requirement: Publish only reviewed release notes

The workflow SHALL publish by replacing the body of an existing GitHub release with the complete inspected draft. Before mutation, it SHALL display the difference between the current remote body and the proposed body and require interactive confirmation unless an explicit non-interactive confirmation flag is supplied. It SHALL refuse publication when the local draft no longer matches the inspected content and SHALL verify after publication that the remote body exactly matches the draft.

#### Scenario: Publish reviewed notes interactively
- **WHEN** an inspected draft matches its inspection record, the corresponding GitHub release exists, and the maintainer confirms the displayed change
- **THEN** the workflow replaces the release body and verifies that the published body equals the local draft

#### Scenario: Cancel publication
- **WHEN** the maintainer declines confirmation
- **THEN** the workflow exits without changing the GitHub release

#### Scenario: Publish non-interactively
- **WHEN** the maintainer supplies the explicit non-interactive confirmation flag for a currently inspected draft
- **THEN** the workflow publishes without prompting while retaining all other validation and post-publication verification

#### Scenario: Draft changed after inspection
- **WHEN** the local draft content differs from the content recorded at inspection
- **THEN** publication fails without changing the GitHub release and directs the maintainer to inspect again

#### Scenario: Release does not exist
- **WHEN** no GitHub release exists for the selected tag
- **THEN** publication fails without creating a release or changing the release workflow

#### Scenario: Remote update or verification fails
- **WHEN** GitHub rejects the update or the fetched release body does not exactly match the draft afterward
- **THEN** the workflow reports failure and does not claim successful publication

### Requirement: Integrate with the maintainer release process

The repository SHALL document the local generation, inspection, optional editing, and publication sequence after the tag-triggered release workflow creates the GitHub release. The project's canonical checks SHALL exercise the deterministic script behavior without invoking a model or mutating a real GitHub release.

#### Scenario: Maintainer follows release documentation
- **WHEN** a maintainer prepares and pushes a valid release tag
- **THEN** the documentation explains how to wait for release publication, generate and inspect a local draft, and publish the reviewed body without committing the draft

#### Scenario: Run project checks
- **WHEN** the canonical local check command runs
- **THEN** release-note workflow tests use isolated fakes for Pi and GitHub operations and do not require provider credentials, GitHub authentication, or network access
