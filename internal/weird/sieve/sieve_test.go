package sieve

import (
	"math"
	"testing"
)

func TestSieveOpenings(t *testing.T) {
	tests := []struct {
		size   string
		wantMM float64
	}{
		{"#40", .425}, {"No.200", .075}, {"No. 200", .075}, {"4mesh", 4.75}, {"mesh40", .425}, {"3/4in", 19},
	}
	for _, tt := range tests {
		got, err := OpeningMeters(tt.size)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got*1000-tt.wantMM) > 1e-12 {
			t.Errorf("%s: got %g mm", tt.size, got*1000)
		}
	}
}

func TestUnknownSieve(t *testing.T) {
	for _, size := range []string{"#5", "wat", "5mesh"} {
		if _, err := OpeningMeters(size); err == nil {
			t.Errorf("%s unexpectedly accepted", size)
		}
	}
}
