package app

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/koopycat/cf-redirect/internal/cloudflare"
	"github.com/koopycat/cf-redirect/internal/domain"
	"github.com/koopycat/cf-redirect/internal/planner"
)

type fakeAPI struct {
	current       []domain.Redirect
	calls         []string
	deleted       []string
	deleteBatches [][]string
	created       []domain.Redirect
	createErrors  []error
	createCalls   int
	waitError     map[string]error
}

func (f *fakeAPI) ListItems(context.Context, string, string) ([]domain.Redirect, error) {
	f.calls = append(f.calls, "list")
	return f.current, nil
}
func (f *fakeAPI) DeleteItems(_ context.Context, _, _ string, ids []string) (cloudflare.BulkOperation, error) {
	f.calls = append(f.calls, "delete")
	f.deleted = append(f.deleted, ids...)
	f.deleteBatches = append(f.deleteBatches, append([]string(nil), ids...))
	return cloudflare.BulkOperation{ID: fmt.Sprintf("delete-op-%d", len(f.deleteBatches)), Status: "pending"}, nil
}
func (f *fakeAPI) CreateItems(_ context.Context, _, _ string, items []domain.Redirect) (cloudflare.BulkOperation, error) {
	f.calls = append(f.calls, "create")
	f.createCalls++
	if len(f.createErrors) > 0 {
		err := f.createErrors[0]
		f.createErrors = f.createErrors[1:]
		if err != nil {
			return cloudflare.BulkOperation{}, err
		}
	}
	f.created = append(f.created, items...)
	return cloudflare.BulkOperation{ID: "create-op", Status: "pending"}, nil
}
func (f *fakeAPI) WaitBulkOperation(_ context.Context, _, id string, _ time.Duration) (cloudflare.BulkOperation, error) {
	f.calls = append(f.calls, "wait-"+id)
	if err := f.waitError[id]; err != nil {
		return cloudflare.BulkOperation{ID: id, Status: "failed"}, err
	}
	return cloudflare.BulkOperation{ID: id, Status: "completed"}, nil
}

func oldItem() domain.Redirect {
	return domain.Redirect{ID: "old-id", Source: "https://old.example", Target: "https://target.example/old", StatusCode: 308, PreserveQueryString: true, Comment: "keep"}
}

func TestExecutorUpsertsUpdatesWithUnchangedSourcesWithoutDeleting(t *testing.T) {
	old := oldItem()
	after := old
	after.Target = "https://target.example/new"
	add := domain.New("https://add.example", "https://target.example/add")
	plan := planner.Plan{Changes: []planner.Change{
		{Kind: planner.Update, Before: &old, After: &after},
		{Kind: planner.Add, After: &add},
	}}
	api := &fakeAPI{current: []domain.Redirect{old}, waitError: map[string]error{}}
	report, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{"list", "create", "wait-create-op"}
	if !reflect.DeepEqual(api.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", api.calls, wantCalls)
	}
	if len(api.deleted) != 0 || len(api.created) != 2 || api.created[0].ID != "" || !api.created[0].PreserveQueryString || api.created[0].Comment != "keep" || !api.created[1].PreserveQueryString {
		t.Fatalf("bad requests: deleted=%v created=%#v", api.deleted, api.created)
	}
	if len(report.Phases) != 1 || report.Phases[0].Phase != CreatePhase || !report.Phases[0].Completed {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestExecutorDeletesSourceChangingUpdateBeforeCreating(t *testing.T) {
	old := oldItem()
	after := old
	after.Source = "https://new.example"
	plan := planner.Plan{Changes: []planner.Change{{Kind: planner.Update, Before: &old, After: &after}}}
	api := &fakeAPI{current: []domain.Redirect{old}, waitError: map[string]error{}}
	_, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(context.Background(), plan)
	if err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{"list", "delete", "wait-delete-op-1", "create", "wait-create-op"}
	if !reflect.DeepEqual(api.calls, wantCalls) || !reflect.DeepEqual(api.deleted, []string{"old-id"}) {
		t.Fatalf("calls = %v, deleted = %v", api.calls, api.deleted)
	}
}

func TestExecutorBatchesLargeDeletesAndWaitsBetweenBatches(t *testing.T) {
	current := make([]domain.Redirect, mutationBatchSize+1)
	changes := make([]planner.Change, len(current))
	for i := range current {
		current[i] = domain.Redirect{
			ID:         fmt.Sprintf("id-%04d", i),
			Source:     fmt.Sprintf("https://source-%04d.example", i),
			Target:     "https://target.example",
			StatusCode: 301,
		}
		item := current[i]
		changes[i] = planner.Change{Kind: planner.Delete, Before: &item}
	}
	api := &fakeAPI{current: current, waitError: map[string]error{}}
	report, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(context.Background(), planner.Plan{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{"list", "delete", "wait-delete-op-1", "delete", "wait-delete-op-2"}
	if !reflect.DeepEqual(api.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", api.calls, wantCalls)
	}
	if len(api.deleteBatches) != 2 || len(api.deleteBatches[0]) != mutationBatchSize || len(api.deleteBatches[1]) != 1 {
		t.Fatalf("delete batches = %v", slicesToLengths(api.deleteBatches))
	}
	if len(report.Phases) != 2 || !report.Phases[0].Completed || report.Phases[0].Requested != mutationBatchSize || !report.Phases[1].Completed || report.Phases[1].Requested != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func slicesToLengths[T any](items [][]T) []int {
	lengths := make([]int, len(items))
	for i := range items {
		lengths[i] = len(items[i])
	}
	return lengths
}

func TestExecutorBatchesLargeCreatesAndWaitsBetweenBatches(t *testing.T) {
	changes := make([]planner.Change, mutationBatchSize+1)
	for i := range changes {
		item := domain.New(fmt.Sprintf("https://source-%04d.example", i), "https://target.example")
		changes[i] = planner.Change{Kind: planner.Add, After: &item}
	}
	api := &fakeAPI{waitError: map[string]error{}}
	report, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(context.Background(), planner.Plan{Changes: changes})
	if err != nil {
		t.Fatal(err)
	}
	wantCalls := []string{"list", "create", "wait-create-op", "create", "wait-create-op"}
	if !reflect.DeepEqual(api.calls, wantCalls) {
		t.Fatalf("calls = %v, want %v", api.calls, wantCalls)
	}
	if len(api.created) != mutationBatchSize+1 || len(report.Phases) != 2 || report.Phases[0].Requested != mutationBatchSize || report.Phases[1].Requested != 1 {
		t.Fatalf("created=%d report=%#v", len(api.created), report)
	}
}

func TestExecutorRetriesRateLimitedMutationAndReportsLiveWait(t *testing.T) {
	item := domain.New("https://source.example", "https://target.example")
	rateLimit := &cloudflare.APIError{StatusCode: 429, HasRetryAfter: true}
	api := &fakeAPI{createErrors: []error{rateLimit}, waitError: map[string]error{}}
	var progress []string

	report, err := (Executor{
		API: api, AccountID: "account", ListID: "list", RateLimitBaseDelay: time.Nanosecond,
		Progress: func(message string) { progress = append(progress, message) },
	}).Apply(context.Background(), planner.Plan{Changes: []planner.Change{{Kind: planner.Add, After: &item}}})
	if err != nil {
		t.Fatal(err)
	}
	if api.createCalls != 2 || len(api.created) != 1 || len(report.Phases) != 1 || !report.Phases[0].Completed {
		t.Fatalf("retry result: calls=%d created=%d report=%#v", api.createCalls, len(api.created), report)
	}
	joined := strings.Join(progress, "\n")
	for _, want := range []string{"rate limit reached", "retrying batch 1/1", "attempt 2/8"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("progress %q does not contain %q", joined, want)
		}
	}
}

func TestExecutorRateLimitWaitCanBeCancelled(t *testing.T) {
	item := domain.New("https://source.example", "https://target.example")
	rateLimit := &cloudflare.APIError{StatusCode: 429, HasRetryAfter: true, RetryAfter: time.Minute}
	api := &fakeAPI{createErrors: []error{rateLimit}, waitError: map[string]error{}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(
		ctx, planner.Plan{Changes: []planner.Change{{Kind: planner.Add, After: &item}}},
	)
	if !errors.Is(err, context.Canceled) || api.createCalls != 1 {
		t.Fatalf("cancelled retry: calls=%d err=%v", api.createCalls, err)
	}
}

func TestExecutorStopsAfterRateLimitRetryBudget(t *testing.T) {
	item := domain.New("https://source.example", "https://target.example")
	errors429 := make([]error, defaultRateLimitAttempts)
	for i := range errors429 {
		errors429[i] = &cloudflare.APIError{StatusCode: 429, HasRetryAfter: true}
	}
	api := &fakeAPI{createErrors: errors429, waitError: map[string]error{}}

	_, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(
		context.Background(), planner.Plan{Changes: []planner.Change{{Kind: planner.Add, After: &item}}},
	)
	var apiErr *cloudflare.APIError
	if !errors.As(err, &apiErr) || api.createCalls != defaultRateLimitAttempts {
		t.Fatalf("exhausted retry: calls=%d err=%v", api.createCalls, err)
	}
}

func TestExecutorDoesNotRetryNonRateLimitMutationFailure(t *testing.T) {
	item := domain.New("https://source.example", "https://target.example")
	api := &fakeAPI{createErrors: []error{&cloudflare.APIError{StatusCode: 503}}, waitError: map[string]error{}}
	_, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(
		context.Background(), planner.Plan{Changes: []planner.Change{{Kind: planner.Add, After: &item}}},
	)
	if err == nil || api.createCalls != 1 {
		t.Fatalf("non-rate-limit failure: calls=%d err=%v", api.createCalls, err)
	}
}

func TestExecutorReportsPartialFailure(t *testing.T) {
	old := oldItem()
	after := old
	after.Source = "https://new.example"
	plan := planner.Plan{Changes: []planner.Change{{Kind: planner.Update, Before: &old, After: &after}}}
	api := &fakeAPI{current: []domain.Redirect{old}, waitError: map[string]error{"create-op": errors.New("create failed")}}
	report, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(context.Background(), plan)
	if err == nil || !report.Partial() || len(report.Phases) != 2 || !report.Phases[0].Completed || report.Phases[1].Err == nil {
		t.Fatalf("expected partial failure, report=%#v err=%v", report, err)
	}
	var executionError *ExecutionError
	if !errors.As(err, &executionError) || !executionError.Report.Partial() {
		t.Fatalf("missing report in error: %v", err)
	}
}

func TestExecutorRejectsStalePlanBeforeMutation(t *testing.T) {
	old := oldItem()
	after := old
	after.Target = "https://target.example/new"
	plan := planner.Plan{Changes: []planner.Change{{Kind: planner.Update, Before: &old, After: &after}}}
	stale := old
	stale.Comment = "someone changed this"
	api := &fakeAPI{current: []domain.Redirect{stale}, waitError: map[string]error{}}
	report, err := (Executor{API: api, AccountID: "account", ListID: "list"}).Apply(context.Background(), plan)
	if err == nil || report.Revalidated || !reflect.DeepEqual(api.calls, []string{"list"}) {
		t.Fatalf("stale apply should not mutate: report=%#v err=%v calls=%v", report, err, api.calls)
	}
}

func TestRevalidateDetectsSourceCollision(t *testing.T) {
	add := domain.New("https://add.example", "https://target.example")
	plan := planner.Plan{Changes: []planner.Change{{Kind: planner.Add, After: &add}}}
	current := []domain.Redirect{{ID: "new-id", Source: add.Source, Target: "https://other.example", StatusCode: 301}}
	if err := Revalidate(plan, current); err == nil {
		t.Fatal("expected collision")
	}
}
