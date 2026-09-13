// Package notify derives D-7 / D-3 / due reminders from clock.Result
// and emits them through a pluggable Notifier.
//
// This is the #5 seat. It does not change possession-clock rules
// (see clock/) and does not implement evidence or return packs.
//
// Not legal advice. Deadlines come from clock.Compute; operators must
// verify current statute.
package notify

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/furyheimdall/deposit-return-clock/clock"
)

// Seat identifies this package in the module layout.
const Seat = "notify"

// Kind is a reminder milestone relative to the possession-clock deadline.
type Kind string

const (
	// KindD7 fires 7 calendar days before DeadlineOn.
	KindD7 Kind = "D-7"
	// KindD3 fires 3 calendar days before DeadlineOn.
	KindD3 Kind = "D-3"
	// KindDue fires on DeadlineOn.
	KindDue Kind = "due"
)

// CanonKinds is the locked MVP reminder set, in fire order.
var CanonKinds = []Kind{KindD7, KindD3, KindDue}

// DaysBeforeDeadline is the canon offset for k. Unknown kinds return -1.
func (k Kind) DaysBeforeDeadline() int {
	switch k {
	case KindD7:
		return 7
	case KindD3:
		return 3
	case KindDue:
		return 0
	default:
		return -1
	}
}

// Item is one tenancy the reminder job can scan. Result must come from
// clock.Compute (or an equivalent public clock API).
type Item struct {
	ID     string
	Result clock.Result
}

// NewItem computes the possession clock for in and wraps it for reminders.
func NewItem(id string, in clock.Input) (Item, error) {
	res, err := clock.Compute(in)
	if err != nil {
		return Item{}, err
	}
	return Item{ID: id, Result: res}, nil
}

// Reminder is one scheduled emission derived from a clock.Result deadline.
type Reminder struct {
	ID         string
	Kind       Kind
	FireOn     clock.Date
	DeadlineOn clock.Date
	Result     clock.Result
}

// Event is the stable JSON payload notifiers send. Smallest wire for E3.
type Event struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	FireOn       string `json:"fire_on"`
	DeadlineOn   string `json:"deadline_on"`
	State        string `json:"state"`
	EarlyExit    bool   `json:"early_exit"`
	Branch       string `json:"branch"`
	RuleID       string `json:"rule_id"`
	CalendarDays int    `json:"calendar_days"`
}

// Event returns the notifier payload for r.
func (r Reminder) Event() Event {
	return Event{
		ID:           r.ID,
		Kind:         string(r.Kind),
		FireOn:       r.FireOn.String(),
		DeadlineOn:   r.DeadlineOn.String(),
		State:        string(r.Result.State),
		EarlyExit:    r.Result.EarlyExit,
		Branch:       string(r.Result.Branch),
		RuleID:       r.Result.Rule.ID,
		CalendarDays: r.Result.CalendarDays,
	}
}

// Schedule derives D-7, D-3, and due reminders from item.Result.DeadlineOn.
// It does not consult "today"; callers filter with RemindersOn or DueWithin.
func Schedule(item Item) []Reminder {
	deadline := item.Result.DeadlineOn
	if deadline.IsZero() {
		return nil
	}
	out := make([]Reminder, 0, len(CanonKinds))
	for _, k := range CanonKinds {
		days := k.DaysBeforeDeadline()
		if days < 0 {
			continue
		}
		out = append(out, Reminder{
			ID:         item.ID,
			Kind:       k,
			FireOn:     deadline.AddDays(-days),
			DeadlineOn: deadline,
			Result:     item.Result,
		})
	}
	return out
}

// DueWithin returns items whose deadline falls on today through today+n
// (inclusive). n is calendar days. Overdue items (deadline before today)
// are omitted. n < 0 yields nil.
func DueWithin(today clock.Date, n int, items []Item) []Item {
	if n < 0 || today.IsZero() {
		return nil
	}
	end := today.AddDays(n)
	var out []Item
	for _, it := range items {
		d := it.Result.DeadlineOn
		if d.IsZero() {
			continue
		}
		if !d.Before(today) && !d.After(end) {
			out = append(out, it)
		}
	}
	return out
}

// RemindersOn returns canon reminders whose FireOn equals today.
func RemindersOn(today clock.Date, items []Item) []Reminder {
	if today.IsZero() {
		return nil
	}
	var out []Reminder
	for _, it := range items {
		for _, r := range Schedule(it) {
			if r.FireOn.Equal(today) {
				out = append(out, r)
			}
		}
	}
	return out
}

// DaysUntilDeadline is calendar days from today to res.DeadlineOn.
// Negative means overdue.
func DaysUntilDeadline(today clock.Date, res clock.Result) int {
	if today.IsZero() || res.DeadlineOn.IsZero() {
		return 0
	}
	return int(res.DeadlineOn.UTC().Sub(today.UTC()) / (24 * time.Hour))
}

// Clock is a source of "now" so jobs and tests can inject a fake clock.
type Clock interface {
	Now() time.Time
}

// SystemClock uses time.Now.
type SystemClock struct{}

// Now returns the current time.
func (SystemClock) Now() time.Time { return time.Now() }

// FixedClock is a test/fake clock pinned to Instant.
type FixedClock struct {
	Instant time.Time
}

// Now returns Instant.
func (c FixedClock) Now() time.Time { return c.Instant }

// Today returns the UTC civil date of c. A nil Clock uses time.Now.
func Today(c Clock) clock.Date {
	if c == nil {
		return clock.DateFromTime(time.Now())
	}
	return clock.DateFromTime(c.Now())
}

var (
	// ErrNilNotifier is returned when Fire/Emit is asked to send without a Notifier.
	ErrNilNotifier = errors.New("notify: notifier is required")
	// ErrNegativeWithin is returned when a job window is negative.
	ErrNegativeWithin = errors.New("notify: within must be >= 0")
)

// Options configures a due-within list + optional fire pass.
type Options struct {
	// Clock supplies "today" when Today is zero.
	Clock Clock
	// Today overrides Clock when set.
	Today clock.Date
	// Within is the due-within-N window in calendar days (inclusive).
	Within int
	// Fire emits reminders whose FireOn is today.
	Fire bool
	// Notifier receives fired reminders. Required when Fire is true.
	Notifier Notifier
}

func (o Options) today() clock.Date {
	if !o.Today.IsZero() {
		return o.Today
	}
	return Today(o.Clock)
}

// Report is the result of a due-within list + optional fire.
type Report struct {
	Today clock.Date
	Due   []Item
	Fired []Reminder
}

// Run lists items due within opt.Within days and, when opt.Fire is set,
// emits today's D-7 / D-3 / due reminders through opt.Notifier.
func Run(ctx context.Context, items []Item, opt Options) (Report, error) {
	if opt.Within < 0 {
		return Report{}, ErrNegativeWithin
	}
	today := opt.today()
	rep := Report{
		Today: today,
		Due:   DueWithin(today, opt.Within, items),
	}
	if !opt.Fire {
		return rep, nil
	}
	if opt.Notifier == nil {
		return Report{}, ErrNilNotifier
	}
	fired := RemindersOn(today, items)
	if err := Emit(ctx, opt.Notifier, fired); err != nil {
		return Report{}, err
	}
	rep.Fired = fired
	return rep, nil
}

// Emit sends each reminder through n, in order. Stops on the first error.
func Emit(ctx context.Context, n Notifier, rs []Reminder) error {
	if n == nil {
		return ErrNilNotifier
	}
	for _, r := range rs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := n.Notify(ctx, r); err != nil {
			return fmt.Errorf("notify: %s %s: %w", r.Kind, r.ID, err)
		}
	}
	return nil
}
