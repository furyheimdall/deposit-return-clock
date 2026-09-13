package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/furyheimdall/deposit-return-clock/clock"
)

func TestSeat(t *testing.T) {
	if Seat != "notify" {
		t.Fatalf("notify seat: got %q", Seat)
	}
}

func caJune1(t *testing.T) clock.Result {
	t.Helper()
	res, err := clock.Compute(clock.Input{
		State:       clock.California,
		VacatedOn:   clock.NewDate(2026, time.June, 1),
		LeaseEndsOn: clock.NewDate(2026, time.June, 1),
	})
	if err != nil {
		t.Fatalf("clock.Compute: %v", err)
	}
	return res
}

func TestScheduleDerivedFromClockResult(t *testing.T) {
	res := caJune1(t)
	if !res.DeadlineOn.Equal(clock.NewDate(2026, time.June, 22)) {
		t.Fatalf("canon CA deadline: got %s", res.DeadlineOn)
	}

	got := Schedule(Item{ID: "ca-1", Result: res})
	if len(got) != 3 {
		t.Fatalf("schedule len: got %d want 3", len(got))
	}
	want := []struct {
		kind Kind
		fire clock.Date
	}{
		{KindD7, clock.NewDate(2026, time.June, 15)},
		{KindD3, clock.NewDate(2026, time.June, 19)},
		{KindDue, clock.NewDate(2026, time.June, 22)},
	}
	for i, w := range want {
		if got[i].Kind != w.kind || !got[i].FireOn.Equal(w.fire) {
			t.Fatalf("schedule[%d]: kind=%s fire=%s want %s %s", i, got[i].Kind, got[i].FireOn, w.kind, w.fire)
		}
		if got[i].ID != "ca-1" || !got[i].DeadlineOn.Equal(res.DeadlineOn) {
			t.Fatalf("schedule[%d]: id/deadline mismatch", i)
		}
		if got[i].Result.Rule.ID != "CA-vacate-21" {
			t.Fatalf("schedule[%d]: result not from clock: %s", i, got[i].Result.Rule.ID)
		}
	}
}

func TestScheduleZeroDeadline(t *testing.T) {
	if got := Schedule(Item{ID: "x"}); got != nil {
		t.Fatalf("zero deadline: got %#v", got)
	}
}

func TestCanonOffsets(t *testing.T) {
	if KindD7.DaysBeforeDeadline() != 7 || KindD3.DaysBeforeDeadline() != 3 || KindDue.DaysBeforeDeadline() != 0 {
		t.Fatalf("canon offsets: D-7=%d D-3=%d due=%d", KindD7.DaysBeforeDeadline(), KindD3.DaysBeforeDeadline(), KindDue.DaysBeforeDeadline())
	}
	if Kind("nope").DaysBeforeDeadline() != -1 {
		t.Fatalf("unknown kind should be -1")
	}
}

func TestRemindersOnFakeClock(t *testing.T) {
	item := Item{ID: "ca-1", Result: caJune1(t)}
	cases := []struct {
		today clock.Date
		kind  Kind
	}{
		{clock.NewDate(2026, time.June, 15), KindD7},
		{clock.NewDate(2026, time.June, 19), KindD3},
		{clock.NewDate(2026, time.June, 22), KindDue},
	}
	for _, tc := range cases {
		got := RemindersOn(tc.today, []Item{item})
		if len(got) != 1 || got[0].Kind != tc.kind {
			t.Fatalf("today %s: got %#v want %s", tc.today, got, tc.kind)
		}
	}
	if got := RemindersOn(clock.NewDate(2026, time.June, 16), []Item{item}); len(got) != 0 {
		t.Fatalf("off-day should emit nothing: %#v", got)
	}
	if got := RemindersOn(clock.Date{}, []Item{item}); got != nil {
		t.Fatalf("zero today: %#v", got)
	}
}

func TestDueWithin(t *testing.T) {
	item := Item{ID: "ca-1", Result: caJune1(t)}
	today := clock.NewDate(2026, time.June, 15)

	if got := DueWithin(today, 7, []Item{item}); len(got) != 1 {
		t.Fatalf("within 7 on D-7: got %d", len(got))
	}
	if got := DueWithin(today, 6, []Item{item}); len(got) != 0 {
		t.Fatalf("within 6 should exclude June 22: got %d", len(got))
	}
	if got := DueWithin(clock.NewDate(2026, time.June, 22), 0, []Item{item}); len(got) != 1 {
		t.Fatalf("within 0 on due day: got %d", len(got))
	}
	if got := DueWithin(clock.NewDate(2026, time.June, 23), 7, []Item{item}); len(got) != 0 {
		t.Fatalf("overdue excluded: got %d", len(got))
	}
	if got := DueWithin(today, -1, []Item{item}); got != nil {
		t.Fatalf("negative n: %#v", got)
	}
}

func TestDueWithinUsesClockDeadlinesPerState(t *testing.T) {
	ca, err := NewItem("ca", clock.Input{
		State:       clock.California,
		VacatedOn:   clock.NewDate(2026, time.June, 1),
		LeaseEndsOn: clock.NewDate(2026, time.June, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	ny, err := NewItem("ny", clock.Input{
		State:       clock.NewYork,
		VacatedOn:   clock.NewDate(2026, time.June, 1),
		LeaseEndsOn: clock.NewDate(2026, time.June, 1),
	})
	if err != nil {
		t.Fatal(err)
	}
	// CA deadline 2026-06-22; NY deadline 2026-06-15
	today := clock.NewDate(2026, time.June, 15)
	got := DueWithin(today, 0, []Item{ca, ny})
	if len(got) != 1 || got[0].ID != "ny" {
		t.Fatalf("due today: %#v", got)
	}
	got = DueWithin(today, 7, []Item{ca, ny})
	if len(got) != 2 {
		t.Fatalf("within 7: got %d", len(got))
	}
}

func TestNewItemUnsupportedState(t *testing.T) {
	_, err := NewItem("tx", clock.Input{State: "TX", VacatedOn: clock.NewDate(2026, time.June, 1)})
	if !errors.Is(err, clock.ErrUnsupportedState) {
		t.Fatalf("want unsupported state, got %v", err)
	}
}

func TestRunFireWithFakeClockAndNotifier(t *testing.T) {
	item := Item{ID: "ca-1", Result: caJune1(t)}
	clk := FixedClock{Instant: time.Date(2026, time.June, 15, 12, 0, 0, 0, time.UTC)}
	rec := &RecordingNotifier{}
	rep, err := Run(context.Background(), []Item{item}, Options{
		Clock:    clk,
		Within:   7,
		Fire:     true,
		Notifier: rec,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Today.Equal(clock.NewDate(2026, time.June, 15)) {
		t.Fatalf("today: %s", rep.Today)
	}
	if len(rep.Due) != 1 || len(rep.Fired) != 1 || rep.Fired[0].Kind != KindD7 {
		t.Fatalf("report: due=%d fired=%v", len(rep.Due), rep.Fired)
	}
	sent := rec.Sent()
	if len(sent) != 1 || sent[0].Kind != KindD7 || sent[0].ID != "ca-1" {
		t.Fatalf("recording: %#v", sent)
	}
}

func TestRunListOnlyDoesNotNotify(t *testing.T) {
	item := Item{ID: "ca-1", Result: caJune1(t)}
	rec := &RecordingNotifier{}
	rep, err := Run(context.Background(), []Item{item}, Options{
		Today:    clock.NewDate(2026, time.June, 15),
		Within:   7,
		Fire:     false,
		Notifier: rec,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Due) != 1 || len(rep.Fired) != 0 || len(rec.Sent()) != 0 {
		t.Fatalf("list-only should not fire: %#v sent=%d", rep, len(rec.Sent()))
	}
}

func TestRunErrors(t *testing.T) {
	if _, err := Run(context.Background(), nil, Options{Within: -1}); !errors.Is(err, ErrNegativeWithin) {
		t.Fatalf("negative: %v", err)
	}
	if _, err := Run(context.Background(), nil, Options{Today: clock.NewDate(2026, 1, 1), Fire: true}); !errors.Is(err, ErrNilNotifier) {
		t.Fatalf("nil notifier: %v", err)
	}
}

func TestEmitContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Emit(ctx, &RecordingNotifier{}, []Reminder{{ID: "x", Kind: KindDue}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

func TestWriterNotifierJSONLine(t *testing.T) {
	var buf bytes.Buffer
	r := Reminder{
		ID:         "ca-1",
		Kind:       KindD7,
		FireOn:     clock.NewDate(2026, time.June, 15),
		DeadlineOn: clock.NewDate(2026, time.June, 22),
		Result:     caJune1(t),
	}
	if err := (WriterNotifier{W: &buf}).Notify(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	var ev Event
	if err := json.Unmarshal(buf.Bytes(), &ev); err != nil {
		t.Fatalf("json: %v body=%s", err, buf.String())
	}
	if ev.Kind != "D-7" || ev.ID != "ca-1" || ev.FireOn != "2026-06-15" || ev.DeadlineOn != "2026-06-22" || ev.State != "CA" {
		t.Fatalf("event: %#v", ev)
	}
}

func TestFileNotifierAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reminders.jsonl")
	n := FileNotifier{Path: path}
	r := Reminder{ID: "a", Kind: KindDue, FireOn: clock.NewDate(2026, 6, 22), DeadlineOn: clock.NewDate(2026, 6, 22), Result: caJune1(t)}
	if err := n.Notify(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	r.ID = "b"
	if err := n.Notify(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines: %q", body)
	}
}

func TestFileNotifierRequiresPath(t *testing.T) {
	err := (FileNotifier{}).Notify(context.Background(), Reminder{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWebhookNotifier(t *testing.T) {
	var got Event
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost || req.Header.Get("Content-Type") != "application/json" {
			t.Errorf("method/header: %s %s", req.Method, req.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(req.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	n := WebhookNotifier{URL: srv.URL, Client: srv.Client()}
	r := Reminder{
		ID: "ca-1", Kind: KindD3,
		FireOn: clock.NewDate(2026, 6, 19), DeadlineOn: clock.NewDate(2026, 6, 22),
		Result: caJune1(t),
	}
	if err := n.Notify(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if got.Kind != "D-3" || got.ID != "ca-1" {
		t.Fatalf("webhook body: %#v", got)
	}
}

func TestWebhookNotifierHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	err := (WebhookNotifier{URL: srv.URL, Client: srv.Client()}).Notify(context.Background(), Reminder{Kind: KindDue})
	if err == nil {
		t.Fatal("expected HTTP error")
	}
}

func TestNewNotifier(t *testing.T) {
	n, err := NewNotifier("stdout", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := n.(WriterNotifier); !ok {
		t.Fatalf("stdout: %T", n)
	}
	if _, err := NewNotifier("file", "", ""); err == nil {
		t.Fatal("file without path")
	}
	if _, err := NewNotifier("webhook", "", ""); err == nil {
		t.Fatal("webhook without url")
	}
	if _, err := NewNotifier("sms", "", ""); err == nil {
		t.Fatal("unknown kind")
	}
	n, err = NewNotifier("file", "/tmp/x.jsonl", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := n.(FileNotifier); !ok {
		t.Fatalf("file: %T", n)
	}
}

func TestLoadCasesArrayAndObject(t *testing.T) {
	arr := `[
	  {"id":"ca-1","state":"CA","vacated_on":"2026-06-01","lease_ends_on":"2026-06-01"}
	]`
	items, err := LoadCases(strings.NewReader(arr))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "ca-1" || !items[0].Result.DeadlineOn.Equal(clock.NewDate(2026, time.June, 22)) {
		t.Fatalf("array: %#v", items)
	}

	obj := `{"cases":[{"id":"ny-1","state":"NY","vacated_on":"2026-06-01","lease_ends_on":"2026-06-01"}]}`
	items, err = LoadCases(strings.NewReader(obj))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Result.CalendarDays != 14 {
		t.Fatalf("object: %#v", items)
	}
}

func TestLoadCasesEmpty(t *testing.T) {
	items, err := LoadCases(strings.NewReader("  "))
	if err != nil || items != nil {
		t.Fatalf("empty: %v %#v", err, items)
	}
}

func TestLoadCasesUnsupportedState(t *testing.T) {
	_, err := LoadCases(strings.NewReader(`[{"id":"tx","state":"TX","vacated_on":"2026-06-01"}]`))
	if !errors.Is(err, clock.ErrUnsupportedState) {
		t.Fatalf("got %v", err)
	}
}

func TestLoadCasesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cases.json")
	body := `[{"id":"fl-1","state":"FL","vacated_on":"2026-06-01","claim_against_deposit":true}]`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := LoadCasesFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Result.Branch != clock.BranchClaim || items[0].Result.CalendarDays != 30 {
		t.Fatalf("FL claim: %#v", items[0].Result)
	}
}

func TestFormatDue(t *testing.T) {
	var buf bytes.Buffer
	item := Item{ID: "ca-1", Result: caJune1(t)}
	FormatDue(&buf, clock.NewDate(2026, time.June, 15), []Item{item})
	got := buf.String()
	if !strings.Contains(got, "today=2026-06-15 due=1") || !strings.Contains(got, "ca-1\tCA\tdeadline=2026-06-22\tdays=7") {
		t.Fatalf("format: %q", got)
	}
}

func TestDaysUntilDeadline(t *testing.T) {
	res := caJune1(t)
	if d := DaysUntilDeadline(clock.NewDate(2026, time.June, 15), res); d != 7 {
		t.Fatalf("days=%d", d)
	}
	if d := DaysUntilDeadline(clock.NewDate(2026, time.June, 23), res); d != -1 {
		t.Fatalf("overdue days=%d", d)
	}
}

func TestTodayNilClock(t *testing.T) {
	d := Today(nil)
	if d.IsZero() {
		t.Fatal("Today(nil) should use time.Now")
	}
}

func TestSystemClock(t *testing.T) {
	if (SystemClock{}).Now().IsZero() {
		t.Fatal("system clock")
	}
}
