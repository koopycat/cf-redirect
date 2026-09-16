package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zalando/go-keyring"

	"github.com/koopycat/cf-redirect/internal/auth"
	"github.com/koopycat/cf-redirect/internal/domain"
	"github.com/koopycat/cf-redirect/internal/planner"
	"github.com/koopycat/cf-redirect/internal/textsafe"
	"github.com/koopycat/cf-redirect/internal/version"
	"github.com/spf13/cobra"
)

func TestRenderPlanIncludesDeterministicMarkersAndCounts(t *testing.T) {
	old := domain.Redirect{ID: "one", Source: "https://old.example", Target: "https://before.example", StatusCode: 301}
	updated := old
	updated.Target = "https://after.example"
	added := domain.New("https://add.example", "https://target.example")
	plan := planner.Plan{
		Changes: []planner.Change{
			{Kind: planner.Add, After: &added},
			{Kind: planner.Update, Before: &old, After: &updated},
			{Kind: planner.Delete, Before: &old},
		},
		SkippedExisting: 2,
	}
	var output bytes.Buffer
	if err := renderPlan(&output, plan); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Plan: 1 add, 1 update, 1 delete, 2 skipped existing", "+ https://add.example", "~ https://old.example -> https://before.example => https://old.example -> https://after.example", "- https://old.example"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("plan output %q does not contain %q", output.String(), want)
		}
	}
}

func TestRenderPlanShowsAllSkippedImport(t *testing.T) {
	plan := planner.Plan{SkippedExisting: 3}
	if !plan.Empty() {
		t.Fatal("reporting skipped rows must not make a plan actionable")
	}
	var output bytes.Buffer
	if err := renderPlan(&output, plan); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "Plan: 0 add, 0 update, 0 delete, 3 skipped existing\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestRenderPlanOmitsSkippedCountWhenZero(t *testing.T) {
	var output bytes.Buffer
	if err := renderPlan(&output, planner.Plan{}); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "Plan: 0 add, 0 update, 0 delete\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

// A failing output must never let a mutation proceed without a visible plan.
func TestRenderPlanFailsOnWriterError(t *testing.T) {
	added := domain.New("https://add.example", "https://target.example")
	plan := planner.Plan{Changes: []planner.Change{{Kind: planner.Add, After: &added}}}
	if err := renderPlan(failingWriter{}, plan); err == nil {
		t.Fatal("renderPlan must report a failing output writer")
	}
}

// Non-terminal stdin must trigger a clear error instead of a hidden prompt.
func TestReadPasswordInteractiveRequiresTTY(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetIn(strings.NewReader("not-a-tty\n"))
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if _, err := readPasswordInteractive(cmd, "Cloudflare API token: "); err == nil {
		t.Fatal("readPasswordInteractive must fail when stdin is not a terminal")
	}
}

func TestFindSourceRequiresExactMatchAndID(t *testing.T) {
	items := []domain.Redirect{{ID: "id-1", Source: "https://example.com/a", Target: "https://target.example", StatusCode: 301}}
	item, err := findSource(items, "https://example.com/a")
	if err != nil || item.ID != "id-1" {
		t.Fatalf("findSource() = %#v, %v", item, err)
	}
	if _, err := findSource(items, "https://example.com"); err == nil {
		t.Fatal("partial source must not resolve")
	}
}

func TestRenderPlanSanitizesRemoteText(t *testing.T) {
	item := domain.New("source\x1b[2J", "target\x07")
	plan := planner.Plan{Changes: []planner.Change{{Kind: planner.Add, After: &item}}}
	var output bytes.Buffer
	if err := renderPlan(&output, plan); err != nil {
		t.Fatal(err)
	}
	if strings.ContainsRune(output.String(), '\x1b') || strings.ContainsRune(output.String(), '\x07') {
		t.Fatalf("plan output retained control characters: %q", output.String())
	}
}

func TestTerminalCommentSanitization(t *testing.T) {
	if got := textsafe.StripControls("normal\x1b[31mred\x07\nline"); got != "normal[31mredline" {
		t.Fatalf("StripControls() = %q", got)
	}
	items := []domain.Redirect{{Source: "example.com\x1b[2J", Target: "https://target.example\x07", StatusCode: 301, Comment: "unsafe\x1b[2Jcomment"}}
	for _, format := range []string{"table", "json", "csv"} {
		var output bytes.Buffer
		if err := renderRedirects(&output, items, format); err != nil {
			t.Fatal(err)
		}
		if strings.ContainsRune(output.String(), '\x1b') {
			t.Fatalf("%s output retained escape: %q", format, output.String())
		}
	}
}

func TestClearCommandRequiresNoArgumentsAndHasMutationGuards(t *testing.T) {
	root := NewRootCmd()
	clear, _, err := root.Find([]string{"clear"})
	if err != nil {
		t.Fatal(err)
	}
	if clear.Args == nil || clear.Args(clear, []string{"unexpected"}) == nil {
		t.Fatal("clear must reject positional arguments")
	}
	for _, name := range []string{"dry-run", "yes"} {
		if clear.Flags().Lookup(name) == nil {
			t.Fatalf("clear is missing --%s", name)
		}
	}
}

func TestLinuxKeyringHelpExplainsHeadlessSetupAndEnvironmentFallback(t *testing.T) {
	failure := fmt.Errorf("%w: Secret Service is not installed", auth.ErrKeyringUnavailable)
	got := keyringErrorHelpForOS(failure, "linux").Error()
	for _, want := range []string{
		"sudo apt install dbus-user-session gnome-keyring",
		"CLOUDFLARE_API_TOKEN",
		"docs/authentication.md#headless-linux-and-wsl",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("error %q does not contain %q", got, want)
		}
	}
}

func TestKeyringHelpLeavesNonKeyringErrorsUnchanged(t *testing.T) {
	failure := errors.New("token is too large")
	if got := keyringErrorHelpForOS(failure, "linux"); got != failure {
		t.Fatalf("keyringErrorHelpForOS() = %v, want original error", got)
	}
}

func TestLogoutUsesAccountFlagWithoutListIDOrPersistedConfig(t *testing.T) {
	t.Setenv("CLOUDFLARE_LIST_ID", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	keyring.MockInit()
	if err := keyring.Set(auth.KeyringService, auth.KeyringUser+":review-account", "account-token"); err != nil {
		t.Fatal(err)
	}

	root := NewRootCmd()
	root.SetArgs([]string{"--account-id", "review-account", "logout"})
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, err := keyring.Get(auth.KeyringService, auth.KeyringUser+":review-account"); !errors.Is(err, keyring.ErrNotFound) || got != "" {
		t.Fatalf("account token still exists: %q, %v", got, err)
	}
	if got, want := output.String(), "Stored API token removed.\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestRootVersionFlag(t *testing.T) {
	root := NewRootCmd()
	root.SetArgs([]string{"--version"})
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "cf-redirect version "+version.String()+"\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestWriteCSVFileCreatesAndAtomicallyReplacesExport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "redirects.csv")
	items := []domain.Redirect{domain.New("example.com/old/", "https://www.example.com/new/")}
	if err := writeCSVFile(path, items); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil {
		t.Fatal(err)
	} else if want := "source,target\nexample.com/old/,https://www.example.com/new/\n"; string(got) != want {
		t.Fatalf("file = %q, want %q", got, want)
	}

	replacement := []domain.Redirect{domain.New("example.com/next/", "https://www.example.com/final/")}
	if err := writeCSVFile(path, replacement); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(path); err != nil {
		t.Fatal(err)
	} else if strings.Contains(string(got), "old") || !strings.Contains(string(got), "next") {
		t.Fatalf("file was not replaced cleanly: %q", got)
	}
}

func TestProgressDisplayWritesStableUpdatesForNonTerminalOutput(t *testing.T) {
	var output bytes.Buffer
	progress := newProgressDisplay(&output, false, 0)
	progress.Start("Starting apply…")
	progress.Update("create phase: adding batch 4/10 (500 item(s))…")
	progress.Update("create phase: Cloudflare rate limit reached; retrying batch 5/10 in 30s…")
	progress.Stop()

	got := output.String()
	for _, want := range []string{"Starting apply…\n", "batch 4/10", "rate limit reached"} {
		if !strings.Contains(got, want) {
			t.Fatalf("progress output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "\r") || strings.Contains(got, "\x1b") {
		t.Fatalf("non-terminal progress contains control sequences: %q", got)
	}
}

func TestProgressDisplayAnimatesAndClearsInteractiveLine(t *testing.T) {
	var output bytes.Buffer
	progress := newProgressDisplay(&output, true, 80)
	progress.Start("create phase: adding batch 4/10 (500 item(s))…")
	progress.Update("create phase: waiting for operation op-123…")
	progress.Stop()

	got := output.String()
	if !strings.Contains(got, "\r\x1b[2K") || !strings.Contains(got, "waiting for operation") {
		t.Fatalf("interactive progress output = %q", got)
	}
}

func TestRootIncludesRequiredCommands(t *testing.T) {
	root := NewRootCmd()
	for _, name := range []string{"list", "search", "export", "add", "edit", "delete", "clear", "import", "config", "auth", "login", "logout", "status", "tui"} {
		if _, _, err := root.Find([]string{name}); err != nil {
			t.Fatalf("command %q is missing: %v", name, err)
		}
	}
}
