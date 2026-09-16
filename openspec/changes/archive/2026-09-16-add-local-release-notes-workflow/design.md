## Context

See `proposal.md` for motivation. The tag-triggered GitHub Actions workflow validates the tag, tests and builds the application, creates the GitHub release with `gh release create --generate-notes`, and updates Homebrew. Recent generated bodies contain only a compare link because most changes are direct commits rather than merged pull requests.

Release-note drafting is a maintainer operation rather than application behavior. It spans local Git history, a locally authenticated Pi installation, local review tools, and the GitHub CLI. The generated text must remain reviewable and must not gain permission to execute repository instructions. The workflow must coexist with the current release action rather than move model credentials into CI.

## Goals / Non-Goals

**Goals:**
- Make one local command responsible for deterministic tag resolution, draft lifecycle, review state, and GitHub publication.
- Keep the model's role limited to converting supplied evidence into user-facing prose.
- Make publication explicit, review-gated, repeatable, and testable without external services.
- Store drafts outside the repository without exposing them to normal Git status or commits.

**Non-Goals:**
- Use Pi to create tags, GitHub releases, binaries, checksums, or Homebrew updates.
- Introduce a reusable agent skill or an interactive Pi prompt command.
- Automatically infer or bump the next semantic version.
- Maintain a repository-tracked changelog or alter application version embedding.
- Add release-note generation to GitHub Actions.

## Decisions

### Use one shell command with explicit subcommands

Add `scripts/release-notes` with `generate`, `inspect`, and `publish` subcommands. The script owns validation and state transitions; Pi is a subprocess used only by `generate`.

A dedicated skill was rejected because tag selection, file lifecycle, confirmation, and remote mutation are deterministic mechanics that should be usable without an outer agent session and covered by ordinary tests. A prompt template was also rejected because those templates target interactive `/command` expansion. The generation instructions instead live in `scripts/release-notes-prompt.md`, where the script can load a stable, reviewable prompt directly.

### Store drafts and inspection records under the OS temporary directory

Resolve a repository-specific root under `${TMPDIR:-/tmp}/cf-redirect-release-notes/` by hashing the canonical repository root, then store each draft as `<repository-id>/<tag>.md`. Store an adjacent inspection record containing a digest of the reviewed draft. This keeps all generated state outside both the working tree and `.git`, avoids collisions between separate clones, and remains available across separate `generate`, `inspect`, and `publish` invocations while the operating system retains temporary files.

`generate` creates the private directory with restrictive permissions, writes to a temporary sibling, and atomically renames it only after Pi succeeds and validation passes. It refuses to overwrite by default; `--force` replaces the draft and removes any previous inspection record. `inspect` records the digest only after the pager or editor exits successfully and the draft remains non-empty. `publish` recomputes the digest and requires an exact match.

Alternatives considered:
- Git-private storage under `.git`: durable and naturally clone-specific, but explicitly rejected because generated release content should not live inside repository metadata.
- A repository directory plus `.gitignore`: visible project state and easier accidental handling.
- A single predictable file directly under `/tmp`: vulnerable to collisions between clones and users.
- `mktemp` with a random directory on every command: subsequent inspect and publish commands cannot reliably locate the generated draft.
- Rely only on a confirmation prompt: does not establish that the exact content was inspected.

### Resolve release ranges locally and fail closed

The optional tag argument must match stable SemVer with a `v` prefix and resolve to a local tag. Without an argument, derive it from `internal/version/VERSION`. Find the highest-version stable tag that precedes the selected tag in version order and is an ancestor of it. Fail if no valid predecessor exists rather than silently selecting an unclear baseline.

Generation evidence includes the two versions, commit subjects and bodies, a changed-file summary, and the Git diff for that range. The prompt directs Pi to omit release bookkeeping, routine tests, and implementation detail while describing only behavior supported by the evidence. Shell code appends the repository compare URL after generation so links and tag identities are deterministic.

An explicit baseline option and first-release empty-tree behavior were rejected for the initial workflow because the repository already has a release history and a wrong baseline can produce misleading public notes.

### Derive repository identity instead of hardcoding it

Use `gh repo view --json nameWithOwner` for GitHub operations and compare-link construction, with the authenticated CLI and current checkout determining repository identity. This avoids embedding `koopycat/cf-redirect` in reusable mechanics while still failing clearly when the checkout or GitHub authentication cannot be resolved.

### Isolate Pi from agent capabilities

Invoke Pi in print mode without a session, tools, discovered context files, extensions, skills, or prompt templates. Supply the versioned prompt and generated evidence explicitly. Do not give Pi a GitHub token or any publication capability.

This reduces prompt-injection impact: repository content can influence prose but cannot cause command execution or file mutation. It does not make generated prose trustworthy, so human inspection remains mandatory. The full release diff may be sent to the configured model provider; documentation will state this so maintainers do not use generation on content they are unwilling to send to that provider.

A deterministic commit-bullet fallback was rejected because silently publishing lower-quality output blurs generation failures. Failure leaves the previous draft untouched and lets the maintainer retry or write the draft manually before inspection.

### Replace the release body idempotently

`publish` fetches the existing body, shows a unified diff against the local draft, and asks for confirmation. `--yes` suppresses only that prompt; it does not bypass tag, draft, inspection, release-existence, update, or verification checks. Publication uses `gh release edit --notes-file`, then fetches the body and compares exact content.

Appending was rejected because reruns can duplicate text and because the current generated compare link belongs at the end of the final body. The local draft is therefore the complete desired release body. The existing GitHub Action remains unchanged and initially creates the release; local publication replaces its placeholder notes afterward.

### Test through command fakes and isolated repositories

The shell test creates temporary Git repositories and prepends fake `pi` and `gh` executables to `PATH`. Tests cover tag/default resolution, protected overwrite, atomic Pi failure, inspection digest invalidation, cancellation, non-interactive publication, absent releases, and exact post-update verification. No test calls a model or GitHub.

Add the test to `just test`, which makes it part of `just check`. Application integration tests are unaffected because this workflow does not touch Cloudflare behavior.

## Risks / Trade-offs

- [A model can hallucinate or omit a change] -> Require inspection of the exact content and make publication a separate confirmed command.
- [Committed source or messages can contain prompt injection] -> Disable all Pi tools and discovered agent resources; treat model output as untrusted draft text.
- [The release diff can contain sensitive material] -> State that evidence is sent to the configured provider and keep credentials out of arguments and generated files; maintainers must not invoke generation for unsuitable content.
- [Large release diffs can exceed model context] -> Fail visibly if Pi cannot process the supplied range; do not silently truncate evidence in the initial implementation.
- [A remote release can change after inspection] -> Always fetch and display the current remote-to-local diff immediately before publication.
- [GitHub accepts an update but subsequent verification fails transiently] -> Report failure and preserve the local draft and inspection record so the maintainer can inspect the remote state and retry.
- [OS cleanup can remove temporary drafts between commands] -> Treat missing state as a clear regeneration error and document that drafts are ephemeral and local to one machine.

## Migration Plan

1. Add the script, prompt, and isolated tests.
2. Add the script test to the canonical `just test` recipe.
3. Document the local steps after the existing tag-triggered release completes.
4. Leave the GitHub Actions workflow unchanged, allowing immediate rollback by removing the local tooling and documentation without affecting release publication.
