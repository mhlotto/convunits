package units

import (
	"math"
	"strings"
	"testing"
)

func TestConversions(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		in       float64
		from, to string
		want     float64
	}{
		{1, "m", "cm", 100}, {1000, "ms", "s", 1}, {1, "kg", "g", 1000},
		{1, "lb", "kg", 0.45359237}, {60, "mph", "km/h", 96.56064},
		{1, "N", "kg*m/s^2", 1}, {1, "J", "N*m", 1}, {1, "W", "J/s", 1}, {1, "Pa", "N/m^2", 1},
		{32, "F", "C", 0}, {100, "C", "F", 212}, {273.15, "K", "C", 0},
		{1, "MB", "B", 1e6}, {1, "MiB", "B", 1048576}, {8, "b", "B", 1},
		{1, "C", "A*s", 1}, {1, "F", "A^2*s^4/(kg*m^2)", 1},
		{1, "g/cm^3", "kg/m^3", 1000}, {1, "gal/min", "gpm", 1},
	}
	for _, tt := range tests {
		t.Run(tt.from+"_to_"+tt.to, func(t *testing.T) {
			got, err := r.Convert(tt.in, tt.from, tt.to)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got.Value-tt.want) > math.Max(1e-9, math.Abs(tt.want)*1e-9) {
				t.Fatalf("got %.12g want %.12g", got.Value, tt.want)
			}
		})
	}
}

func TestIncompatible(t *testing.T) {
	_, err := NewRegistry().Convert(1, "N", "s")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "incompatible dimensions") || !strings.Contains(err.Error(), "kg*m/s^2") {
		t.Fatalf("unhelpful error: %v", err)
	}
}

func TestAffineInCompound(t *testing.T) {
	_, err := NewRegistry().Convert(1, "celsius/s", "K/s")
	if err == nil || !strings.Contains(err.Error(), "affine") {
		t.Fatalf("got %v", err)
	}
}

func TestAffineZeroIsNormalized(t *testing.T) {
	got, err := NewRegistry().Convert(32, "F", "C")
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != 0 {
		t.Fatalf("got %g, want exact zero", got.Value)
	}
}

func TestReciprocalFuelConsumption(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		value    float64
		from, to string
		want     float64
	}{
		{30, "mpg", "L/100km", 7.840486111111111},
		{7.840486111111111, "L/100km", "mpg", 30},
		{20, "km/L", "L/100km", 5},
		{5, "L/100km", "km/L", 20},
	}
	for _, tt := range tests {
		got, err := r.Convert(tt.value, tt.from, tt.to)
		if err != nil {
			t.Fatalf("%s to %s: %v", tt.from, tt.to, err)
		}
		if math.Abs(got.Value-tt.want) > 1e-9 {
			t.Fatalf("%g %s to %s: got %.12g, want %.12g", tt.value, tt.from, tt.to, got.Value, tt.want)
		}
	}
}

func TestFuelConsumptionRejectsZero(t *testing.T) {
	_, err := NewRegistry().Convert(0, "mpg", "L/100km")
	if err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("got %v", err)
	}
}
