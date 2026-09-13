package evidence

import "testing"

func TestSeat(t *testing.T) {
	if Seat != "evidence" {
		t.Fatalf("evidence seat: got %q", Seat)
	}
}
