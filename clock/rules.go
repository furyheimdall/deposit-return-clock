package clock

// Anchor is the event that starts a statutory return clock.
type Anchor string

const (
	// AnchorVacate starts the clock on the day the tenant surrenders possession.
	AnchorVacate Anchor = "vacate"
	// AnchorLeaseTermination starts the clock on lease termination
	// (NJ: contractual term end, or earlier re-let on the early-exit branch).
	AnchorLeaseTermination Anchor = "lease_termination"
)

// Branch selects a state's alternative clock when the statute splits.
type Branch string

const (
	BranchNone      Branch = ""
	BranchNoClaim   Branch = "no_claim"   // FL: no deduction claim
	BranchClaim     Branch = "claim"      // FL: written claim notice
	BranchLeaseTerm Branch = "lease_term" // NJ: term end (no accelerating re-let)
	BranchRelet     Branch = "relet"      // NJ: early vacate, unit re-let
)

// Rule is one data-driven statutory clock. Logic matches these rows;
// do not add per-state day-count conditionals outside this table.
type Rule struct {
	State        State
	ID           string
	Statute      string
	Anchor       Anchor
	CalendarDays int
	// MatchClaim, when non-nil, selects this row only when Input.ClaimAgainstDeposit equals *MatchClaim.
	MatchClaim *bool
}

func boolPtr(v bool) *bool { return &v }

// rules is the locked MVP table (CA / NY / FL / NJ only).
// Day counts are the Epic #1 / README canon numbers.
var rules = []Rule{
	{
		State:        California,
		ID:           "CA-vacate-21",
		Statute:      "Cal. Civ. Code § 1950.5(g)",
		Anchor:       AnchorVacate,
		CalendarDays: 21,
	},
	{
		State:        NewYork,
		ID:           "NY-vacate-14",
		Statute:      "N.Y. Gen. Oblig. Law § 7-108",
		Anchor:       AnchorVacate,
		CalendarDays: 14,
	},
	{
		State:        Florida,
		ID:           "FL-noclaim-15",
		Statute:      "Fla. Stat. § 83.49(3)(a)",
		Anchor:       AnchorVacate,
		CalendarDays: 15,
		MatchClaim:   boolPtr(false),
	},
	{
		State:        Florida,
		ID:           "FL-claim-30",
		Statute:      "Fla. Stat. § 83.49(3)(a)",
		Anchor:       AnchorVacate,
		CalendarDays: 30,
		MatchClaim:   boolPtr(true),
	},
	{
		State:        NewJersey,
		ID:           "NJ-term-30",
		Statute:      "N.J.S.A. 46:8-21.1",
		Anchor:       AnchorLeaseTermination,
		CalendarDays: 30,
	},
}

func rulesFor(state State) []Rule {
	out := make([]Rule, 0, 2)
	for _, r := range rules {
		if r.State == state {
			out = append(out, r)
		}
	}
	return out
}

func matchRule(in Input) (Rule, error) {
	found := rulesFor(in.State)
	if len(found) == 0 {
		return Rule{}, &UnsupportedStateError{State: string(in.State)}
	}
	for _, r := range found {
		if r.MatchClaim != nil && in.ClaimAgainstDeposit != *r.MatchClaim {
			continue
		}
		return r, nil
	}
	return Rule{}, &UnsupportedStateError{State: string(in.State)}
}

// resolveAnchor returns the civil date the matched rule's day count is added to.
// Branch selection for NJ re-let is the only place law differs from a flat add-days.
func resolveAnchor(in Input, rule Rule) (anchor Date, branch Branch, err error) {
	switch rule.Anchor {
	case AnchorLeaseTermination:
		if in.LeaseEndsOn.IsZero() {
			return Date{}, BranchNone, ErrMissingLeaseEndsOn
		}
		if in.EarlyExit && !in.ReletOn.IsZero() && !in.ReletOn.After(in.LeaseEndsOn) {
			return in.ReletOn, BranchRelet, nil
		}
		// Normal end, or early vacate not yet re-let: contractual term.
		// If the tenant held over past term, use the later surrender date.
		if !in.EarlyExit && !in.VacatedOn.IsZero() && in.VacatedOn.After(in.LeaseEndsOn) {
			return in.VacatedOn, BranchLeaseTerm, nil
		}
		return in.LeaseEndsOn, BranchLeaseTerm, nil
	case AnchorVacate:
		if in.VacatedOn.IsZero() {
			return Date{}, BranchNone, ErrMissingVacatedOn
		}
		branch = BranchNone
		if rule.MatchClaim != nil {
			if *rule.MatchClaim {
				branch = BranchClaim
			} else {
				branch = BranchNoClaim
			}
		}
		return in.VacatedOn, branch, nil
	default:
		return Date{}, BranchNone, ErrMissingVacatedOn
	}
}
