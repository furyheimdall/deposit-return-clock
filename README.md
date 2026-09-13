# Deposit Return Clock

**One-liner:** Per-state possession clock + deduction evidence log → a deadline-correct itemized return pack.

For **solo landlords and small PMs** who need to hit statutory deposit-return windows without a full PMS.

**Copy anchors:** *Wrong start date = auto loss.* · *Camera roll ≠ court pack.*

## Pilot state clocks (CA / NY / FL / NJ)

> **Not legal advice.** Deadlines below are a dig-matrix summary for product UX (as of dig 2026-09-13). Statutes change; verify against current law before relying on a deadline.

| State | Return deadline (pilot model) | Clock start / early-exit note | Risk highlight |
| --- | --- | --- | --- |
| **CA** | **21 calendar days** after tenant **vacates** — itemized statement + balance | Clock = **possession / vacate**, not scheduled lease end | Bad-faith exposure often discussed as deposit + up to **2×** (forum “3×”) |
| **NY** | **14 days** after tenant **vacated** — itemized statement **and** remaining deposit | Runs from **vacate/possession** | Miss deadline → **forfeit** right to retain any of the deposit |
| **FL** | **No claim:** return ≤**15 days** after **termination of rental agreement**. **With claim:** notice ≤**30 days** (then remittance path after tenant object window) | Clock from **termination**; early abandon / notice quirks are **fact-sensitive** — keep 15/30 branches explicit in UI | Miss claim-notice window → **forfeit claim against deposit** |
| **NJ** | **30 days** after **termination of lease** — return + interest + itemization | Statute = **lease termination**, not bare move-out; premature vacation may shift clock toward **re-let** (case-law sensitive) | Wrongful withhold → **double** + costs risk |

## MVP IN

1. State clocks above + early-exit vs lease-end branch UI
2. Move-out photo / receipt / note evidence timeline tied to each deduction
3. Itemized return pack export (PDF + checklist)
4. Reminders: D-7 / D-3 / due

## MVP OUT

- Full PMS / accounting / lease admin (AppFolio/Buildium-class)
- Banking / escrow replacement (Baselane-class)
- Lawyer advice / lawsuit automation
- Tenant-facing claim app
- All 50 states on day one (4-state pilot first)
- Deadbugz / MCP (separate track)

## Not this

| This is | This is not |
| --- | --- |
| Deadline + evidence → return pack | A property-management suite |
| Landlord / small-PM workflow | A tenant refund marketplace |
| 4-state pilot clock | Nationwide legal counsel |

## Who it’s for

Solo landlords and small PM shops on Sheets/Drive inertia who keep missing statutory windows — not enterprises replacing AppFolio/Buildium.

## Get started (thin self-serve)

```text
clone → pick state + possession/termination date → attach evidence → export pack
```

No quote form / seat-pricing hero on this page. OSS-first core; cloud niceties later if any.

## Competitive wedge (short)

Baselane/Avail: deadline clock not first-class. TenantCloud/TurboTenant: partial. Full PMS: overkill. Gap = **statutory clock + evidence → pack**.

## License

TBD (prefer OSS-friendly for clock/pack core).
