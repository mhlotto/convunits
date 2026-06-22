package scales

import (
	"math"
	"strings"
	"testing"
)

func closeEnough(got, want float64) bool {
	return math.Abs(got-want) <= math.Max(1e-12, math.Abs(want)*1e-9)
}

func TestScalarScaleConversions(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		value      float64
		input, out string
		want       float64
	}{
		{10, "dB", "power-ratio", 10},
		{20, "dB", "amplitude-ratio", 10},
		{10, "power-ratio", "dB", 10},
		{1, "bel", "dB", 10},
		{10, "dB", "bel", 1},
		{7, "pH", "mol/L", 1e-7},
		{6, "pH", "H+", 1e-6},
		{1e-7, "mol/L", "pH", 7},
		{5, "mag", "brightness-ratio", 0.01},
		{100, "brightness-ratio", "mag", -5},
		{20, "m/s", "beaufort", 8},
		{72, "km/h", "beaufort", 8},
		{20 / 0.44704, "mph", "beaufort", 8},
		{20 * 3600 / 1852, "knot", "beaufort", 8},
		{36, "km/h", "m/s", 10},
		{10, "m/s", "mph", 10 / 0.44704},
		{12, "awg", "diameter-mm", 2.052525388},
		{12, "awg", "diameter-in", 2.052525388 / 25.4},
		{2.0525, "diameter-mm", "awg", 12.00010669},
	}
	for _, tt := range tests {
		t.Run(tt.input+"_to_"+tt.out, func(t *testing.T) {
			got, err := r.Convert(tt.value, tt.input, tt.out)
			if err != nil {
				t.Fatal(err)
			}
			if got.Value == nil || !closeEnough(*got.Value, tt.want) {
				t.Fatalf("got %+v, want %g", got, tt.want)
			}
		})
	}
}

func TestBeaufortRange(t *testing.T) {
	got, err := NewRegistry().Convert(5, "beaufort", "m/s")
	if err != nil {
		t.Fatal(err)
	}
	if got.Min == nil || got.Max == nil || *got.Min != 8.0 || *got.Max != 10.7 {
		t.Fatalf("got %+v", got)
	}
}

func TestUnboundedBeaufortRanges(t *testing.T) {
	r := NewRegistry()
	calm, err := r.Convert(0, "beaufort", "m/s")
	if err != nil {
		t.Fatal(err)
	}
	if calm.Min != nil || calm.Max == nil || *calm.Max != 0.5 {
		t.Fatalf("calm: %+v", calm)
	}
	hurricane, err := r.Convert(12, "beaufort", "m/s")
	if err != nil {
		t.Fatal(err)
	}
	if hurricane.Min == nil || *hurricane.Min != 32.7 || hurricane.Max != nil {
		t.Fatalf("hurricane: %+v", hurricane)
	}
}

func TestScaleErrors(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		value                   float64
		input, output, contains string
	}{
		{0, "power-ratio", "dB", "greater than zero"},
		{-1, "brightness-ratio", "mag", "greater than zero"},
		{2.5, "beaufort", "m/s", "invalid Beaufort"},
		{41, "awg", "diameter-mm", "invalid AWG"},
		{1, "dB", "pH", "incompatible scale families"},
		{1, "a4", "m/s", "paper size"},
		{1, "wat", "dB", "unknown scale"},
	}
	for _, tt := range tests {
		_, err := r.Convert(tt.value, tt.input, tt.output)
		if err == nil || !strings.Contains(err.Error(), tt.contains) {
			t.Errorf("%g %s to %s: got %v, want %q", tt.value, tt.input, tt.output, err, tt.contains)
		}
	}
}

func TestPaperSizes(t *testing.T) {
	tests := []struct {
		name          string
		width, height float64
	}{
		{"a4", 210, 297}, {"letter", 215.9, 279.4}, {"legal", 215.9, 355.6}, {"tabloid", 279.4, 431.8},
	}
	for _, tt := range tests {
		got, err := LookupPaperSize(tt.name)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.WidthM*1000-tt.width) > 1e-9 || math.Abs(got.HeightM*1000-tt.height) > 1e-9 {
			t.Errorf("%s: %+v", tt.name, got)
		}
	}
}
