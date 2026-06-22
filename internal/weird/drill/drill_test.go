package drill

import (
	"math"
	"testing"
)

func TestDrillDiameters(t *testing.T) {
	tests := []struct {
		size   string
		wantIn float64
	}{
		{"#7", .2010}, {"A", .2340}, {"a", .2340}, {"Z", .4130}, {"1/4", .25}, {"6.8mm", 6.8 / 25.4},
	}
	for _, tt := range tests {
		got, err := DiameterMeters(tt.size)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got/0.0254-tt.wantIn) > 1e-12 {
			t.Errorf("%s: got %g in, want %g", tt.size, got/0.0254, tt.wantIn)
		}
	}
}

func TestUnknownDrill(t *testing.T) {
	for _, size := range []string{"7", "#81", "AA", "1/128"} {
		if _, err := DiameterMeters(size); err == nil {
			t.Errorf("%s unexpectedly accepted", size)
		}
	}
}

func TestDrillListings(t *testing.T) {
	for _, category := range []string{"number", "letter", "fractional"} {
		entries, err := Entries(category)
		if err != nil || len(entries) == 0 {
			t.Errorf("%s: entries=%d err=%v", category, len(entries), err)
		}
	}
}
