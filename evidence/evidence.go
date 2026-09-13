// Package evidence records a photo / receipt / memo evidence timeline.
//
// Events are append-only: once stored they are never updated or removed.
// Events() returns a chronological copy ordered by CapturedAt, then append
// sequence (stable for equal timestamps).
package evidence

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Seat identifies this package in the module layout.
const Seat = "evidence"

// Kind is the evidence event type.
type Kind string

const (
	KindPhoto      Kind = "photo"
	KindReceipt    Kind = "receipt"
	KindAttachment Kind = "attachment" // generic attachment
	KindMemo       Kind = "memo"
)

// Event is one timeline entry: an attachment (or memo) plus capture time and note.
type Event struct {
	ID         string
	Kind       Kind
	CapturedAt time.Time
	Note       string
	URI        string // photo / receipt / generic attachment locator
	seq        int    // append index; used as a chronological tiebreaker
}

var (
	ErrNilTimeline    = errors.New("evidence: nil timeline")
	ErrInvalidKind    = errors.New("evidence: invalid kind")
	ErrZeroCapturedAt = errors.New("evidence: captured_at is required")
	ErrEmptyEvent     = errors.New("evidence: note or uri is required")
	ErrDuplicateID    = errors.New("evidence: duplicate id")
)

// Valid reports whether k is a known evidence kind.
func (k Kind) Valid() bool {
	switch k {
	case KindPhoto, KindReceipt, KindAttachment, KindMemo:
		return true
	default:
		return false
	}
}

// Timeline is an append-only evidence log.
type Timeline struct {
	seq    int
	events []Event
	byID   map[string]int
}

// NewTimeline returns an empty append-only timeline.
func NewTimeline() *Timeline {
	return &Timeline{byID: make(map[string]int)}
}

// Len returns the number of stored events.
func (t *Timeline) Len() int {
	if t == nil {
		return 0
	}
	return len(t.events)
}

// Append stores e. If ID is empty, a sequential id (evt-N) is assigned.
// The stored event is returned. Callers cannot edit or delete prior events.
func (t *Timeline) Append(e Event) (Event, error) {
	if t == nil {
		return Event{}, ErrNilTimeline
	}
	if !e.Kind.Valid() {
		return Event{}, fmt.Errorf("%w: %q", ErrInvalidKind, e.Kind)
	}
	if e.CapturedAt.IsZero() {
		return Event{}, ErrZeroCapturedAt
	}
	if strings.TrimSpace(e.Note) == "" && strings.TrimSpace(e.URI) == "" {
		return Event{}, ErrEmptyEvent
	}
	t.seq++
	e.seq = t.seq
	if strings.TrimSpace(e.ID) == "" {
		e.ID = fmt.Sprintf("evt-%d", e.seq)
	}
	if t.byID == nil {
		t.byID = make(map[string]int)
	}
	if _, exists := t.byID[e.ID]; exists {
		return Event{}, fmt.Errorf("%w: %s", ErrDuplicateID, e.ID)
	}
	t.byID[e.ID] = len(t.events)
	t.events = append(t.events, e)
	return e, nil
}

// Events returns a chronological copy of the log (CapturedAt, then append seq).
func (t *Timeline) Events() []Event {
	if t == nil || len(t.events) == 0 {
		return []Event{}
	}
	out := make([]Event, len(t.events))
	copy(out, t.events)
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].CapturedAt.Equal(out[j].CapturedAt) {
			return out[i].CapturedAt.Before(out[j].CapturedAt)
		}
		return out[i].seq < out[j].seq
	})
	return out
}

// ByID looks up a stored event by id.
func (t *Timeline) ByID(id string) (Event, bool) {
	if t == nil || t.byID == nil {
		return Event{}, false
	}
	i, ok := t.byID[id]
	if !ok {
		return Event{}, false
	}
	return t.events[i], true
}
