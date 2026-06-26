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
		{"schwarzschild-radius", map[string]Input{"mass": input(1, "Msun")}, "km", 2.953, .001},
		{"schwarzschild-radius", map[string]Input{"mass": input(1, "Mearth")}, "mm", 8.87, .01},
		{"gravity-force", map[string]Input{"mass1": input(1, "Mearth"), "mass2": input(1, "kg"), "distance": input(1, "Re")}, "N", 9.82, .01},
		{"surface-gravity", map[string]Input{"mass": input(1, "Mearth"), "radius": input(1, "Re")}, "m/s^2", 9.82, .01},
		{"surface-gravity", map[string]Input{"mass": input(1, "Mearth"), "radius": input(1, "Re")}, "g0", 1.001, .001},
		{"density", map[string]Input{"mass": input(1, "kg"), "volume": input(1, "L")}, "kg/m^3", 1000, 1e-9},
		{"density", map[string]Input{"mass": input(1, "lb"), "volume": input(1, "gal")}, "g/cm^3", 0.119826, 1e-6},
		{"sphere-volume", map[string]Input{"radius": input(10, "cm")}, "L", 4.18879, 1e-5},
		{"sphere-volume", map[string]Input{"radius": input(1, "Re")}, "km^3", 1.08321e12, 1e7},
		{"sphere-area", map[string]Input{"radius": input(10, "cm")}, "cm^2", 1256.637, .001},
		{"sphere-area", map[string]Input{"radius": input(1, "Re")}, "km^2", 5.10066e8, 1e4},
		{"circle-area", map[string]Input{"radius": input(1, "ft")}, "in^2", 452.389, .001},
		{"cylinder-volume", map[string]Input{"radius": input(10, "cm"), "height": input(1, "m")}, "L", 31.4159, 1e-4},
		{"kinetic-energy", map[string]Input{"mass": input(1, "kg"), "speed": input(10, "m/s")}, "J", 50, 1e-12},
		{"kinetic-energy", map[string]Input{"mass": input(1500, "kg"), "speed": input(60, "mph")}, "kWh", 0.1498835712, 1e-10},
		{"momentum", map[string]Input{"mass": input(1, "kg"), "speed": input(10, "m/s")}, "kg*m/s", 10, 1e-12},
		{"power", map[string]Input{"energy": input(1, "kWh"), "time": input(1, "h")}, "W", 1000, 1e-9},
		{"energy", map[string]Input{"power": input(1000, "W"), "time": input(1, "h")}, "kWh", 1, 1e-12},
		{"pressure", map[string]Input{"force": input(1, "N"), "area": input(1, "m^2")}, "Pa", 1, 1e-12},
		{"pressure", map[string]Input{"force": input(1, "lbf"), "area": input(1, "in^2")}, "psi", 1, 1e-12},
		{"flow-rate", map[string]Input{"volume": input(1, "gal"), "time": input(1, "min")}, "gpm", 1, 1e-12},
		{"pace", map[string]Input{"distance": input(1, "mi"), "time": input(8, "min")}, "min/mi", 8, 1e-12},
		{"pace", map[string]Input{"distance": input(5, "km"), "time": input(25, "min")}, "min/km", 5, 1e-12},
		{"speed", map[string]Input{"distance": input(5, "km"), "time": input(25, "min")}, "mph", 7.45645, 1e-5},
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
		{"density", map[string]Input{"mass": input(1, "kg")}, "kg/m^3", "missing required argument --volume"},
		{"kinetic-energy", map[string]Input{"mass": input(1, "kg"), "speed": input(10, "kg")}, "J", "--speed"},
		{"pressure", map[string]Input{"force": input(1, "N"), "area": input(1, "m^2")}, "J", "output unit"},
	}
	for _, tt := range tests {
		_, err := c.Compute(tt.name, tt.inputs, tt.output)
		if err == nil || !strings.Contains(err.Error(), tt.contains) {
			t.Errorf("%s: got %v, want %q", tt.name, err, tt.contains)
		}
	}
}
