package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func run(args ...string) (int, string, string) {
	var out, err bytes.Buffer
	code := New(&out, &err).Run(args)
	return code, out.String(), err.String()
}

func TestAttachedAndSpacedInput(t *testing.T) {
	for _, args := range [][]string{{"1m", "cm"}, {"1", "m", "cm"}} {
		code, out, err := run(args...)
		if code != 0 || out != "100 cm\n" {
			t.Fatalf("%v: code=%d out=%q err=%q", args, code, out, err)
		}
	}
}

func TestFlagsBeforeOperands(t *testing.T) {
	code, out, err := run("--precision", "4", "10kg", "lb")
	if code != 0 || out != "22.05 lb\n" {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestJSON(t *testing.T) {
	code, out, err := run("--json", "10kg", "lb")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var v struct {
		Input struct {
			Value float64
			Unit  string
		}
		Output struct {
			Value float64
			Unit  string
		}
		Dimension string
	}
	if e := json.Unmarshal([]byte(out), &v); e != nil {
		t.Fatal(e)
	}
	if v.Input.Value != 10 || v.Input.Unit != "kg" || v.Output.Unit != "lb" || v.Dimension != "mass" {
		t.Fatalf("%+v", v)
	}
}

func TestErrors(t *testing.T) {
	code, _, err := run("1N", "s")
	if code == 0 || !strings.Contains(err, "incompatible dimensions") {
		t.Fatalf("code=%d err=%q", code, err)
	}
}

func TestFuelConsumptionCLI(t *testing.T) {
	code, out, err := run("30mpg", "L/100km")
	if code != 0 || out != "7.840486111 L/100km\n" {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestSolveCLI(t *testing.T) {
	code, out, err := run("solve", "10N", "s", "--given", "mass=2kg", "--given", "distance=5m")
	if code != 0 || out != "1 s\n" {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestSolveRangeCLI(t *testing.T) {
	code, out, err := run("solve", "10N", "s", "--given", "mass=1..3kg", "--given", "distance=4..6m")
	if code != 0 || out != "0.632455532-1.341640786 s\n" {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestSolveJSON(t *testing.T) {
	code, out, err := run("solve", "10N", "s", "--given=mass=2kg", "--given=distance=5m", "--json")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Variable string
		Minimum  float64
		Maximum  float64
		Unit     string
	}
	if json.Unmarshal([]byte(out), &got) != nil || got.Variable != "time" || got.Minimum != 1 || got.Maximum != 1 || got.Unit != "s" {
		t.Fatalf("output=%q decoded=%+v", out, got)
	}
}

func TestAncientUnitsListingIncludesApproximationNote(t *testing.T) {
	code, out, err := run("units", "ancient-length")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	if !strings.Contains(out, "cubit") || !strings.Contains(out, "approximate: varies by culture; using 18 inches") {
		t.Fatalf("unexpected listing:\n%s", out)
	}
}

func TestCollisionSensitiveCategoryListings(t *testing.T) {
	tests := []struct {
		category string
		contains []string
	}{
		{"ancient-length", []string{"cubit", "approximate: varies by culture; using 18 inches"}},
		{"astronomical-length", []string{"lightmin", "symbol lm is reserved for lumen", "approximate: nominal solar radius"}},
		{"radiation", []string{"Gy", "absorbed dose (J/kg)", "rad is reserved for radians"}},
		{"typography-length", []string{"point", "pt remains the pint symbol", "approximate: historically variable"}},
	}
	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			code, out, err := run("units", tt.category)
			if code != 0 {
				t.Fatalf("code=%d err=%q", code, err)
			}
			for _, want := range tt.contains {
				if !strings.Contains(out, want) {
					t.Errorf("output does not contain %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestScaleCLI(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"scale", "10", "dB", "power-ratio"}, "10 power-ratio\n"},
		{[]string{"scale", "20", "dB", "amplitude-ratio"}, "10 amplitude-ratio\n"},
		{[]string{"scale", "7", "pH", "mol/L"}, "1e-07 mol/L\n"},
		{[]string{"scale", "5", "mag", "brightness-ratio"}, "0.01 brightness-ratio\n"},
		{[]string{"scale", "5", "beaufort", "m/s"}, "8-10.7 m/s\n"},
		{[]string{"scale", "5", "beaufort", "mph"}, "17.89549034-23.93521832 mph\n"},
		{[]string{"scale", "20", "m/s", "beaufort"}, "8 beaufort\n"},
		{[]string{"scale", "12", "awg", "diameter-mm"}, "2.052525388 diameter-mm\n"},
		{[]string{"scale", "-5", "mag", "brightness-ratio"}, "100 brightness-ratio\n"},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 || out != tt.want {
			t.Errorf("%v: code=%d out=%q err=%q, want %q", tt.args, code, out, err, tt.want)
		}
	}
}

func TestScaleRangeJSON(t *testing.T) {
	code, out, err := run("scale", "--json", "5", "beaufort", "m/s")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Output struct {
			Min, Max float64
			Unit     string
		}
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Output.Min != 8 || got.Output.Max != 10.7 || got.Output.Unit != "m/s" {
		t.Fatalf("%+v", got)
	}
}

func TestScaleListing(t *testing.T) {
	code, out, err := run("scales")
	if code != 0 {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
	for _, want := range []string{"dB", "bel", "power-ratio", "amplitude-ratio", "pH", "H+", "mol/L", "Molar", "mag", "brightness-ratio", "beaufort"} {
		if !strings.Contains(out, want) {
			t.Errorf("scales output missing %q:\n%s", want, out)
		}
	}
	code, out, err = run("scales", "ratio")
	if code != 0 || !strings.Contains(out, "dB") || strings.Contains(out, "pH") {
		t.Fatalf("filtered listing: code=%d out=%q err=%q", code, out, err)
	}
}

func TestPaperSizeCLI(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"size", "a4", "mm"}, "210 x 297 mm\n"},
		{[]string{"scale-size", "letter", "in"}, "8.5 x 11 in\n"},
		{[]string{"paper", "legal", "cm"}, "21.59 x 35.56 cm\n"},
		{[]string{"paper", "a0", "m"}, "0.841 x 1.189 m\n"},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 || out != tt.want {
			t.Errorf("%v: code=%d out=%q err=%q", tt.args, code, out, err)
		}
	}
}

func TestPaperErrorsAndListings(t *testing.T) {
	for _, tt := range []struct {
		args     []string
		contains string
	}{
		{[]string{"paper", "a4", "kg"}, "must be length"},
		{[]string{"paper", "wat", "mm"}, "unknown paper size"},
	} {
		code, _, err := run(tt.args...)
		if code == 0 || !strings.Contains(err, tt.contains) {
			t.Errorf("%v: code=%d err=%q", tt.args, code, err)
		}
	}
	code, out, err := run("papers", "photo")
	if code != 0 || !strings.Contains(out, "photo4x6") || strings.Contains(out, "letter") {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestWireCLI(t *testing.T) {
	tests := []struct{ gauge, unit, contains string }{
		{"12awg", "mm", "approximately 2.052525388 mm diameter"},
		{"0000awg", "mm", "approximately 11.684 mm diameter"},
		{"-3awg", "in", "approximately 0.46 in diameter"},
		{"awg40", "mm", "approximately 0.07987108513 mm diameter"},
	}
	for _, tt := range tests {
		code, out, err := run("wire", tt.gauge, tt.unit)
		if code != 0 || !strings.Contains(out, tt.contains) {
			t.Errorf("%s: code=%d out=%q err=%q", tt.gauge, code, out, err)
		}
	}
	for _, args := range [][]string{{"wire", "12", "kg"}, {"wire", "wat", "mm"}} {
		code, _, _ := run(args...)
		if code == 0 {
			t.Errorf("%v unexpectedly succeeded", args)
		}
	}
	code, out, err := run("wires")
	if code != 0 || !strings.Contains(out, "0000awg") {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestDrillCLI(t *testing.T) {
	tests := []struct{ size, unit, want string }{
		{"#7", "mm", "approximately 5.1054 mm diameter\n"},
		{"A", "mm", "approximately 5.9436 mm diameter\n"},
		{"Z", "in", "approximately 0.413 in diameter\n"},
		{"1/4", "mm", "approximately 6.35 mm diameter\n"},
		{"6.8mm", "in", "approximately 0.2677165354 in diameter\n"},
	}
	for _, tt := range tests {
		code, out, err := run("drill", tt.size, tt.unit)
		if code != 0 || out != tt.want {
			t.Errorf("%s: code=%d out=%q err=%q", tt.size, code, out, err)
		}
	}
	for _, args := range [][]string{{"drill", "#7", "kg"}, {"drill", "wat", "mm"}} {
		code, _, _ := run(args...)
		if code == 0 {
			t.Errorf("%v unexpectedly succeeded", args)
		}
	}
	for _, category := range []string{"number", "letter", "fractional"} {
		code, out, err := run("drills", category)
		if code != 0 || out == "" {
			t.Errorf("%s: code=%d err=%q", category, code, err)
		}
	}
}

func TestSieveCLI(t *testing.T) {
	tests := []struct{ size, unit, want string }{
		{"#40", "mm", "approximately 0.425 mm opening\n"},
		{"No.200", "um", "approximately 75 um opening\n"},
		{"4mesh", "mm", "approximately 4.75 mm opening\n"},
	}
	for _, tt := range tests {
		code, out, err := run("sieve", tt.size, tt.unit)
		if code != 0 || out != tt.want {
			t.Errorf("%s: code=%d out=%q err=%q", tt.size, code, out, err)
		}
	}
	for _, args := range [][]string{{"sieve", "#40", "kg"}, {"sieve", "wat", "mm"}} {
		code, _, _ := run(args...)
		if code == 0 {
			t.Errorf("%v unexpectedly succeeded", args)
		}
	}
	code, out, err := run("sieves")
	if code != 0 || !strings.Contains(out, "#400") {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestFormulaCLI(t *testing.T) {
	tests := []struct {
		args   []string
		prefix string
	}{
		{[]string{"formula", "escape-velocity", "--mass", "1Mearth", "--radius", "1Re", "km/s"}, "approximately 11.186"},
		{[]string{"formula", "orbital-period", "--mass", "1Msun", "--radius", "1au", "d"}, "approximately 365.25"},
		{[]string{"formula", "orbital-speed", "--mass", "1Msun", "--radius", "1au", "km/s"}, "approximately 29.785"},
		{[]string{"formula", "freefall-time", "--height", "100m", "s"}, "approximately 4.516"},
		{[]string{"formula", "pendulum-period", "--length", "1m", "s"}, "approximately 2.006"},
		{[]string{"formula", "bmi", "--mass", "180lb", "--height", "6ft", "bmi"}, "24.412"},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 || !strings.HasPrefix(out, tt.prefix) {
			t.Errorf("%v: code=%d out=%q err=%q", tt.args, code, out, err)
		}
	}
	for _, tt := range []struct {
		args     []string
		contains string
	}{
		{[]string{"formula", "escape-velocity", "--mass", "1kg", "m/s"}, "missing required argument --radius"},
		{[]string{"formula", "escape-velocity", "--mass", "1m", "--radius", "1m", "m/s"}, "--mass"},
		{[]string{"formula", "escape-velocity", "--mass", "1kg", "--radius", "1m", "kg"}, "output unit"},
	} {
		code, _, err := run(tt.args...)
		if code == 0 || !strings.Contains(err, tt.contains) {
			t.Errorf("%v: code=%d err=%q", tt.args, code, err)
		}
	}
	code, out, err := run("formulas")
	if code != 0 || !strings.Contains(out, "escape-velocity") || !strings.Contains(out, "bmi") {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestWrongScaleCommandErrors(t *testing.T) {
	tests := []struct {
		args     []string
		contains string
	}{
		{[]string{"scale", "1", "a4", "m/s"}, "paper size"},
		{[]string{"size", "dB", "mm"}, "scalar scale"},
		{[]string{"scale", "1", "dB", "pH"}, "incompatible scale families"},
	}
	for _, tt := range tests {
		code, _, err := run(tt.args...)
		if code == 0 || !strings.Contains(err, tt.contains) {
			t.Errorf("%v: code=%d err=%q", tt.args, code, err)
		}
	}
}

func TestShoeCLI(t *testing.T) {
	tests := []struct{ system, size, unit, want string }{
		{"us-men", "10", "in", "approximately 10.66666667 in foot length\n"},
		{"us-men", "10", "yd", "approximately 0.2962962963 yd foot length\n"},
		{"us-women", "8.5", "cm", "approximately 24.97666667 cm foot length\n"},
		{"uk-adult", "9", "cm", "approximately 27.09333333 cm foot length\n"},
		{"eu", "43", "cm", "approximately 27.16666667 cm foot length\n"},
		{"jp", "27", "in", "approximately 10.62992126 in foot length\n"},
		{"mondo", "270", "cm", "approximately 27 cm foot length\n"},
	}
	for _, tt := range tests {
		code, out, err := run("shoe", tt.system, tt.size, tt.unit)
		if code != 0 || out != tt.want {
			t.Errorf("%s: code=%d out=%q err=%q", tt.system, code, out, err)
		}
	}
}

func TestShoeErrors(t *testing.T) {
	tests := []struct {
		args     []string
		contains string
	}{
		{[]string{"shoe", "us", "10", "cm"}, `ambiguous shoe system "us"; use us-men or us-women`},
		{[]string{"shoe", "us-kids", "10", "cm"}, "children"},
		{[]string{"shoe", "us-men", "wat", "cm"}, "invalid shoe size"},
		{[]string{"shoe", "us-men", "10", "kg"}, "must be length"},
	}
	for _, tt := range tests {
		code, _, err := run(tt.args...)
		if code == 0 || !strings.Contains(err, tt.contains) {
			t.Errorf("%v: code=%d err=%q", tt.args, code, err)
		}
	}
}

func TestShoeSystemsListing(t *testing.T) {
	code, out, err := run("shoe", "systems")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	want := "eu\njp\nmondo\nuk-adult\nus-men\nus-women\n"
	if out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}
