// Package clock computes statutory security-deposit return deadlines
// for the locked MVP states: California, New York, Florida, and New Jersey.
//
// This is a pure library. It does not send reminders (see notify/) and
// does not guess rules for other states.
//
// Not legal advice. Day counts are the Epic #1 / README canon numbers;
// operators must verify current statute. See clock/README.md.
package clock

import (
	"errors"
	"fmt"
	"time"
)

// Seat identifies this package in the module layout.
const Seat = "clock"

// State is a USPS two-letter code. Only CA, NY, FL, and NJ are implemented.
type State string

const (
	California State = "CA"
	NewYork    State = "NY"
	Florida    State = "FL"
	NewJersey  State = "NJ"
)

// SupportedStates is the locked MVP set, in canon order.
func SupportedStates() []State {
	return []State{California, NewYork, Florida, NewJersey}
}

// Input is the tenancy-end / possession facts needed to start a clock.
// Dates are civil calendar dates; time-of-day is ignored.
type Input struct {
	State State

	// VacatedOn is the date the tenant surrendered possession (moved out).
	// Required for CA, NY, and FL. Used on NJ when the tenant held over
	// past the contractual term.
	VacatedOn Date

	// LeaseEndsOn is the contractual lease-term end. Required for NJ.
	LeaseEndsOn Date

	// EarlyExit is true when the tenant left (or the tenancy ended) before
	// the scheduled term. Required to take New Jersey's re-let branch.
	EarlyExit bool

	// ClaimAgainstDeposit selects Florida's 30-day written-claim branch.
	// Ignored in other states.
	ClaimAgainstDeposit bool

	// ReletOn is the date the unit was re-let. Used only for New Jersey
	// when EarlyExit is true; that date is treated as lease termination
	// (Mitchell v. First Real Estate Equities, 293 N.J. Super. 547 (App. Div. 1996)).
	ReletOn Date
}

// Result is the computed statutory return deadline and the facts #5 reminders
// (and later packs) can hang off: deadline date, state, early-exit, branch.
type Result struct {
	State        State
	EarlyExit    bool
	Branch       Branch
	AnchorOn     Date
	DeadlineOn   Date
	CalendarDays int
	Rule         Rule
}

// Deadline is the statutory return-by date as midnight UTC. Convenience
// for notify/ and other consumers that prefer time.Time.
func (r Result) Deadline() time.Time {
	return r.DeadlineOn.UTC()
}

var (
	// ErrUnsupportedState is the sentinel for an unknown / out-of-MVP state.
	// Never invent a 50-state default.
	ErrUnsupportedState = errors.New("clock: unsupported state")

	// ErrMissingVacatedOn is returned when a vacate-anchored rule has no possession-end date.
	ErrMissingVacatedOn = errors.New("clock: vacated_on is required")

	// ErrMissingLeaseEndsOn is returned when a New Jersey computation has no lease-term end.
	ErrMissingLeaseEndsOn = errors.New("clock: lease_ends_on is required")
)

// UnsupportedStateError names the rejected state. errors.Is(err, ErrUnsupportedState) is true.
type UnsupportedStateError struct {
	State string
}

func (e *UnsupportedStateError) Error() string {
	return fmt.Sprintf("clock: unsupported state %q (supported: CA, NY, FL, NJ)", e.State)
}

func (e *UnsupportedStateError) Is(target error) bool {
	return target == ErrUnsupportedState
}

func (e *UnsupportedStateError) Unwrap() error {
	return ErrUnsupportedState
}

// Compute returns the statutory deposit-return deadline for in.
// Unknown states return an explicit error; they are never mapped to another state's rule.
func Compute(in Input) (Result, error) {
	rule, err := matchRule(in)
	if err != nil {
		return Result{}, err
	}
	anchor, branch, err := resolveAnchor(in, rule)
	if err != nil {
		return Result{}, err
	}
	return Result{
		State:        in.State,
		EarlyExit:    in.EarlyExit,
		Branch:       branch,
		AnchorOn:     anchor,
		DeadlineOn:   anchor.AddDays(rule.CalendarDays),
		CalendarDays: rule.CalendarDays,
		Rule:         rule,
	}, nil
}

// LookupRules returns the data-driven rule rows for state.
func LookupRules(state State) ([]Rule, error) {
	found := rulesFor(state)
	if len(found) == 0 {
		return nil, &UnsupportedStateError{State: string(state)}
	}
	out := make([]Rule, len(found))
	copy(out, found)
	return out, nil
}
