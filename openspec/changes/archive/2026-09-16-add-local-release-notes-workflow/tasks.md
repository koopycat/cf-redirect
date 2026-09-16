## 1. Script Foundation and Generation

- [x] 1.1 Add the executable `scripts/release-notes` command with `generate`, `inspect`, and `publish` argument parsing, stable-tag/default-version validation, repository-specific paths under `${TMPDIR:-/tmp}`, restrictive temporary-directory permissions, dependency checks, and actionable usage errors; verify focused shell tests cover valid and invalid command forms plus isolation between repositories.
- [x] 1.2 Add `scripts/release-notes-prompt.md` and implement deterministic previous-tag selection and release evidence assembly; verify tests cover explicit and embedded-version tags, ancestry-aware baselines, missing tags, and missing predecessors.
- [x] 1.3 Implement atomic, overwrite-protected generation through a non-interactive Pi process with sessions, tools, context files, extensions, skills, and prompt templates disabled, then append the deterministic compare link; verify fake-Pi tests cover successful output, `--force`, empty output, process failure, exact flags, prompt/evidence input, and preservation of existing drafts.

## 2. Inspection and Publication

- [x] 2.1 Implement pager-based inspection and `--edit` support with a digest record created only after successful review of a non-empty draft; verify tests cover successful viewing/editing, cancellation, empty edited content, and inspection invalidation after regeneration or modification.
- [x] 2.2 Implement GitHub repository and release lookup, current-body comparison, interactive confirmation, and cancellation without remote mutation; verify fake-`gh` tests cover an existing release, missing release, displayed differences, accepted confirmation, and declined confirmation.
- [x] 2.3 Implement complete-body publication via `gh release edit --notes-file`, `--yes` prompt bypass, exact post-update verification, and failure reporting; verify tests cover reviewed-draft enforcement, successful interactive and non-interactive publication, changed-after-inspection rejection, update failure, and verification mismatch.

## 3. Integration and Documentation

- [x] 3.1 Add the isolated release-note script test suite to `just test`; verify the suite runs without network access, Pi provider credentials, or GitHub authentication.
- [x] 3.2 Update `CONTRIBUTING.md` to document prerequisites, provider data exposure, the post-release `generate` / `inspect --edit` / `publish` sequence, ephemeral OS-temporary draft storage, overwrite behavior, cleanup implications, and recovery from generation or publication failures; verify documented commands match script help.
- [x] 3.3 Run `devenv shell -- just check`, `devenv shell -- just race`, `devenv shell -- just integration-mock`, and `devenv shell -- just build` to verify the maintainer workflow and existing application behavior remain healthy.
