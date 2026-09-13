# Deposit Return Clock

Per-state possession clock + deduction evidence log → deadline-correct itemized return pack.

OSS core for US residential security-deposit return workflows. **Not a full PMS.**

**Copy anchors:** Wrong start date = auto loss. / Camera roll ≠ court pack.

**ICP:** Solo LL / small PM (rejects heavy PMS; Sheets/Drive inertia).

> **Not legal advice.** Dig-sourced deadlines; verify current statute.
> Possession-clock cites and early-exit notes: [`clock/README.md`](clock/README.md).

## MVP IN (locked)

1. **CA / NY / FL / NJ** statutory return clock + early move-out vs lease-end branch

   | State | Clock |
   | --- | --- |
   | **CA** | vacate + **21** |
   | **NY** | vacate + **14** |
   | **FL** | no-claim ≤**15** / claim-notice ≤**30** |
   | **NJ** | lease-term + **30** (early vacate re-let sensitive) |

2. Photo / receipt / memo **evidence timeline** → linked deduction lines
3. **Itemized return pack** export (PDF + checklist)
4. Deadline reminders **D-7 / D-3 / due**

## MVP OUT (HOLD)

- Full PMS / accounting / lease admin
- Tenant claim / dispute app as day-1
- Banking / escrow replacement
- Lawyer / litigation automation
- All 50 states on day-1
- MCP / Deadbugz coupling

## Package seats

| Package | Seat | Status |
| --- | --- | --- |
| [`clock/`](clock/) | Possession clock (CA / NY / FL / NJ) | Implemented (#4). Pure library; see [`clock/README.md`](clock/README.md). |
| [`evidence/`](evidence/) | Append-only evidence timeline (photo / receipt / attachment / memo) | Implemented (#3) |
| [`pack/`](pack/) | Itemized return pack: deduction lines, remaining balance, PDF + checklist | Implemented (#3) |
| [`notify/`](notify/) | Reminders D-7 / D-3 / due | Implemented (#5). Schedule from `clock.Result`; pluggable stdout / file / webhook. CLI: `go run ./cmd/drc-notify`. See [`notify/README.md`](notify/README.md). |

`pack` currently accepts a caller-supplied mock `DueBy`; wiring to `clock.Result` is a follow-up.

## Develop

```bash
go test ./...
go run ./cmd/drc-notify -cases cases.json -within 7
go run ./cmd/drc-notify -cases cases.json -fire -notifier stdout
```

Module: [`github.com/furyheimdall/deposit-return-clock`](https://github.com/furyheimdall/deposit-return-clock) · License: [MIT](LICENSE)

See [CONTRIBUTING.md](CONTRIBUTING.md). This repository is **not a full PMS**.
