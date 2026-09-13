# Deposit Return Clock

Per-state possession clock + deduction evidence log → deadline-correct itemized return pack.

OSS core for US residential security-deposit return workflows. **Not a full PMS.**

**Copy anchors:** Wrong start date = auto loss. / Camera roll ≠ court pack.

**ICP:** Solo LL / small PM (rejects heavy PMS; Sheets/Drive inertia).

> **Not legal advice.** Dig-sourced deadlines; verify current statute.

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

Scaffold only. Domain logic lands in later issues — do not treat stubs as complete.

| Package | Seat | Later issue |
| --- | --- | --- |
| [`clock/`](clock/) | Possession clock (CA / NY / FL / NJ) | #4 |
| [`evidence/`](evidence/) | Evidence timeline | #3 |
| [`pack/`](pack/) | Itemized return pack (PDF + checklist) | #3 |
| [`notify/`](notify/) | Reminders D-7 / D-3 / due | #5 |

## Develop

```bash
go test ./...
```

Module: [`github.com/furyheimdall/deposit-return-clock`](https://github.com/furyheimdall/deposit-return-clock) · License: [MIT](LICENSE)

See [CONTRIBUTING.md](CONTRIBUTING.md). This repository is **not a full PMS**.
