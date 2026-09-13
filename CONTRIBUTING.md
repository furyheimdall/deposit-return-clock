# Contributing

This repository is the OSS core of **Deposit Return Clock**. It is **not a full PMS**.

Stay inside the locked MVP IN box in the [README](README.md): CA/NY/FL/NJ possession clock, evidence timeline, itemized return pack (PDF + checklist), and reminders D-7 / D-3 / due.

Do not send PRs for a full PMS, tenant claim app, banking/escrow, lawyer automation, 50-state day-1 tables, or MCP/Deadbugz coupling.

```bash
go test ./...
```

Keep `clock/`, `evidence/`, `pack/`, and `notify/` as the package seats. `clock/` is the possession-clock library (#4). `evidence/` and `pack/` implement the #3 timeline and return-pack slice. `notify/` remains a stub until #5. Do not add 50-state tables or reminder jobs in `clock/`.
