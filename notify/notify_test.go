package notify

import "testing"

func TestSeat(t *testing.T) {
	if Seat != "notify" {
		t.Fatalf("notify seat: got %q", Seat)
	}
}
