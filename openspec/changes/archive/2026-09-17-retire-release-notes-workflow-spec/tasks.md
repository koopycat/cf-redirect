## 1. Retire the Capability

- [x] 1.1 Confirm the main spec is retirement-eligible: it has a title, a `## Purpose`, and a `## Requirements` section containing only the five requirements named in the delta, with no other sections or fenced examples.
- [x] 1.2 Apply the REMOVED delta so `openspec/specs/release-notes-workflow/spec.md` and its now-empty directory are deleted, leaving no empty `## Requirements` section.
- [x] 1.3 Verify no main spec, document, test, or configuration in this repository still references the `release-notes-workflow` capability and that `openspec validate --all` passes.

## 2. Verification

- [x] 2.1 Run `openspec validate --specs` and confirm six remaining capabilities validate with no failure.
- [x] 2.2 Run `devenv shell -- just check` to confirm the deletion changed no application or test behavior.