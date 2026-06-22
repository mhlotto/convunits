package shoes

import (
	"math"
	"strings"
	"testing"
)

func TestFootLengthMeters(t *testing.T) {
	tests := []struct {
		system           string
		size, wantMeters float64
	}{
		{"us-men", 10, (10.0 + 22) / 3 * 0.0254},
		{"us-women", 8.5, (8.5 + 21) / 3 * 0.0254},
		{"uk-adult", 9, (9.0 + 23) / 3 * 0.0254},
		{"eu", 43, (43.0*2/3 - 1.5) / 100},
		{"jp", 27, 0.27},
		{"mondo", 270, 0.27},
	}
	for _, tt := range tests {
		got, err := FootLengthMeters(tt.system, tt.size)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got-tt.wantMeters) > 1e-12 {
			t.Errorf("%s: got %g want %g", tt.system, got, tt.wantMeters)
		}
	}
}

func TestShoeErrors(t *testing.T) {
	_, err := FootLengthMeters("us", 10)
	if err == nil || err.Error() != `ambiguous shoe system "us"; use us-men or us-women` {
		t.Errorf("us: %v", err)
	}
	for _, system := range []string{"us-kids", "kids"} {
		_, err := FootLengthMeters(system, 10)
		if err == nil || !strings.Contains(err.Error(), "unsupported children's") {
			t.Errorf("%s: %v", system, err)
		}
	}
	if _, err := FootLengthMeters("wat", 10); err == nil || !strings.Contains(err.Error(), "unknown shoe system") {
		t.Fatal(err)
	}
	if _, err := FootLengthMeters("us-men", 10.25); err == nil || !strings.Contains(err.Error(), "whole or half") {
		t.Fatal(err)
	}
}

func TestSystems(t *testing.T) {
	want := []string{"eu", "jp", "mondo", "uk-adult", "us-men", "us-women"}
	got := Systems()
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
