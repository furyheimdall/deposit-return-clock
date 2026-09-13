# Contributing

This repository is the OSS core of **Deposit Return Clock**. It is **not a full PMS**.

Stay inside the locked MVP IN box in the [README](README.md): CA/NY/FL/NJ possession clock, evidence timeline, itemized return pack (PDF + checklist), and reminders D-7 / D-3 / due.

Do not send PRs for a full PMS, tenant claim app, banking/escrow, lawyer automation, 50-state day-1 tables, or MCP/Deadbugz coupling.

```bash
go test ./...
```

Keep `clock/`, `evidence/`, `pack/`, and `notify/` as the package seats. Stubs are intentional until those slices land.
