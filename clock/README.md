# clock

Pure library: statutory **security-deposit return deadlines** for the locked MVP states **CA / NY / FL / NJ**.

> **Not legal advice.** Dig-sourced deadlines; operators must verify current statute.
> Day counts below are the Epic #1 / README **canon numbers**. This package implements those numbers; it is not a 50-state survey and it will not guess an unsupported state.

Reminders (D-7 / D-3 / due) live in [`notify/`](../notify/) and are **not** implemented here. `Result` exposes `DeadlineOn`, `State`, `EarlyExit`, and `Branch` so #5 can consume this API later.

## Public API

```go
res, err := clock.Compute(clock.Input{
    State:       clock.California,
    VacatedOn:   clock.NewDate(2026, time.June, 1),
    LeaseEndsOn: clock.NewDate(2026, time.June, 1),
})
// res.DeadlineOn == 2026-06-22  (vacate + 21)
// res.Deadline() is midnight UTC on that civil date
```

Unknown states return `clock.ErrUnsupportedState` (`*UnsupportedStateError`). There is no silent fallback.

## Statute table (canon)

| State | Clock (canon) | Anchor | Statute (cite) |
| --- | --- | --- | --- |
| **CA** | vacate + **21** days | possession surrendered | [Cal. Civ. Code § 1950.5(g)](https://leginfo.legislature.ca.gov/faces/codes_displaySection.xhtml?lawCode=CIV&sectionNum=1950.5) |
| **NY** | vacate + **14** days | possession surrendered | [N.Y. Gen. Oblig. Law § 7-108](https://www.nysenate.gov/legislation/laws/GOB/7-108) |
| **FL** | no-claim ≤**15** / claim notice ≤**30** | possession surrendered | [Fla. Stat. § 83.49(3)(a)](http://www.leg.state.fl.us/statutes/index.cfm?App_mode=Display_Statute&URL=0000-0099/0083/Sections/0083.49.html) |
| **NJ** | lease term + **30** (early vacate → re-let sensitive) | lease termination | [N.J.S.A. 46:8-21.1](https://law.justia.com/codes/new-jersey/title-46/section-46-8-21-1/) |

Rules are **data-driven** (`rules` in `rules.go`). Day counts are not hard-coded in `if state ==` branches.

## Early move-out / early-termination

| State | Does early exit change the clock? |
| --- | --- |
| **CA** | No. Clock is always vacate + 21. `EarlyExit` is recorded for downstream consumers. |
| **NY** | No. Clock is always vacate + 14. |
| **FL** | Day count is the **claim** branch (15 vs 30), not early exit. Early exit is recorded; both claim and no-claim still run from vacate. |
| **NJ** | **Yes, when the unit is re-let.** Premature vacate is a breach, not termination (`Mitchell v. First Real Estate Equities`, 293 N.J. Super. 547 (App. Div. 1996)). The 30-day clock runs from contractual term end until re-let; re-let on or before term end is treated as termination (`BranchRelet`, re-let + 30). Early exit **without** a re-let date stays on term + 30. Holdover past term uses the later vacate date. |

Set `Input.EarlyExit` explicitly. The library does not infer it from dates.

## Out of scope

- Other states (explicit error; no 50-state table)
- Landlord accounting ledger / PMS sync
- Reminders (`notify/`)
- Evidence timeline / return pack
