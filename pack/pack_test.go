package pack

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/furyheimdall/deposit-return-clock/clock"
	"github.com/furyheimdall/deposit-return-clock/evidence"
)

func TestSeat(t *testing.T) {
	if Seat != "pack" {
		t.Fatalf("pack seat: got %q", Seat)
	}
	if clock.Seat != "clock" {
		t.Fatalf("clock seat missing: %q", clock.Seat)
	}
}

func sampleTimeline(t *testing.T) *evidence.Timeline {
	t.Helper()
	tl := evidence.NewTimeline()
	ts1 := time.Date(2026, 8, 15, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 9, 1, 14, 30, 0, 0, time.UTC)
	if _, err := tl.Append(evidence.Event{ID: "photo-1", Kind: evidence.KindPhoto, CapturedAt: ts2, URI: "photos/carpet.jpg", Note: "stain at move-out"}); err != nil {
		t.Fatal(err)
	}
	if _, err := tl.Append(evidence.Event{ID: "rcpt-1", Kind: evidence.KindReceipt, CapturedAt: ts1, URI: "receipts/clean.pdf", Note: "vendor invoice"}); err != nil {
		t.Fatal(err)
	}
	return tl
}

func samplePack(t *testing.T) Pack {
	t.Helper()
	tl := sampleTimeline(t)
	// Mock CA vacate+21 until clock/#4.
	due := time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC)
	p, err := Build(Input{
		Tenant:   "Ada Tenant",
		Property: "12 Oak St #4",
		Deposit:  200000, // $2,000.00
		DueBy:    due,
		Timeline: tl,
		Lines: []Line{
			{Description: "Carpet cleaning", Amount: 15000, EvidenceIDs: []string{"photo-1", "rcpt-1"}},
			{Description: "Broken blind", Amount: 4000, EvidenceIDs: []string{"photo-1"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRemainingBalance(t *testing.T) {
	p := samplePack(t)
	if p.TotalDeductions != 19000 {
		t.Fatalf("deductions=%d", p.TotalDeductions)
	}
	if p.Remaining != 181000 { // 2000.00 - 190.00
		t.Fatalf("remaining=%d want 181000", p.Remaining)
	}
	if p.Remaining.String() != "$1810.00" {
		t.Fatalf("format=%q", p.Remaining)
	}
}

func TestRemainingClampsAtZero(t *testing.T) {
	p, err := Build(Input{
		Deposit: 5000,
		Lines:   []Line{{Description: "Repaint", Amount: 9000}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Remaining != 0 {
		t.Fatalf("remaining=%d", p.Remaining)
	}
	if p.TotalDeductions != 9000 {
		t.Fatalf("deductions=%d", p.TotalDeductions)
	}
}

func TestBuildRejectsUnknownEvidence(t *testing.T) {
	tl := sampleTimeline(t)
	_, err := Build(Input{
		Deposit:  1000,
		Timeline: tl,
		Lines:    []Line{{Description: "Mystery", Amount: 100, EvidenceIDs: []string{"nope"}}},
	})
	if !errors.Is(err, ErrUnknownEvidence) {
		t.Fatalf("err=%v", err)
	}
}

func TestBuildValidation(t *testing.T) {
	if _, err := Build(Input{Deposit: -1}); !errors.Is(err, ErrNegativeDeposit) {
		t.Fatalf("deposit: %v", err)
	}
	if _, err := Build(Input{Deposit: 1, Lines: []Line{{Description: "x", Amount: -2}}}); !errors.Is(err, ErrNegativeLine) {
		t.Fatalf("line: %v", err)
	}
	if _, err := Build(Input{Deposit: 1, Lines: []Line{{Amount: 1}}}); !errors.Is(err, ErrEmptyLine) {
		t.Fatalf("desc: %v", err)
	}
}

func TestChecklistMarkdownSmoke(t *testing.T) {
	md := samplePack(t).ChecklistMarkdown()
	for _, needle := range []string{
		"# Itemized security deposit return",
		"Carpet cleaning",
		"$150.00",
		"**Remaining balance:** $1810.00",
		"`rcpt-1`", // chronological: receipt captured earlier
		"`photo-1`",
		"- [x] Remaining balance computed toward return",
		"Not legal advice",
	} {
		if !strings.Contains(md, needle) {
			t.Fatalf("markdown missing %q\n%s", needle, md)
		}
	}
	// Timeline order: receipt (Aug 15) before photo (Sep 1).
	if i, j := strings.Index(md, "`rcpt-1`"), strings.Index(md, "`photo-1`"); i < 0 || j < 0 || i > j {
		t.Fatalf("timeline order in markdown: rcpt=%d photo=%d", i, j)
	}
}

func TestChecklistJSONSmoke(t *testing.T) {
	raw, err := samplePack(t).ChecklistJSON()
	if err != nil {
		t.Fatal(err)
	}
	var doc ChecklistDoc
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.RemainingCents != 181000 {
		t.Fatalf("remaining_cents=%d", doc.RemainingCents)
	}
	if len(doc.Lines) != 2 || len(doc.Evidence) != 2 {
		t.Fatalf("lines=%d evidence=%d", len(doc.Lines), len(doc.Evidence))
	}
	if doc.Evidence[0].ID != "rcpt-1" || doc.Evidence[1].ID != "photo-1" {
		t.Fatalf("json timeline order: %+v", doc.Evidence)
	}
	if len(doc.Checklist) == 0 {
		t.Fatal("empty checklist")
	}
}

func TestPDFSmoke(t *testing.T) {
	raw, err := samplePack(t).PDF()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(raw, []byte("%PDF-")) {
		t.Fatalf("missing PDF header: %q", raw[:min(20, len(raw))])
	}
	if !bytes.Contains(raw, []byte("%%EOF")) {
		t.Fatal("missing PDF EOF marker")
	}
	for _, needle := range []string{
		"ITEMIZED SECURITY DEPOSIT RETURN",
		"Carpet cleaning",
		"Remaining balance: $1810.00",
		"rcpt-1",
		"photo-1",
	} {
		if !bytes.Contains(raw, []byte(needle)) {
			t.Fatalf("pdf missing %q (len=%d)", needle, len(raw))
		}
	}
	if len(raw) < 400 {
		t.Fatalf("pdf too small: %d", len(raw))
	}
}
