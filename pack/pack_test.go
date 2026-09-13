package pack

import "testing"

func TestSeat(t *testing.T) {
	if Seat != "pack" {
		t.Fatalf("pack seat: got %q", Seat)
	}
}
