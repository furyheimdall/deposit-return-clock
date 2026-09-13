# notify

Deadline reminders **D-7 / D-3 / due**, derived from [`clock.Result`](../clock/) (possession-clock deadline). Pluggable notifier: **stdout / file / webhook stub**.

> **Not legal advice.** Day counts and deadlines come from `clock.Compute`. Operators must verify current statute.

This package does **not** change clock rules, expand past CA/NY/FL/NJ, or implement evidence/pack (#3). No Push/SMS SaaS, CRM, or MCP.

## Public API (E3 wire)

```go
item, err := notify.NewItem("unit-1", clock.Input{
    State:     clock.California,
    VacatedOn: clock.NewDate(2026, time.June, 1),
})
// item.Result.DeadlineOn == 2026-06-22  (from clock)

reminders := notify.Schedule(item)
// D-7 → 2026-06-15, D-3 → 2026-06-19, due → 2026-06-22

due := notify.DueWithin(today, 7, []notify.Item{item})
fire := notify.RemindersOn(today, []notify.Item{item})
_ = notify.Emit(ctx, notify.WriterNotifier{W: os.Stdout}, fire)
```

`Notifier` is the plug: `WriterNotifier` (stdout), `FileNotifier`, `WebhookNotifier` (HTTP POST JSON), plus `RecordingNotifier` for tests. Inject `FixedClock` or `Options.Today` instead of wall time.

## Job / CLI

```bash
go run ./cmd/drc-notify -cases cases.json -within 7
go run ./cmd/drc-notify -cases cases.json -today 2026-06-15 -fire -notifier stdout
```

`-cases` is a JSON array (or `{"cases":[...]}`) of tenancy facts. Deadlines are always `clock.Compute` — this package does not re-derive statutory day counts.

```json
[
  {
    "id": "apt-1",
    "state": "CA",
    "vacated_on": "2026-06-01",
    "lease_ends_on": "2026-06-01"
  }
]
```

| Flag | Role |
| --- | --- |
| `-within N` | list cases whose **deadline** is today through today+N |
| `-fire` | emit reminders whose **FireOn** is today (D-7 / D-3 / due) |
| `-notifier` | `stdout` (default) / `file` / `webhook` |
| `-today` | YYYY-MM-DD fake clock |

## Out of scope

- Push / SMS SaaS, CRM campaigns, MCP
- 50-state tables (unknown state → `clock.ErrUnsupportedState`)
- Evidence timeline / itemized pack (#3)
- Possession-clock rule changes (#4)
