package solve

import (
	"math"
	"strings"
	"testing"

	"convunits/internal/units"
)

func scalar(value float64, unit string) Quantity {
	return Quantity{Range: Interval{Min: value, Max: value}, Unit: unit}
}

func TestSolveEveryVariable(t *testing.T) {
	s := New(units.NewRegistry())
	tests := []struct {
		input  Quantity
		output string
		givens map[string]Quantity
		want   float64
	}{
		{scalar(10, "N"), "s", map[string]Quantity{"mass": scalar(2, "kg"), "distance": scalar(5, "m")}, 1},
		{scalar(2, "kg"), "N", map[string]Quantity{"distance": scalar(5, "m"), "time": scalar(1, "s")}, 10},
		{scalar(10, "N"), "kg", map[string]Quantity{"distance": scalar(5, "m"), "time": scalar(1, "s")}, 2},
		{scalar(10, "N"), "m", map[string]Quantity{"mass": scalar(2, "kg"), "time": scalar(1, "s")}, 5},
	}
	for _, tt := range tests {
		got, err := s.Solve(tt.input, tt.output, tt.givens)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.Range.Min-tt.want) > 1e-12 || got.Range.Min != got.Range.Max {
			t.Fatalf("got %+v, want %g", got.Range, tt.want)
		}
	}
}

func TestSolveConvertsGivenAndOutputUnits(t *testing.T) {
	s := New(units.NewRegistry())
	got, err := s.Solve(
		scalar(10, "N"), "ms",
		map[string]Quantity{"mass": scalar(2000, "g"), "distance": scalar(500, "cm")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.Range.Min != 1000 || got.Range.Max != 1000 {
		t.Fatalf("got %+v", got.Range)
	}
}

func TestSolveRangePropagation(t *testing.T) {
	s := New(units.NewRegistry())
	got, err := s.Solve(
		scalar(10, "N"), "s",
		map[string]Quantity{
			"mass":     {Range: Interval{Min: 1, Max: 3}, Unit: "kg"},
			"distance": {Range: Interval{Min: 4, Max: 6}, Unit: "m"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(got.Range.Min-math.Sqrt(0.4)) > 1e-12 || math.Abs(got.Range.Max-math.Sqrt(1.8)) > 1e-12 {
		t.Fatalf("got %+v", got.Range)
	}
}

func TestRangePropagationForEveryTarget(t *testing.T) {
	s := New(units.NewRegistry())
	tests := []struct {
		input  Quantity
		output string
		givens map[string]Quantity
		min    float64
		max    float64
	}{
		{
			Quantity{Range: Interval{Min: 1, Max: 2}, Unit: "kg"}, "N",
			map[string]Quantity{
				"distance": {Range: Interval{Min: 3, Max: 4}, Unit: "m"},
				"time":     {Range: Interval{Min: 2, Max: 3}, Unit: "s"},
			},
			1.0 / 3, 2,
		},
		{
			Quantity{Range: Interval{Min: 10, Max: 20}, Unit: "N"}, "kg",
			map[string]Quantity{
				"distance": {Range: Interval{Min: 4, Max: 5}, Unit: "m"},
				"time":     {Range: Interval{Min: 2, Max: 3}, Unit: "s"},
			},
			8, 45,
		},
		{
			Quantity{Range: Interval{Min: 10, Max: 20}, Unit: "N"}, "m",
			map[string]Quantity{
				"mass": {Range: Interval{Min: 4, Max: 5}, Unit: "kg"},
				"time": {Range: Interval{Min: 2, Max: 3}, Unit: "s"},
			},
			8, 45,
		},
	}
	for _, tt := range tests {
		got, err := s.Solve(tt.input, tt.output, tt.givens)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.Range.Min-tt.min) > 1e-12 || math.Abs(got.Range.Max-tt.max) > 1e-12 {
			t.Fatalf("target %s: got %+v, want %g..%g", tt.output, got.Range, tt.min, tt.max)
		}
	}
}

func TestSolveMissingGiven(t *testing.T) {
	_, err := New(units.NewRegistry()).Solve(
		scalar(10, "N"), "s", map[string]Quantity{"mass": scalar(2, "kg")},
	)
	if err == nil || !strings.Contains(err.Error(), "missing --given distance") {
		t.Fatalf("got %v", err)
	}
}
