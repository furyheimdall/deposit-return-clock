package clock

import (
	"errors"
	"testing"
	"time"
)

func TestSeat(t *testing.T) {
	if Seat != "clock" {
		t.Fatalf("clock seat: got %q", Seat)
	}
}

func TestSupportedStatesLockedMVP(t *testing.T) {
	got := SupportedStates()
	want := []State{California, NewYork, Florida, NewJersey}
	if len(got) != len(want) {
		t.Fatalf("SupportedStates: got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SupportedStates[%d]: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestComputeUnknownStateExplicitError(t *testing.T) {
	vacate := NewDate(2026, time.June, 1)
	cases := []State{"", "TX", "WA", "ca", "California", "US-CA", "NYC"}
	for _, st := range cases {
		_, err := Compute(Input{State: st, VacatedOn: vacate, LeaseEndsOn: vacate})
		if err == nil {
			t.Fatalf("state %q: expected error, got nil", st)
		}
		if !errors.Is(err, ErrUnsupportedState) {
			t.Fatalf("state %q: errors.Is(ErrUnsupportedState)=false; err=%v", st, err)
		}
		var u *UnsupportedStateError
		if !errors.As(err, &u) || u.State != string(st) {
			t.Fatalf("state %q: want UnsupportedStateError naming it; err=%v", st, err)
		}
	}
}

func TestLookupRulesUnknownState(t *testing.T) {
	_, err := LookupRules("TX")
	if !errors.Is(err, ErrUnsupportedState) {
		t.Fatalf("LookupRules TX: %v", err)
	}
}

func TestComputeMissingDates(t *testing.T) {
	t.Run("CA missing vacate", func(t *testing.T) {
		_, err := Compute(Input{State: California})
		if !errors.Is(err, ErrMissingVacatedOn) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("NJ missing lease end", func(t *testing.T) {
		_, err := Compute(Input{
			State:     NewJersey,
			VacatedOn: NewDate(2026, time.March, 1),
			EarlyExit: true,
		})
		if !errors.Is(err, ErrMissingLeaseEndsOn) {
			t.Fatalf("got %v", err)
		}
	})
}

func TestComputeCanonDeadlines(t *testing.T) {
	vacate := NewDate(2026, time.June, 1)
	term := NewDate(2026, time.June, 30)

	type want struct {
		days     int
		deadline Date
		anchor   Date
		branch   Branch
		statute  string
		ruleID   string
		early    bool
	}

	cases := []struct {
		name string
		in   Input
		want want
	}{
		{
			name: "CA normal end",
			in: Input{
				State:       California,
				VacatedOn:   vacate,
				LeaseEndsOn: vacate,
			},
			want: want{
				days: 21, deadline: NewDate(2026, time.June, 22), anchor: vacate,
				statute: "Cal. Civ. Code § 1950.5(g)", ruleID: "CA-vacate-21",
			},
		},
		{
			name: "CA early exit still vacate+21",
			in: Input{
				State:       California,
				VacatedOn:   NewDate(2026, time.May, 10),
				LeaseEndsOn: term,
				EarlyExit:   true,
			},
			want: want{
				days: 21, deadline: NewDate(2026, time.May, 31),
				anchor: NewDate(2026, time.May, 10), early: true,
				statute: "Cal. Civ. Code § 1950.5(g)", ruleID: "CA-vacate-21",
			},
		},
		{
			name: "NY normal end",
			in: Input{
				State:       NewYork,
				VacatedOn:   vacate,
				LeaseEndsOn: vacate,
			},
			want: want{
				days: 14, deadline: NewDate(2026, time.June, 15), anchor: vacate,
				statute: "N.Y. Gen. Oblig. Law § 7-108", ruleID: "NY-vacate-14",
			},
		},
		{
			name: "NY early exit still vacate+14",
			in: Input{
				State:       NewYork,
				VacatedOn:   NewDate(2026, time.May, 10),
				LeaseEndsOn: term,
				EarlyExit:   true,
			},
			want: want{
				days: 14, deadline: NewDate(2026, time.May, 24),
				anchor: NewDate(2026, time.May, 10), early: true,
				statute: "N.Y. Gen. Oblig. Law § 7-108", ruleID: "NY-vacate-14",
			},
		},
		{
			name: "FL normal no-claim",
			in: Input{
				State:               Florida,
				VacatedOn:           vacate,
				LeaseEndsOn:         vacate,
				ClaimAgainstDeposit: false,
			},
			want: want{
				days: 15, deadline: NewDate(2026, time.June, 16), anchor: vacate,
				branch: BranchNoClaim, statute: "Fla. Stat. § 83.49(3)(a)", ruleID: "FL-noclaim-15",
			},
		},
		{
			name: "FL normal claim notice",
			in: Input{
				State:               Florida,
				VacatedOn:           vacate,
				LeaseEndsOn:         vacate,
				ClaimAgainstDeposit: true,
			},
			want: want{
				days: 30, deadline: NewDate(2026, time.July, 1), anchor: vacate,
				branch: BranchClaim, statute: "Fla. Stat. § 83.49(3)(a)", ruleID: "FL-claim-30",
			},
		},
		{
			name: "FL early exit no-claim still vacate+15",
			in: Input{
				State:               Florida,
				VacatedOn:           NewDate(2026, time.May, 10),
				LeaseEndsOn:         term,
				EarlyExit:           true,
				ClaimAgainstDeposit: false,
			},
			want: want{
				days: 15, deadline: NewDate(2026, time.May, 25),
				anchor: NewDate(2026, time.May, 10), early: true, branch: BranchNoClaim,
				statute: "Fla. Stat. § 83.49(3)(a)", ruleID: "FL-noclaim-15",
			},
		},
		{
			name: "FL early exit claim still vacate+30",
			in: Input{
				State:               Florida,
				VacatedOn:           NewDate(2026, time.May, 10),
				LeaseEndsOn:         term,
				EarlyExit:           true,
				ClaimAgainstDeposit: true,
			},
			want: want{
				days: 30, deadline: NewDate(2026, time.June, 9),
				anchor: NewDate(2026, time.May, 10), early: true, branch: BranchClaim,
				statute: "Fla. Stat. § 83.49(3)(a)", ruleID: "FL-claim-30",
			},
		},
		{
			name: "NJ normal term+30",
			in: Input{
				State:       NewJersey,
				VacatedOn:   term,
				LeaseEndsOn: term,
			},
			want: want{
				days: 30, deadline: NewDate(2026, time.July, 30), anchor: term,
				branch: BranchLeaseTerm, statute: "N.J.S.A. 46:8-21.1", ruleID: "NJ-term-30",
			},
		},
		{
			name: "NJ early exit without re-let stays on term+30",
			in: Input{
				State:       NewJersey,
				VacatedOn:   NewDate(2026, time.May, 10),
				LeaseEndsOn: term,
				EarlyExit:   true,
			},
			want: want{
				days: 30, deadline: NewDate(2026, time.July, 30), anchor: term,
				early: true, branch: BranchLeaseTerm,
				statute: "N.J.S.A. 46:8-21.1", ruleID: "NJ-term-30",
			},
		},
		{
			name: "NJ early exit with re-let uses re-let+30",
			in: Input{
				State:       NewJersey,
				VacatedOn:   NewDate(2026, time.May, 10),
				LeaseEndsOn: term,
				EarlyExit:   true,
				ReletOn:     NewDate(2026, time.May, 20),
			},
			want: want{
				days: 30, deadline: NewDate(2026, time.June, 19),
				anchor: NewDate(2026, time.May, 20), early: true, branch: BranchRelet,
				statute: "N.J.S.A. 46:8-21.1", ruleID: "NJ-term-30",
			},
		},
		{
			name: "NJ holdover uses later vacate as termination",
			in: Input{
				State:       NewJersey,
				VacatedOn:   NewDate(2026, time.July, 5),
				LeaseEndsOn: term,
			},
			want: want{
				days: 30, deadline: NewDate(2026, time.August, 4),
				anchor: NewDate(2026, time.July, 5), branch: BranchLeaseTerm,
				statute: "N.J.S.A. 46:8-21.1", ruleID: "NJ-term-30",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Compute(tc.in)
			if err != nil {
				t.Fatalf("Compute: %v", err)
			}
			if got.State != tc.in.State {
				t.Errorf("State: got %q", got.State)
			}
			if got.EarlyExit != tc.want.early {
				t.Errorf("EarlyExit: got %v want %v", got.EarlyExit, tc.want.early)
			}
			if got.Branch != tc.want.branch {
				t.Errorf("Branch: got %q want %q", got.Branch, tc.want.branch)
			}
			if got.CalendarDays != tc.want.days {
				t.Errorf("CalendarDays: got %d want %d", got.CalendarDays, tc.want.days)
			}
			if !got.AnchorOn.Equal(tc.want.anchor) {
				t.Errorf("AnchorOn: got %s want %s", got.AnchorOn, tc.want.anchor)
			}
			if !got.DeadlineOn.Equal(tc.want.deadline) {
				t.Errorf("DeadlineOn: got %s want %s", got.DeadlineOn, tc.want.deadline)
			}
			if got.Rule.Statute != tc.want.statute {
				t.Errorf("Rule.Statute: got %q want %q", got.Rule.Statute, tc.want.statute)
			}
			if got.Rule.ID != tc.want.ruleID {
				t.Errorf("Rule.ID: got %q want %q", got.Rule.ID, tc.want.ruleID)
			}
			if !got.Deadline().Equal(tc.want.deadline.UTC()) {
				t.Errorf("Deadline() time: got %s", got.Deadline())
			}
		})
	}
}

func TestNJReletAfterTermDoesNotAccelerate(t *testing.T) {
	term := NewDate(2026, time.June, 30)
	got, err := Compute(Input{
		State:       NewJersey,
		VacatedOn:   NewDate(2026, time.May, 10),
		LeaseEndsOn: term,
		EarlyExit:   true,
		ReletOn:     NewDate(2026, time.July, 15),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Branch != BranchLeaseTerm || !got.AnchorOn.Equal(term) {
		t.Fatalf("re-let after term should stay on term; branch=%s anchor=%s", got.Branch, got.AnchorOn)
	}
}

func TestFLClaimFlagDoesNotAffectOtherStates(t *testing.T) {
	vacate := NewDate(2026, time.June, 1)
	got, err := Compute(Input{
		State:               California,
		VacatedOn:           vacate,
		ClaimAgainstDeposit: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.CalendarDays != 21 || got.Branch != BranchNone {
		t.Fatalf("CA+claim: days=%d branch=%q", got.CalendarDays, got.Branch)
	}
}

func TestDateAddDaysAndUTC(t *testing.T) {
	d := NewDate(2026, time.January, 31)
	if got := d.AddDays(1); !got.Equal(NewDate(2026, time.February, 1)) {
		t.Fatalf("AddDays: %s", got)
	}
	if d.String() != "2026-01-31" {
		t.Fatalf("String: %q", d.String())
	}
	if !(Date{}).IsZero() || (Date{}).String() != "" {
		t.Fatal("zero Date")
	}
	if DateFromTime(d.UTC().Add(15 * time.Hour)).After(d) {
		t.Fatal("DateFromTime should stay on the UTC civil date")
	}
}

func TestLookupRulesDataDrivenDays(t *testing.T) {
	wantDays := map[State][]int{
		California: {21},
		NewYork:    {14},
		Florida:    {15, 30},
		NewJersey:  {30},
	}
	for st, days := range wantDays {
		rs, err := LookupRules(st)
		if err != nil {
			t.Fatalf("%s: %v", st, err)
		}
		if len(rs) != len(days) {
			t.Fatalf("%s: %d rules", st, len(rs))
		}
		for i, d := range days {
			if rs[i].CalendarDays != d {
				t.Fatalf("%s rule %d: days %d want %d", st, i, rs[i].CalendarDays, d)
			}
		}
	}
}
