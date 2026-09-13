// Package pack builds an itemized security-deposit return pack.
//
// Deduction lines plus remaining balance export as a checklist (markdown/JSON)
// and a simple PDF. Statutory due dates are accepted as caller-supplied mocks
// until clock/#4 lands — this package does not implement CA/NY/FL/NJ rules.
package pack

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/furyheimdall/deposit-return-clock/clock"
	"github.com/furyheimdall/deposit-return-clock/evidence"
)

// Seat identifies this package in the module layout.
const Seat = "pack"

// Cents is a USD amount in cents (avoids float money).
type Cents int64

// String formats c as $D.CC (ASCII).
func (c Cents) String() string {
	sign := ""
	v := int64(c)
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s$%d.%02d", sign, v/100, v%100)
}

// Line is one itemized deduction, optionally linked to evidence event IDs.
type Line struct {
	Description string
	Amount      Cents
	EvidenceIDs []string
}

// Input is the data needed to assemble a return pack.
type Input struct {
	Tenant   string
	Property string
	Deposit  Cents
	Lines    []Line
	// DueBy is the statutory return deadline. Mock dates are OK until
	// the clock package (#4) computes CA/NY/FL/NJ deadlines.
	DueBy time.Time
	// Timeline is optional. When set, every Line.EvidenceIDs value must exist.
	Timeline *evidence.Timeline
}

// Pack is a computed itemized return statement.
type Pack struct {
	Tenant          string
	Property        string
	Deposit         Cents
	Lines           []Line
	TotalDeductions Cents
	Remaining       Cents // max(0, Deposit - TotalDeductions)
	DueBy           time.Time
	DueByNote       string
	Evidence        []evidence.Event
}

var (
	ErrNegativeDeposit = errors.New("pack: deposit must be >= 0")
	ErrNegativeLine    = errors.New("pack: deduction amount must be >= 0")
	ErrEmptyLine       = errors.New("pack: deduction description is required")
	ErrUnknownEvidence = errors.New("pack: unknown evidence id")
)

// Build computes totals and snapshots the evidence timeline.
func Build(in Input) (Pack, error) {
	if in.Deposit < 0 {
		return Pack{}, ErrNegativeDeposit
	}
	lines := make([]Line, 0, len(in.Lines))
	var total Cents
	for i, line := range in.Lines {
		if line.Amount < 0 {
			return Pack{}, fmt.Errorf("%w: line %d", ErrNegativeLine, i)
		}
		if strings.TrimSpace(line.Description) == "" {
			return Pack{}, fmt.Errorf("%w: line %d", ErrEmptyLine, i)
		}
		ids := append([]string(nil), line.EvidenceIDs...)
		if in.Timeline != nil {
			for _, id := range ids {
				if _, ok := in.Timeline.ByID(id); !ok {
					return Pack{}, fmt.Errorf("%w: %s (line %d)", ErrUnknownEvidence, id, i)
				}
			}
		}
		line.EvidenceIDs = ids
		lines = append(lines, line)
		total += line.Amount
	}
	remaining := in.Deposit - total
	if remaining < 0 {
		remaining = 0
	}
	var ev []evidence.Event
	if in.Timeline != nil {
		ev = in.Timeline.Events()
	}
	dueNote := "mock deadline (" + clock.Seat + "/#4 pending)"
	if in.DueBy.IsZero() {
		dueNote = "deadline unset; supply DueBy or wait for clock/#4"
	}
	return Pack{
		Tenant:          in.Tenant,
		Property:        in.Property,
		Deposit:         in.Deposit,
		Lines:           lines,
		TotalDeductions: total,
		Remaining:       remaining,
		DueBy:           in.DueBy,
		DueByNote:       dueNote,
		Evidence:        ev,
	}, nil
}
