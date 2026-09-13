package clock

import "testing"

func TestSeat(t *testing.T) {
	if Seat != "clock" {
		t.Fatalf("clock seat: got %q", Seat)
	}
}
