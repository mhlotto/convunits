package formula

import (
	"math"
	"strings"
	"testing"

	"convunits/internal/units"
)

func input(value float64, unit string) Input { return Input{Value: value, Unit: unit} }

func TestFormulas(t *testing.T) {
	c := New(units.NewRegistry())
	tests := []struct {
		name            string
		inputs          map[string]Input
		output          string
		want, tolerance float64
	}{
		{"escape-velocity", map[string]Input{"mass": input(1, "Mearth"), "radius": input(1, "Re")}, "km/s", 11.186, .001},
		{"orbital-period", map[string]Input{"mass": input(1, "Msun"), "radius": input(1, "au")}, "d", 365.25, .02},
		{"orbital-speed", map[string]Input{"mass": input(1, "Msun"), "radius": input(1, "au")}, "km/s", 29.78, .02},
		{"freefall-time", map[string]Input{"height": input(100, "m")}, "s", 4.515, .01},
		{"pendulum-period", map[string]Input{"length": input(1, "m")}, "s", 2.006, .001},
		{"bmi", map[string]Input{"mass": input(180, "lb"), "height": input(6, "ft")}, "kg/m^2", 24.41, .02},
		{"bmi", map[string]Input{"mass": input(180, "lb"), "height": input(6, "ft")}, "bmi", 24.41, .02},
	}
	for _, tt := range tests {
		got, err := c.Compute(tt.name, tt.inputs, tt.output)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(got.Value-tt.want) > tt.tolerance {
			t.Errorf("%s: got %g %s, want %g", tt.name, got.Value, got.Unit, tt.want)
		}
	}
}

func TestFormulaErrors(t *testing.T) {
	c := New(units.NewRegistry())
	tests := []struct {
		name             string
		inputs           map[string]Input
		output, contains string
	}{
		{"escape-velocity", map[string]Input{"mass": input(1, "kg")}, "m/s", "missing required argument --radius"},
		{"escape-velocity", map[string]Input{"mass": input(1, "m"), "radius": input(1, "m")}, "m/s", "--mass"},
		{"escape-velocity", map[string]Input{"mass": input(1, "kg"), "radius": input(1, "m")}, "kg", "output unit"},
		{"bmi", map[string]Input{"mass": input(1, "kg"), "height": input(1, "m"), "radius": input(1, "m")}, "bmi", "does not accept --radius"},
	}
	for _, tt := range tests {
		_, err := c.Compute(tt.name, tt.inputs, tt.output)
		if err == nil || !strings.Contains(err.Error(), tt.contains) {
			t.Errorf("%s: got %v, want %q", tt.name, err, tt.contains)
		}
	}
}
