package notify

import (
	"fmt"
	"io"

	"github.com/furyheimdall/deposit-return-clock/clock"
)

// FormatDue writes a due-within list (id, state, deadline, days, branch).
func FormatDue(w io.Writer, today clock.Date, items []Item) {
	_, _ = fmt.Fprintf(w, "today=%s due=%d\n", today, len(items))
	for _, it := range items {
		days := DaysUntilDeadline(today, it.Result)
		_, _ = fmt.Fprintf(w, "%s\t%s\tdeadline=%s\tdays=%d\tbranch=%s\n",
			it.ID, it.Result.State, it.Result.DeadlineOn, days, it.Result.Branch)
	}
}
