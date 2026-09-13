package pack

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ChecklistItem is one operator-facing export row.
type ChecklistItem struct {
	Item string `json:"item"`
	Done bool   `json:"done"`
}

// ChecklistDoc is the JSON export of an itemized return pack.
type ChecklistDoc struct {
	Tenant               string          `json:"tenant,omitempty"`
	Property             string          `json:"property,omitempty"`
	DepositCents         int64           `json:"deposit_cents"`
	TotalDeductionsCents int64           `json:"total_deductions_cents"`
	RemainingCents       int64           `json:"remaining_cents"`
	DueBy                string          `json:"due_by,omitempty"`
	DueByNote            string          `json:"due_by_note"`
	Lines                []checklistLine `json:"lines"`
	Evidence             []checklistEv   `json:"evidence"`
	Checklist            []ChecklistItem `json:"checklist"`
	Disclaimer           string          `json:"disclaimer"`
}

type checklistLine struct {
	Description string   `json:"description"`
	AmountCents int64    `json:"amount_cents"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type checklistEv struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	CapturedAt string `json:"captured_at"`
	Note       string `json:"note,omitempty"`
	URI        string `json:"uri,omitempty"`
}

func disclaimer() string {
	return "Not legal advice. Dig-sourced deadlines; verify current statute."
}

func (p Pack) checklistItems() []ChecklistItem {
	linked := true
	for _, line := range p.Lines {
		if len(line.EvidenceIDs) == 0 {
			linked = false
			break
		}
	}
	if len(p.Lines) == 0 {
		linked = false
	}
	return []ChecklistItem{
		{Item: "Itemized deduction lines recorded", Done: len(p.Lines) > 0},
		{Item: "Remaining balance computed toward return", Done: true},
		{Item: "Each deduction linked to evidence", Done: linked},
		{Item: "Evidence timeline attached (captured_at order)", Done: len(p.Evidence) > 0},
		{Item: "Return deadline noted (mock until clock/#4)", Done: !p.DueBy.IsZero()},
	}
}

// ChecklistJSON exports the pack as a structured checklist document.
func (p Pack) ChecklistJSON() ([]byte, error) {
	lines := make([]checklistLine, 0, len(p.Lines))
	for _, line := range p.Lines {
		ids := line.EvidenceIDs
		if ids == nil {
			ids = []string{}
		}
		lines = append(lines, checklistLine{
			Description: line.Description,
			AmountCents: int64(line.Amount),
			EvidenceIDs: ids,
		})
	}
	ev := make([]checklistEv, 0, len(p.Evidence))
	for _, e := range p.Evidence {
		ev = append(ev, checklistEv{
			ID:         e.ID,
			Kind:       string(e.Kind),
			CapturedAt: e.CapturedAt.UTC().Format(time.RFC3339),
			Note:       e.Note,
			URI:        e.URI,
		})
	}
	due := ""
	if !p.DueBy.IsZero() {
		due = p.DueBy.UTC().Format(time.RFC3339)
	}
	doc := ChecklistDoc{
		Tenant:               p.Tenant,
		Property:             p.Property,
		DepositCents:         int64(p.Deposit),
		TotalDeductionsCents: int64(p.TotalDeductions),
		RemainingCents:       int64(p.Remaining),
		DueBy:                due,
		DueByNote:            p.DueByNote,
		Lines:                lines,
		Evidence:             ev,
		Checklist:            p.checklistItems(),
		Disclaimer:           disclaimer(),
	}
	return json.MarshalIndent(doc, "", "  ")
}

// ChecklistMarkdown exports the pack as a markdown checklist / statement.
func (p Pack) ChecklistMarkdown() string {
	var b strings.Builder
	b.WriteString("# Itemized security deposit return\n\n")
	if p.Tenant != "" {
		fmt.Fprintf(&b, "- Tenant: %s\n", p.Tenant)
	}
	if p.Property != "" {
		fmt.Fprintf(&b, "- Property: %s\n", p.Property)
	}
	fmt.Fprintf(&b, "- Deposit: %s\n", p.Deposit)
	if !p.DueBy.IsZero() {
		fmt.Fprintf(&b, "- Due by: %s\n", p.DueBy.UTC().Format("2006-01-02"))
	}
	fmt.Fprintf(&b, "- Deadline note: %s\n\n", p.DueByNote)

	b.WriteString("## Deductions\n\n")
	if len(p.Lines) == 0 {
		b.WriteString("_No deductions._\n\n")
	} else {
		b.WriteString("| Description | Amount | Evidence |\n| --- | --- | --- |\n")
		for _, line := range p.Lines {
			ids := strings.Join(line.EvidenceIDs, ", ")
			if ids == "" {
				ids = "—"
			}
			fmt.Fprintf(&b, "| %s | %s | %s |\n", line.Description, line.Amount, ids)
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "**Total deductions:** %s\n\n", p.TotalDeductions)
	fmt.Fprintf(&b, "**Remaining balance:** %s\n\n", p.Remaining)

	if len(p.Evidence) > 0 {
		b.WriteString("## Evidence timeline\n\n")
		for _, e := range p.Evidence {
			fmt.Fprintf(&b, "- `%s` %s @ %s", e.ID, e.Kind, e.CapturedAt.UTC().Format(time.RFC3339))
			if e.Note != "" {
				fmt.Fprintf(&b, " — %s", e.Note)
			}
			if e.URI != "" {
				fmt.Fprintf(&b, " (%s)", e.URI)
			}
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	b.WriteString("## Checklist\n\n")
	for _, item := range p.checklistItems() {
		box := "[ ]"
		if item.Done {
			box = "[x]"
		}
		fmt.Fprintf(&b, "- %s %s\n", box, item.Item)
	}
	fmt.Fprintf(&b, "\n_%s_\n", disclaimer())
	return b.String()
}
