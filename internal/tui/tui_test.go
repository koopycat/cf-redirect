package tui

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/color"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/koopycat/cf-redirect/internal/domain"
	"github.com/koopycat/cf-redirect/internal/planner"
)

func readyModel(items ...domain.Redirect) model {
	m := newModel(nil, "account", "list")
	m.loading = false
	m.items = items
	m.redirects.SetItems(redirectItems(items))
	m.width = 80
	m.height = 24
	m.resizeList()
	return m
}

func keyPress(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func typeText(t *testing.T, m model, value string) model {
	t.Helper()
	for _, r := range value {
		m = update(t, m, keyPress(r))
	}
	return m
}

func update(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	return updated.(model)
}

func TestApplyErrorPersistsThroughReloadUntilDismissed(t *testing.T) {
	m := readyModel()
	// Simulate: apply fails, then a successful reload arrives. The error must
	// survive the reload so the user can read it.
	first := update(t, m, appliedMsg{err: errors.New("bulk operation ended with status failed")})
	got := update(t, first, loadedMsg{items: nil, err: nil})
	if got.err == nil {
		t.Fatal("apply error was cleared by the reload before the user read it")
	}
	if !strings.Contains(got.View().Content, "bulk operation ended with status failed") {
		t.Fatal("error message not visible in the TUI view")
	}
	// Any key acknowledges and hides the error.
	got = update(t, got, keyPress(tea.KeySpace))
	if got.err != nil {
		t.Fatal("error not dismissed by a key")
	}
}

func TestListFiltersSourceTargetAndComment(t *testing.T) {
	m := readyModel(
		domain.Redirect{ID: "one", Source: "https://one.example", Target: "https://target.example", StatusCode: 301, Comment: "primary"},
		domain.Redirect{ID: "two", Source: "https://two.example", Target: "https://else.example", StatusCode: 301, Comment: "secondary"},
	)

	m = update(t, m, keyPress('/'))
	if !m.redirects.SettingFilter() {
		t.Fatal("/ did not activate list filtering")
	}
	m = typeText(t, m, "PRIMARY")
	// Bubbles computes filters in a command. Route the result back through the
	// model just as Bubble Tea does.
	m.redirects.SetFilterText(m.redirects.FilterValue())

	visible := m.redirects.VisibleItems()
	if len(visible) != 1 || visible[0].(redirectItem).redirect.ID != "one" {
		t.Fatalf("visible items = %#v", visible)
	}
	item, ok := m.current()
	if !ok || item.ID != "one" {
		t.Fatalf("current = %#v, %v", item, ok)
	}
}

func TestListOwnsNavigationAndSelection(t *testing.T) {
	m := readyModel(
		domain.Redirect{ID: "one", Source: "https://one.example", Target: "https://target.example", StatusCode: 301},
		domain.Redirect{ID: "two", Source: "https://two.example", Target: "https://target.example", StatusCode: 301},
	)
	m = update(t, m, keyPress('j'))
	item, ok := m.current()
	if !ok || item.ID != "two" || m.redirects.Index() != 1 {
		t.Fatalf("selected = %#v, %v; index = %d", item, ok, m.redirects.Index())
	}
}

func TestResizeChangesListCapacityAndNarrowLayout(t *testing.T) {
	items := make([]domain.Redirect, 12)
	for i := range items {
		items[i] = domain.Redirect{ID: fmt.Sprint(i), Source: fmt.Sprintf("https://source-%d.example", i), Target: fmt.Sprintf("https://target-%d.example", i)}
	}
	m := readyModel(items...)
	m = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 12})
	widePerPage := m.redirects.Paginator.PerPage
	if got := m.redirects.SelectedItem().(redirectItem).redirect.ID; got != "0" {
		t.Fatalf("selection after resize = %q", got)
	}

	m = update(t, m, tea.WindowSizeMsg{Width: 50, Height: 12})
	if narrowPerPage := m.redirects.Paginator.PerPage; narrowPerPage >= widePerPage {
		t.Fatalf("narrow per-page = %d, wide = %d", narrowPerPage, widePerPage)
	}
	var rendered bytes.Buffer
	redirectDelegate{narrow: true}.Render(&rendered, m.redirects, 0, m.redirects.Items()[0])
	if got := rendered.String(); !strings.Contains(got, "\n") || lipgloss.Width(got) > 50 {
		t.Fatalf("narrow item did not stack within terminal: %q", got)
	}
}

func TestLoadAfterResizeFitsScreenWithHelp(t *testing.T) {
	items := make([]domain.Redirect, 30)
	for i := range items {
		items[i] = domain.Redirect{ID: fmt.Sprint(i), Source: fmt.Sprintf("example.com/path-%02d", i), Target: "https://target.example"}
	}
	m := newModel(nil, "account", "list")
	m = update(t, m, tea.WindowSizeMsg{Width: 100, Height: 20})
	m = update(t, m, loadedMsg{items: items})
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	if len(lines) > 20 || !strings.Contains(lines[len(lines)-1], "quit") {
		t.Fatalf("view has %d lines for a 20-line terminal; last line %q", len(lines), lines[len(lines)-1])
	}
}

func TestReloadPreservesFilteredSelectionByID(t *testing.T) {
	m := readyModel(
		domain.Redirect{ID: "one", Source: "match-one.example", Target: "https://target.example/one"},
		domain.Redirect{ID: "other", Source: "other.example", Target: "https://target.example/other"},
		domain.Redirect{ID: "two", Source: "match-two.example", Target: "https://target.example/two"},
	)
	m.redirects.SetFilterText("match")
	m.redirects.Select(1)
	if current, ok := m.current(); !ok || current.ID != "two" {
		t.Fatalf("precondition selection = %#v, %v", current, ok)
	}

	m = update(t, m, loadedMsg{items: []domain.Redirect{
		{ID: "inserted", Source: "inserted.example", Target: "https://target.example/inserted"},
		{ID: "one", Source: "match-one.example", Target: "https://target.example/one"},
		{ID: "other", Source: "other.example", Target: "https://target.example/other"},
		{ID: "two", Source: "match-two.example", Target: "https://target.example/two"},
	}})
	if current, ok := m.current(); !ok || current.ID != "two" {
		t.Fatalf("selection after reload = %#v, %v", current, ok)
	}
}

func TestDelegateTruncatesUnicodeByDisplayWidth(t *testing.T) {
	item := redirectItem{redirect: domain.Redirect{Source: "https://例え.example/very/long/source", Target: "https://target.example/very/long/path"}}
	l := list.New([]list.Item{item}, redirectDelegate{}, 24, 8)
	l.SetShowTitle(false)
	l.SetShowFilter(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	var rendered bytes.Buffer
	redirectDelegate{}.Render(&rendered, l, 0, item)
	if got := lipgloss.Width(rendered.String()); got > 24 {
		t.Fatalf("rendered width = %d, want <= 24: %q", got, rendered.String())
	}
}

func TestPlanRequiresYAndSanitizesOutput(t *testing.T) {
	m := readyModel()
	m.mode = planMode
	before := domain.Redirect{ID: "one", Source: "old.example", Target: "https://old-target.example", StatusCode: 301}
	after := before
	after.Source = "new.example"
	after.Target = "https://new-target.example"
	m.plan = planner.Plan{Changes: []planner.Change{{Kind: planner.Update, Before: &before, After: &after}}}
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := updated.(model)
	if got.applying || cmd != nil {
		t.Fatal("Enter must not apply a plan")
	}
	view := got.planView()
	// Lip Gloss v2 always renders full color; Bubble Tea downsamples on output.
	plainView := strings.ReplaceAll(ansi.Strip(view), "\n", "")
	for _, want := range []string{before.Source, before.Target, after.Source, after.Target, "Press y"} {
		if !strings.Contains(plainView, want) {
			t.Fatalf("plan view %q missing %q", view, want)
		}
	}
	if sanitized := safeTerminalText("safe\x1b[2J\ntext"); sanitized != "safe[2Jtext" {
		t.Fatalf("sanitized text = %q", sanitized)
	}
}

func TestQuitOrEscapeDuringApplyCancelsLocalWait(t *testing.T) {
	for _, keyMsg := range []tea.KeyPressMsg{keyPress('q'), {Code: tea.KeyEscape}} {
		m := readyModel()
		m.applying = true
		ctx, cancel := context.WithCancel(context.Background())
		m.applyCancel = cancel
		updated, cmd := m.Update(keyMsg)
		got := updated.(model)
		if cmd != nil || !got.applying || got.applyCancel != nil || !strings.Contains(got.status, "Stopping local wait") {
			t.Fatalf("%q did not start cancellation: applying=%v status=%q cmd=%v", keyMsg.String(), got.applying, got.status, cmd)
		}
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Fatalf("%q did not cancel apply context", keyMsg.String())
		}
	}
}

func TestCancelledApplyShowsHonestExitMessage(t *testing.T) {
	m := readyModel()
	m.applying = true
	got := update(t, m, appliedMsg{err: fmt.Errorf("apply plan: %w", context.Canceled)})
	if got.applying || !got.interrupted || got.loading || !strings.Contains(got.status, "Cloudflare operation may still finish") {
		t.Fatalf("cancelled apply state: applying=%v interrupted=%v loading=%v status=%q", got.applying, got.interrupted, got.loading, got.status)
	}
	_, cmd := got.Update(keyPress('q'))
	if cmd == nil {
		t.Fatal("q must exit after local wait is cancelled")
	}
}

func TestApplyingViewShowsLiveProgress(t *testing.T) {
	m := readyModel()
	m.width = 100
	m.height = 30
	m.applying = true
	m.progress.set("create phase: waiting for operation op-123…")
	view := m.View().Content
	for _, want := range []string{"create phase", "op-123", "q/esc stops waiting locally"} {
		if !strings.Contains(view, want) {
			t.Fatalf("applying view missing %q: %q", want, view)
		}
	}
}

func TestListHelpContainsRedirectActions(t *testing.T) {
	m := readyModel(domain.Redirect{ID: "one", Source: "one", Target: "two"})
	view := m.View().Content
	for _, want := range []string{"search", "add", "edit", "delete", "quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("list help missing %q: %q", want, view)
		}
	}
}

func TestPasteFillsFocusedFormFieldAndFilter(t *testing.T) {
	m := readyModel(
		domain.Redirect{ID: "one", Source: "https://one.example", Target: "https://target.example"},
		domain.Redirect{ID: "two", Source: "https://two.example", Target: "https://target.example"},
	)
	m = update(t, m, keyPress('/'))
	m = update(t, m, tea.PasteMsg{Content: "two"})
	if got := m.redirects.FilterValue(); got != "two" {
		t.Fatalf("filter after paste = %q", got)
	}
	m = update(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})

	m = update(t, m, keyPress('a'))
	m = update(t, m, tea.PasteMsg{Content: "https://pasted.example"})
	m = update(t, m, tea.KeyPressMsg{Code: tea.KeyEnter})
	m = update(t, m, tea.PasteMsg{Content: "https://pasted-target.example"})
	if m.source.Value() != "https://pasted.example" || m.target.Value() != "https://pasted-target.example" {
		t.Fatalf("form after paste: source=%q target=%q", m.source.Value(), m.target.Value())
	}
}

func TestPasteIsIgnoredWhileAnErrorIsShown(t *testing.T) {
	m := readyModel()
	m.mode = formMode
	m.err = errors.New("boom")
	m = update(t, m, tea.PasteMsg{Content: "https://pasted.example"})
	if m.source.Value() != "" || m.err == nil {
		t.Fatalf("paste reached the form behind an error: source=%q err=%v", m.source.Value(), m.err)
	}
}

func TestBackgroundColorSelectsLightOrDarkStyles(t *testing.T) {
	m := readyModel(domain.Redirect{ID: "one", Source: "one", Target: "two"})
	if !m.styles.isDark {
		t.Fatal("styles must default to dark until the terminal reports its background")
	}
	dark := m.View().Content
	m = update(t, m, tea.BackgroundColorMsg{Color: color.White})
	if m.styles.isDark {
		t.Fatal("white background did not select light styles")
	}
	if light := m.View().Content; light == dark || ansi.Strip(light) != ansi.Strip(dark) {
		t.Fatal("light styles must change only the colors of the view")
	}
}
