package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/furyheimdall/deposit-return-clock/clock"
)

// FileCase is the JSON shape for a tenancy the CLI/job can load.
// Dates are YYYY-MM-DD. clock.Compute is the only deadline source.
type FileCase struct {
	ID                  string `json:"id"`
	State               string `json:"state"`
	VacatedOn           string `json:"vacated_on"`
	LeaseEndsOn         string `json:"lease_ends_on"`
	EarlyExit           bool   `json:"early_exit"`
	ClaimAgainstDeposit bool   `json:"claim_against_deposit"`
	ReletOn             string `json:"relet_on"`
}

type caseFile struct {
	Cases []FileCase `json:"cases"`
}

// LoadCases reads a JSON array of FileCase, or an object {"cases":[...]}.
func LoadCases(r io.Reader) ([]Item, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return nil, nil
	}

	var rows []FileCase
	switch raw[0] {
	case '[':
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, fmt.Errorf("notify: cases: %w", err)
		}
	case '{':
		var wrap caseFile
		if err := json.Unmarshal(raw, &wrap); err != nil {
			return nil, fmt.Errorf("notify: cases: %w", err)
		}
		rows = wrap.Cases
	default:
		return nil, fmt.Errorf("notify: cases: expected JSON array or object")
	}

	items := make([]Item, 0, len(rows))
	for i, row := range rows {
		item, err := row.Item()
		if err != nil {
			id := row.ID
			if id == "" {
				id = fmt.Sprintf("index %d", i)
			}
			return nil, fmt.Errorf("notify: case %s: %w", id, err)
		}
		items = append(items, item)
	}
	return items, nil
}

// LoadCasesFile opens path and LoadCases. "-" reads stdin.
func LoadCasesFile(path string) ([]Item, error) {
	if path == "" || path == "-" {
		return LoadCases(os.Stdin)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadCases(f)
}

// Item converts c through clock.Compute.
func (c FileCase) Item() (Item, error) {
	in, err := c.Input()
	if err != nil {
		return Item{}, err
	}
	return NewItem(c.ID, in)
}

// Input maps JSON fields onto clock.Input.
func (c FileCase) Input() (clock.Input, error) {
	vacated, err := parseCivilDate(c.VacatedOn)
	if err != nil {
		return clock.Input{}, err
	}
	leaseEnd, err := parseCivilDate(c.LeaseEndsOn)
	if err != nil {
		return clock.Input{}, err
	}
	relet, err := parseCivilDate(c.ReletOn)
	if err != nil {
		return clock.Input{}, err
	}
	return clock.Input{
		State:               clock.State(strings.TrimSpace(c.State)),
		VacatedOn:           vacated,
		LeaseEndsOn:         leaseEnd,
		EarlyExit:           c.EarlyExit,
		ClaimAgainstDeposit: c.ClaimAgainstDeposit,
		ReletOn:             relet,
	}, nil
}

func parseCivilDate(s string) (clock.Date, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return clock.Date{}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return clock.Date{}, fmt.Errorf("date %q: want YYYY-MM-DD", s)
	}
	return clock.DateFromTime(t), nil
}
