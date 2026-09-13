package evidence

import (
	"errors"
	"testing"
	"time"
)

func TestSeat(t *testing.T) {
	if Seat != "evidence" {
		t.Fatalf("evidence seat: got %q", Seat)
	}
}

func TestTimelineOrderByCapturedAt(t *testing.T) {
	tl := NewTimeline()
	late := time.Date(2026, 9, 12, 15, 0, 0, 0, time.UTC)
	early := time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	mid := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)

	// Append out of chronological order to prove Events() reorders.
	if _, err := tl.Append(Event{Kind: KindReceipt, CapturedAt: late, Note: "paint receipt", URI: "receipts/paint.pdf"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tl.Append(Event{Kind: KindPhoto, CapturedAt: early, Note: "move-in wall", URI: "photos/wall.jpg"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tl.Append(Event{Kind: KindMemo, CapturedAt: mid, Note: "tenant acknowledged stain"}); err != nil {
		t.Fatal(err)
	}

	got := tl.Events()
	if len(got) != 3 {
		t.Fatalf("len=%d", len(got))
	}
	wantKinds := []Kind{KindPhoto, KindMemo, KindReceipt}
	for i, k := range wantKinds {
		if got[i].Kind != k {
			t.Errorf("events[%d].Kind=%q want %q", i, got[i].Kind, k)
		}
	}
	if !got[0].CapturedAt.Equal(early) || !got[2].CapturedAt.Equal(late) {
		t.Fatalf("not chronological: %#v", got)
	}
}

func TestTimelineEqualCapturedAtUsesAppendSeq(t *testing.T) {
	tl := NewTimeline()
	ts := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	if _, err := tl.Append(Event{Kind: KindPhoto, CapturedAt: ts, URI: "a.jpg", Note: "first"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tl.Append(Event{Kind: KindAttachment, CapturedAt: ts, URI: "b.pdf", Note: "second"}); err != nil {
		t.Fatal(err)
	}
	got := tl.Events()
	if got[0].Note != "first" || got[1].Note != "second" {
		t.Fatalf("tie-break order: %q then %q", got[0].Note, got[1].Note)
	}
}

func TestTimelineAppendOnly(t *testing.T) {
	tl := NewTimeline()
	ts := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	stored, err := tl.Append(Event{Kind: KindPhoto, CapturedAt: ts, URI: "x.jpg", Note: "original"})
	if err != nil {
		t.Fatal(err)
	}
	got := tl.Events()
	got[0].Note = "mutated by caller"
	again, ok := tl.ByID(stored.ID)
	if !ok {
		t.Fatal("missing id")
	}
	if again.Note != "original" {
		t.Fatalf("store mutated: %q", again.Note)
	}
	if tl.Len() != 1 {
		t.Fatalf("len=%d (delete is not allowed)", tl.Len())
	}
}

func TestAppendValidation(t *testing.T) {
	tl := NewTimeline()
	ts := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		e    Event
		want error
	}{
		{name: "bad kind", e: Event{Kind: "video", CapturedAt: ts, Note: "x"}, want: ErrInvalidKind},
		{name: "zero time", e: Event{Kind: KindPhoto, Note: "x", URI: "a.jpg"}, want: ErrZeroCapturedAt},
		{name: "empty body", e: Event{Kind: KindMemo, CapturedAt: ts}, want: ErrEmptyEvent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tl.Append(tc.e)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
		})
	}

	if _, err := tl.Append(Event{ID: "fixed", Kind: KindReceipt, CapturedAt: ts, URI: "r.pdf", Note: "ok"}); err != nil {
		t.Fatal(err)
	}
	_, err := tl.Append(Event{ID: "fixed", Kind: KindReceipt, CapturedAt: ts, URI: "r2.pdf", Note: "dup"})
	if !errors.Is(err, ErrDuplicateID) {
		t.Fatalf("dup: %v", err)
	}
}

func TestAppendAssignsID(t *testing.T) {
	tl := NewTimeline()
	ts := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	got, err := tl.Append(Event{Kind: KindAttachment, CapturedAt: ts, URI: "scan.bin"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "evt-1" {
		t.Fatalf("id=%q", got.ID)
	}
}
