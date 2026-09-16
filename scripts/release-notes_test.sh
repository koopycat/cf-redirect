#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
suite_tmp=$(mktemp -d)
cleanup() { rm -rf "$suite_tmp"; }
trap cleanup EXIT HUP INT TERM

tests=0

fail_test() {
  printf 'not ok - %s\n' "$*" >&2
  exit 1
}

assert_contains() {
  local file=$1 expected=$2
  grep -Fq -- "$expected" "$file" || {
    printf 'Expected %s to contain:\n%s\nActual:\n' "$file" "$expected" >&2
    cat "$file" >&2
    exit 1
  }
}

assert_not_contains() {
  local file=$1 unexpected=$2
  if grep -Fq -- "$unexpected" "$file"; then
    printf 'Expected %s not to contain:\n%s\nActual:\n' "$file" "$unexpected" >&2
    cat "$file" >&2
    exit 1
  fi
}

assert_eq() {
  local actual=$1 expected=$2 message=${3:-values differ}
  [[ "$actual" == "$expected" ]] || fail_test "$message: expected '$expected', got '$actual'"
}

assert_fails() {
  if "$@"; then
    fail_test "command unexpectedly succeeded: $*"
  fi
}

pass() {
  tests=$((tests + 1))
  printf 'ok %d - %s\n' "$tests" "$1"
}

make_repo() {
  local repo=$1 version=${2:-0.1.2}
  mkdir -p "$repo/internal/version" "$repo/scripts"
  cp "$root/scripts/release-notes" "$repo/scripts/release-notes"
  cp "$root/scripts/release-notes-prompt.md" "$repo/scripts/release-notes-prompt.md"
  chmod +x "$repo/scripts/release-notes"
  printf '%s\n' "$version" > "$repo/internal/version/VERSION"
  git -C "$repo" init -q
  git -C "$repo" config user.email release-notes-test@example.test
  git -C "$repo" config user.name 'Release Notes Test'
  printf 'first\n' > "$repo/feature.txt"
  git -C "$repo" add .
  git -C "$repo" commit -qm 'Initial release'
  git -C "$repo" tag v0.1.0
  printf 'second\n' >> "$repo/feature.txt"
  git -C "$repo" commit -qam 'Add visible feature'
  git -C "$repo" tag v0.1.1
  printf 'third\n' >> "$repo/feature.txt"
  git -C "$repo" commit -qam 'Fix visible behavior'
  git -C "$repo" tag v0.1.2
}

make_fakes() {
  local bin=$1
  mkdir -p "$bin"
  cat > "$bin/pi" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$@" > "$FAKE_PI_ARGS"
cat > "$FAKE_PI_STDIN"
case "${FAKE_PI_MODE:-success}" in
  success) printf '%s\n' "${FAKE_PI_OUTPUT:-## Changes

- Improved redirect handling.}" ;;
  empty) : ;;
  fail) exit 7 ;;
  *) exit 8 ;;
esac
SH
  cat > "$bin/gh" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$FAKE_GH_LOG"
if [[ "$1 $2" == "repo view" ]]; then
  printf '%s\n' "${FAKE_GH_REPO:-example/cf-redirect}"
  exit 0
fi
if [[ "$1 $2" == "release view" ]]; then
  [[ "${FAKE_GH_RELEASE_EXISTS:-true}" == true ]] || exit 1
  cat "$FAKE_GH_BODY"
  exit 0
fi
if [[ "$1 $2" == "release edit" ]]; then
  [[ "${FAKE_GH_EDIT_FAIL:-false}" != true ]] || exit 9
  notes_file=""
  while (($#)); do
    if [[ "$1" == --notes-file ]]; then
      notes_file=$2
      break
    fi
    shift
  done
  [[ -n "$notes_file" ]]
  if [[ "${FAKE_GH_VERIFY_MISMATCH:-false}" == true ]]; then
    printf 'different remote content\n' > "$FAKE_GH_BODY"
  else
    cp "$notes_file" "$FAKE_GH_BODY"
  fi
  exit 0
fi
exit 2
SH
  cat > "$bin/viewer" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$1" > "$FAKE_VIEWED_PATH"
[[ "${FAKE_VIEW_FAIL:-false}" != true ]]
SH
  cat > "$bin/editor" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
[[ "${FAKE_EDIT_FAIL:-false}" != true ]] || exit 6
case "${FAKE_EDIT_MODE:-append}" in
  append) printf '\n- Human edit.\n' >> "$1" ;;
  empty) : > "$1" ;;
  unchanged) : ;;
esac
SH
  chmod +x "$bin/pi" "$bin/gh" "$bin/viewer" "$bin/editor"
}

repo="$suite_tmp/repo"
other_repo="$suite_tmp/other-repo"
fake_bin="$suite_tmp/bin"
state_tmp="$suite_tmp/state"
mkdir -p "$state_tmp"
make_repo "$repo"
make_repo "$other_repo"
make_fakes "$fake_bin"

export PATH="$fake_bin:$PATH"
export TMPDIR="$state_tmp"
export FAKE_PI_ARGS="$suite_tmp/pi-args"
export FAKE_PI_STDIN="$suite_tmp/pi-stdin"
export FAKE_GH_LOG="$suite_tmp/gh-log"
export FAKE_GH_BODY="$suite_tmp/gh-body"
export FAKE_VIEWED_PATH="$suite_tmp/viewed-path"
printf 'Generated placeholder.\n' > "$FAKE_GH_BODY"
: > "$FAKE_GH_LOG"

run_script() {
  local working_repo=$1
  shift
  (cd "$working_repo" && ./scripts/release-notes "$@")
}

draft_path_from_output() {
  awk -F': ' '/^Generated draft: / {print $2}' "$1"
}

# Help and invalid command handling.
run_script "$repo" --help > "$suite_tmp/help"
assert_contains "$suite_tmp/help" 'release-notes generate [TAG] [--force]'
assert_fails run_script "$repo" unknown > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'unknown command: unknown'
assert_fails run_script "$repo" inspect --force > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" '--force is valid only with generate'
pass 'parses commands and rejects invalid forms'

# Default generation, exact constraints, evidence, compare URL, and private storage.
run_script "$repo" generate > "$suite_tmp/generate-default"
draft=$(draft_path_from_output "$suite_tmp/generate-default")
[[ -f "$draft" ]] || fail_test 'default generation did not create a draft'
case "$draft" in "$state_tmp"/cf-redirect-release-notes/*/v0.1.2.md) ;; *) fail_test "unexpected draft path: $draft" ;; esac
[[ "$draft" != "$repo"/* ]] || fail_test 'draft was stored inside the repository'
assert_eq "$(python3 -c 'import os,sys; print(oct(os.stat(sys.argv[1]).st_mode & 0o777)[2:])' "$(dirname "$draft")")" 700 'repository temporary directory permissions'
for flag in --print --no-session --no-tools --no-context-files --no-extensions --no-skills --no-prompt-templates; do
  assert_contains "$FAKE_PI_ARGS" "$flag"
done
assert_contains "$FAKE_PI_STDIN" 'Current release: v0.1.2'
assert_contains "$FAKE_PI_STDIN" 'Previous release: v0.1.1'
assert_contains "$FAKE_PI_STDIN" 'Fix visible behavior'
assert_contains "$FAKE_PI_STDIN" 'diff --git'
assert_contains "$draft" '## Changes'
assert_contains "$draft" '**Full Changelog**: https://github.com/example/cf-redirect/compare/v0.1.1...v0.1.2'
pass 'generates the default tag with constrained Pi and complete evidence'

# Existing drafts are protected; --force replaces and invalidates review.
printf 'keep me\n' > "$draft"
assert_fails run_script "$repo" generate v0.1.2 > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_eq "$(cat "$draft")" 'keep me' 'existing draft was modified without --force'
PAGER=viewer run_script "$repo" inspect v0.1.2 > "$suite_tmp/inspect-keep"
inspection="${draft%.md}.inspected"
[[ -f "$inspection" ]] || fail_test 'inspection record was not created'
FAKE_PI_OUTPUT='Replacement notes.' run_script "$repo" generate v0.1.2 --force > "$suite_tmp/out"
assert_contains "$draft" 'Replacement notes.'
[[ ! -e "$inspection" ]] || fail_test 'forced generation did not invalidate inspection'
pass 'protects drafts and invalidates inspection on forced generation'

# Failed and empty Pi calls preserve an existing draft and leave no transient files.
printf 'stable draft\n' > "$draft"
FAKE_PI_MODE=fail assert_fails run_script "$repo" generate v0.1.2 --force > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_eq "$(cat "$draft")" 'stable draft' 'Pi failure replaced the draft'
FAKE_PI_MODE=empty assert_fails run_script "$repo" generate v0.1.2 --force > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_eq "$(cat "$draft")" 'stable draft' 'empty Pi output replaced the draft'
if find "$(dirname "$draft")" -type f \( -name '*.evidence.*' -o -name '*.pi.*' -o -name '*.draft.*' \) | grep -q .; then
  fail_test 'generation left partial files behind'
fi
pass 'generation failures are atomic'

# Explicit tags, invalid tags, absent tags, missing predecessors, and ancestry selection.
FAKE_PI_MODE=success run_script "$repo" generate v0.1.1 --force > "$suite_tmp/generate-explicit"
explicit_draft=$(draft_path_from_output "$suite_tmp/generate-explicit")
assert_contains "$explicit_draft" '/compare/v0.1.0...v0.1.1'
assert_fails run_script "$repo" generate 0.1.2 > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'stable semantic version'
assert_fails run_script "$repo" generate v9.9.9 > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'local release tag does not exist'
assert_fails run_script "$repo" generate v0.1.0 > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'no preceding stable semantic-version tag'

git -C "$repo" checkout -qb side v0.1.1
echo side >> "$repo/feature.txt"
git -C "$repo" commit -qam 'Side release'
git -C "$repo" tag v0.1.3
git -C "$repo" checkout -q master 2>/dev/null || git -C "$repo" checkout -q main
printf '0.1.4\n' > "$repo/internal/version/VERSION"
echo fourth >> "$repo/feature.txt"
git -C "$repo" add . && git -C "$repo" commit -qm 'Later main release'
git -C "$repo" tag v0.1.4
run_script "$repo" generate v0.1.4 --force > "$suite_tmp/generate-ancestry"
ancestry_draft=$(draft_path_from_output "$suite_tmp/generate-ancestry")
assert_contains "$ancestry_draft" '/compare/v0.1.2...v0.1.4'
assert_not_contains "$ancestry_draft" '/compare/v0.1.3...v0.1.4'
pass 'validates tags and chooses the preceding ancestor'

# Separate repositories receive separate temporary roots.
run_script "$other_repo" generate --force > "$suite_tmp/generate-other"
other_draft=$(draft_path_from_output "$suite_tmp/generate-other")
[[ "$(dirname "$other_draft")" != "$(dirname "$draft")" ]] || fail_test 'separate repositories shared a state directory'
pass 'isolates temporary state between repositories'

# Inspection with pager/editor, failures, and empty edits.
printf 'Draft to review.\n' > "$draft"
PAGER=viewer run_script "$repo" inspect v0.1.2 > "$suite_tmp/inspect"
assert_eq "$(cat "$FAKE_VIEWED_PATH")" "$draft" 'pager received wrong draft'
[[ -f "$inspection" ]] || fail_test 'pager inspection was not recorded'
VISUAL=editor FAKE_EDIT_MODE=append run_script "$repo" inspect v0.1.2 --edit > "$suite_tmp/inspect-edit"
assert_contains "$draft" 'Human edit.'
reviewed_digest=$(cat "$inspection")
assert_eq "$reviewed_digest" "$(git hash-object "$draft")" 'edited draft digest'
FAKE_VIEW_FAIL=true PAGER=viewer assert_fails run_script "$repo" inspect v0.1.2 > "$suite_tmp/out" 2> "$suite_tmp/err"
[[ ! -e "$inspection" ]] || fail_test 'failed inspection left a review record'
printf 'Draft again.\n' > "$draft"
VISUAL=editor FAKE_EDIT_MODE=empty assert_fails run_script "$repo" inspect v0.1.2 --edit > "$suite_tmp/out" 2> "$suite_tmp/err"
[[ ! -e "$inspection" ]] || fail_test 'empty edit left a review record'
pass 'records only successful non-empty inspections'

# Recreate an inspected draft for publication tests.
printf 'Reviewed release body.\n' > "$draft"
printf 'Old remote body.\n' > "$FAKE_GH_BODY"
PAGER=viewer run_script "$repo" inspect v0.1.2 > "$suite_tmp/inspect-publish"
assert_contains "$suite_tmp/inspect-publish" 'Proposed GitHub release body change:'
assert_contains "$suite_tmp/inspect-publish" '-Old remote body.'
assert_contains "$suite_tmp/inspect-publish" '+Reviewed release body.'

# Changed drafts and missing releases fail without edits.
printf 'Changed after review.\n' >> "$draft"
: > "$FAKE_GH_LOG"
assert_fails run_script "$repo" publish v0.1.2 --yes > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'draft changed after inspection'
assert_not_contains "$FAKE_GH_LOG" 'release edit'
printf 'Reviewed release body.\n' > "$draft"
PAGER=viewer run_script "$repo" inspect v0.1.2 > /dev/null
FAKE_GH_RELEASE_EXISTS=false assert_fails run_script "$repo" publish v0.1.2 --yes > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'GitHub release does not exist'
pass 'rejects unreviewed changes and absent releases'

# Non-interactive publication, exact notes file, and verification.
printf 'Old remote body.\n' > "$FAKE_GH_BODY"
: > "$FAKE_GH_LOG"
run_script "$repo" publish v0.1.2 --yes > "$suite_tmp/publish"
cmp -s "$draft" "$FAKE_GH_BODY" || fail_test 'published body does not match draft'
assert_contains "$FAKE_GH_LOG" "release edit v0.1.2 --repo example/cf-redirect --notes-file $draft"
assert_contains "$suite_tmp/publish" 'Published and verified release notes for v0.1.2.'
pass 'publishes and verifies reviewed notes non-interactively'

# Update and verification errors are reported.
printf 'Old remote body.\n' > "$FAKE_GH_BODY"
FAKE_GH_EDIT_FAIL=true assert_fails run_script "$repo" publish v0.1.2 --yes > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'GitHub rejected'
printf 'Old remote body.\n' > "$FAKE_GH_BODY"
FAKE_GH_VERIFY_MISMATCH=true assert_fails run_script "$repo" publish v0.1.2 --yes > "$suite_tmp/out" 2> "$suite_tmp/err"
assert_contains "$suite_tmp/err" 'does not exactly match'
pass 'reports publication and verification failures'

# Interactive cancellation uses a real pseudo-terminal when script(1) is available.
if command -v script >/dev/null 2>&1; then
  printf 'Old remote body.\n' > "$FAKE_GH_BODY"
  : > "$FAKE_GH_LOG"
  printf 'n\n' | script -q "$suite_tmp/tty-output" bash -c "cd '$repo' && ./scripts/release-notes publish v0.1.2" >/dev/null
  assert_not_contains "$FAKE_GH_LOG" 'release edit'
  assert_contains "$suite_tmp/tty-output" 'Publication cancelled'
  pass 'cancels interactive publication without mutation'
fi

printf '1..%d\n' "$tests"
