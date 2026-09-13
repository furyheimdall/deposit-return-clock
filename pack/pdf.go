package pack

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// PDF renders a one-or-more-page itemized return pack (PDF 1.4, Helvetica).
func (p Pack) PDF() ([]byte, error) {
	lines := p.pdfLines()
	pages := paginate(lines, 40)
	return writePDF(pages)
}

func (p Pack) pdfLines() []string {
	out := []string{
		"ITEMIZED SECURITY DEPOSIT RETURN",
		"",
	}
	if p.Tenant != "" {
		out = append(out, "Tenant: "+p.Tenant)
	}
	if p.Property != "" {
		out = append(out, "Property: "+p.Property)
	}
	out = append(out, "Deposit: "+p.Deposit.String())
	if !p.DueBy.IsZero() {
		out = append(out, "Due by: "+p.DueBy.UTC().Format("2006-01-02"))
	}
	out = append(out, "Deadline note: "+p.DueByNote, "")
	out = append(out, "DEDUCTIONS")
	if len(p.Lines) == 0 {
		out = append(out, "  (none)")
	} else {
		for i, line := range p.Lines {
			ids := strings.Join(line.EvidenceIDs, ", ")
			if ids == "" {
				ids = "none"
			}
			out = append(out, fmt.Sprintf("  %d. %s  %s  evidence: %s", i+1, line.Description, line.Amount, ids))
		}
	}
	out = append(out,
		"",
		"Total deductions: "+p.TotalDeductions.String(),
		"Remaining balance: "+p.Remaining.String(),
		"",
	)
	if len(p.Evidence) > 0 {
		out = append(out, "EVIDENCE TIMELINE")
		for _, e := range p.Evidence {
			row := fmt.Sprintf("  %s [%s] %s", e.ID, e.Kind, e.CapturedAt.UTC().Format(time.RFC3339))
			if e.Note != "" {
				row += " — " + e.Note
			}
			if e.URI != "" {
				row += " (" + e.URI + ")"
			}
			out = append(out, row)
		}
		out = append(out, "")
	}
	out = append(out, "CHECKLIST")
	for _, item := range p.checklistItems() {
		mark := "NO"
		if item.Done {
			mark = "YES"
		}
		out = append(out, fmt.Sprintf("  [%s] %s", mark, item.Item))
	}
	out = append(out, "", disclaimer())
	return out
}

func paginate(lines []string, perPage int) [][]string {
	if perPage < 1 {
		perPage = 40
	}
	if len(lines) == 0 {
		return [][]string{{}}
	}
	var pages [][]string
	for len(lines) > 0 {
		n := perPage
		if n > len(lines) {
			n = len(lines)
		}
		pages = append(pages, lines[:n])
		lines = lines[n:]
	}
	return pages
}

func writePDF(pages [][]string) ([]byte, error) {
	if len(pages) == 0 {
		pages = [][]string{{}}
	}
	nPages := len(pages)
	// objects: 1 catalog, 2 pages, 3 font, then per page: pageDict + contents
	fontID := 3
	firstPageID := 4
	nObj := firstPageID + 2*nPages - 1

	w := &pdfWriter{offsets: make([]int, nObj+1)}

	w.buf.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")

	// 1 Catalog
	w.startObj(1)
	w.buf.WriteString("<< /Type /Catalog /Pages 2 0 R >>\n")
	w.endObj()

	// 2 Pages
	kids := make([]string, nPages)
	for i := 0; i < nPages; i++ {
		kids[i] = fmt.Sprintf("%d 0 R", firstPageID+2*i)
	}
	w.startObj(2)
	fmt.Fprintf(&w.buf, "<< /Type /Pages /Kids [%s] /Count %d >>\n", strings.Join(kids, " "), nPages)
	w.endObj()

	// 3 Font
	w.startObj(fontID)
	w.buf.WriteString("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\n")
	w.endObj()

	for i, lines := range pages {
		pageID := firstPageID + 2*i
		contentID := pageID + 1
		stream := pageStream(lines)

		w.startObj(pageID)
		fmt.Fprintf(&w.buf, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents %d 0 R /Resources << /Font << /F1 %d 0 R >> >> >>\n", contentID, fontID)
		w.endObj()

		w.startObj(contentID)
		fmt.Fprintf(&w.buf, "<< /Length %d >>\nstream\n", len(stream))
		w.buf.Write(stream)
		w.buf.WriteString("\nendstream\n")
		w.endObj()
	}

	startxref := w.buf.Len()
	fmt.Fprintf(&w.buf, "xref\n0 %d\n", nObj+1)
	w.buf.WriteString("0000000000 65535 f \n")
	for id := 1; id <= nObj; id++ {
		fmt.Fprintf(&w.buf, "%010d 00000 n \n", w.offsets[id])
	}
	fmt.Fprintf(&w.buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", nObj+1, startxref)
	return w.buf.Bytes(), nil
}

type pdfWriter struct {
	buf     bytes.Buffer
	offsets []int
}

func (w *pdfWriter) startObj(id int) {
	w.offsets[id] = w.buf.Len()
	fmt.Fprintf(&w.buf, "%d 0 obj\n", id)
}

func (w *pdfWriter) endObj() {
	w.buf.WriteString("endobj\n")
}

func pageStream(lines []string) []byte {
	var b bytes.Buffer
	b.WriteString("BT\n/F1 11 Tf\n72 720 Td\n")
	for i, line := range lines {
		if i > 0 {
			b.WriteString("0 -16 Td\n")
		}
		fmt.Fprintf(&b, "(%s) Tj\n", pdfEscape(asciiLine(line)))
	}
	b.WriteString("ET")
	return b.Bytes()
}

func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

func asciiLine(s string) string {
	// Helvetica / WinAnsi: keep the pack readable without embedding fonts.
	s = strings.ReplaceAll(s, "—", "-")
	var b strings.Builder
	for _, r := range s {
		if r < 32 || r > 126 {
			b.WriteByte('?')
			continue
		}
		b.WriteRune(r)
	}
	if b.Len() > 110 {
		return b.String()[:107] + "..."
	}
	return b.String()
}
