package cli

import (
	"bytes"
	"encoding/json"
	"math"
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

func TestDiscoverySearch(t *testing.T) {
	tests := []struct {
		args  []string
		wants []string
	}{
		{[]string{"search", "jupiter"}, []string{"unit", "Rj", "Mjup"}},
		{[]string{"search", "flour"}, []string{"ingredient", "all-purpose-flour"}},
		{[]string{"search", "mph"}, []string{"unit", "mph"}},
		{[]string{"search", "beaufort"}, []string{"scale", "beaufort"}},
		{[]string{"search", "schwarzschild"}, []string{"formula", "schwarzschild-radius"}},
		{[]string{"search", "cubit"}, []string{"unit", "cubit"}},
		{[]string{"search", "#40"}, []string{"sieve", "#40"}},
		{[]string{"search", "a4"}, []string{"paper", "a4"}},
		{[]string{"search", "eval"}, []string{"command", "eval"}},
		{[]string{"search", "jupiter", "--all"}, []string{"Rj", "Mjup"}},
		{[]string{"search", "jupiter", "--kind", "unit"}, []string{"unit", "Rj", "Mjup"}},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 {
			t.Fatalf("%v: code=%d err=%q", tt.args, code, err)
		}
		for _, want := range tt.wants {
			if !strings.Contains(out, want) {
				t.Errorf("%v: output missing %q:\n%s", tt.args, want, out)
			}
		}
	}
}

func TestDiscoveryUnitRankingBeforeScale(t *testing.T) {
	code, out, err := run("aliases", "mph")
	if code != 0 {
		t.Fatalf("aliases mph: code=%d err=%q", code, err)
	}
	unitIndex := strings.Index(out, "kind: unit")
	scaleIndex := strings.Index(out, "kind: scale")
	if unitIndex < 0 || scaleIndex < 0 {
		t.Fatalf("aliases mph should include unit and scale matches:\n%s", out)
	}
	if unitIndex > scaleIndex {
		t.Fatalf("aliases mph should rank unit before scale:\n%s", out)
	}

	code, out, err = run("search", "mph")
	if code != 0 {
		t.Fatalf("search mph: code=%d err=%q", code, err)
	}
	unitIndex = strings.Index(out, "unit")
	scaleIndex = strings.Index(out, "scale")
	if unitIndex < 0 || scaleIndex < 0 {
		t.Fatalf("search mph should include unit and scale matches:\n%s", out)
	}
	if unitIndex > scaleIndex {
		t.Fatalf("search mph should rank unit before scale:\n%s", out)
	}

	code, out, err = run("search", "beaufort")
	if code != 0 || !strings.Contains(out, "scale") || !strings.Contains(out, "beaufort") {
		t.Fatalf("search beaufort: code=%d out=%q err=%q", code, out, err)
	}
}

func TestDiscoveryNumericSearch(t *testing.T) {
	code, out, err := run("search", "38")
	if code != 0 {
		t.Fatalf("search 38: code=%d err=%q", code, err)
	}
	if !strings.Contains(out, "drill") || !strings.Contains(out, "#38") {
		t.Fatalf("search 38 should find normalized drill #38:\n%s", out)
	}
	for _, noisy := range []string{"#62", "11/32", "19/32", "#400", "0.038"} {
		if strings.Contains(out, noisy) {
			t.Fatalf("search 38 returned numeric substring noise %q:\n%s", noisy, out)
		}
	}

	code, out, err = run("search", "#38")
	if code != 0 {
		t.Fatalf("search #38: code=%d err=%q", code, err)
	}
	if !strings.Contains(out, "drill") || !strings.Contains(out, "#38") {
		t.Fatalf("search #38 should find drill #38:\n%s", out)
	}

	code, out, err = run("search", "#40")
	if code != 0 {
		t.Fatalf("search #40: code=%d err=%q", code, err)
	}
	for _, want := range []string{"drill  #40", "sieve  #40", "sieve  No. 40"} {
		if !strings.Contains(out, want) {
			t.Fatalf("search #40 missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "#400") {
		t.Fatalf("search #40 should not include #400 without --all:\n%s", out)
	}

	code, out, err = run("search", "#40", "--all")
	if code != 0 {
		t.Fatalf("search #40 --all: code=%d err=%q", code, err)
	}
	if !strings.Contains(out, "#400") {
		t.Fatalf("search #40 --all should allow broad substring matches:\n%s", out)
	}

	code, out, err = run("search", "38", "--all")
	if code != 0 {
		t.Fatalf("search 38 --all: code=%d err=%q", code, err)
	}
	if !strings.Contains(out, "0.038") {
		t.Fatalf("search 38 --all should allow description substring matches:\n%s", out)
	}
}

func TestDiscoveryAliases(t *testing.T) {
	tests := []struct {
		args  []string
		wants []string
	}{
		{[]string{"aliases", "mph"}, []string{"kind: unit", "mile per hour", "dimensions: m/s"}},
		{[]string{"aliases", "Rj"}, []string{"Jupiter mean radius", "jupiterradius", "approximate"}},
		{[]string{"aliases", "flour"}, []string{"kind: ingredient", "canonical: all-purpose-flour", "density: 508 kg/m^3"}},
		{[]string{"aliases", "--all"}, []string{"unit", "mph", "ingredient", "all-purpose-flour"}},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 {
			t.Fatalf("%v: code=%d err=%q", tt.args, code, err)
		}
		for _, want := range tt.wants {
			if !strings.Contains(out, want) {
				t.Errorf("%v: output missing %q:\n%s", tt.args, want, out)
			}
		}
	}

	code, _, err := run("aliases", "not-a-known-entry")
	if code == 0 || !strings.Contains(err, "no aliases or catalog entry") {
		t.Fatalf("unknown alias: code=%d err=%q", code, err)
	}
}

func TestDiscoveryJSON(t *testing.T) {
	code, out, err := run("--json", "search", "jupiter")
	if code != 0 {
		t.Fatalf("search json: code=%d err=%q", code, err)
	}
	var search struct {
		Command string
		Query   string
		Results []struct {
			Kind string
			Key  string
		}
	}
	if err := json.Unmarshal([]byte(out), &search); err != nil {
		t.Fatal(err)
	}
	if search.Command != "search" || search.Query != "jupiter" || !hasDiscoveryKey(search.Results, "unit", "Rj") {
		t.Fatalf("unexpected search json: %+v", search)
	}

	code, out, err = run("aliases", "--json", "Rj")
	if code != 0 {
		t.Fatalf("aliases json: code=%d err=%q", code, err)
	}
	var aliases struct {
		Command string
		Query   string
		Matches []struct {
			Kind      string
			Key       string
			Dimension string
		}
	}
	if err := json.Unmarshal([]byte(out), &aliases); err != nil {
		t.Fatal(err)
	}
	if aliases.Command != "aliases" || aliases.Query != "Rj" || !hasAliasKey(aliases.Matches, "unit", "Rj") {
		t.Fatalf("unexpected aliases json: %+v", aliases)
	}
}

func hasDiscoveryKey(results []struct {
	Kind string
	Key  string
}, kind, key string) bool {
	for _, result := range results {
		if result.Kind == kind && result.Key == key {
			return true
		}
	}
	return false
}

func hasAliasKey(results []struct {
	Kind      string
	Key       string
	Dimension string
}, kind, key string) bool {
	for _, result := range results {
		if result.Kind == kind && result.Key == key {
			return true
		}
	}
	return false
}

func TestCompareCLI(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"compare", "38in", "Rj"}, "38 in = approximately 1.380612493e-08 Rj\n"},
		{[]string{"compare", "38in", "banana", "smoot", "Rj"}, "38 in is:\n  approximately 5.362222222 banana\n  approximately 0.5671641791 smoot\n  approximately 1.380612493e-08 Rj\n"},
		{[]string{"compare", "38", "in", "banana", "smoot", "Rj"}, "38 in is:\n  approximately 5.362222222 banana\n  approximately 0.5671641791 smoot\n  approximately 1.380612493e-08 Rj\n"},
		{[]string{"compare", "1earthcircumference", "marathon"}, "1 earthcircumference = approximately 949.7574831 marathon\n"},
		{[]string{"compare", "1olympicpool", "cup", "gal"}, "1 olympicpool is:\n  approximately 10566882.09 cup\n  approximately 660430.1309 gal\n"},
		{[]string{"compare", "60mph", "m/s", "km/h", "ft/s"}, "60 mph is:\n  26.8224 m/s\n  96.56064 km/h\n  88 ft/s\n"},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 || out != tt.want {
			t.Errorf("%v: code=%d out=%q err=%q", tt.args, code, out, err)
		}
	}
}

func TestComparePresets(t *testing.T) {
	code, out, err := run("compare", "38in", "--fun")
	if code != 0 {
		t.Fatalf("fun: code=%d err=%q", code, err)
	}
	for _, want := range []string{"banana", "smoot", "footballfield", "marathon", "earthcircumference"} {
		if !strings.Contains(out, want) {
			t.Errorf("--fun output missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "olympicpool") {
		t.Errorf("--fun length output should skip volume-only olympicpool:\n%s", out)
	}

	code, out, err = run("compare", "38in", "--astronomical")
	if code != 0 {
		t.Fatalf("astronomical: code=%d err=%q", code, err)
	}
	for _, want := range []string{"Re", "Rj", "Rsun", "LD", "au", "ls", "lightmin", "lh", "ld", "ly", "pc"} {
		if !strings.Contains(out, want) {
			t.Errorf("--astronomical output missing %q:\n%s", want, out)
		}
	}

	code, out, err = run("compare", "1olympicpool", "--fun")
	if code != 0 || !strings.Contains(out, "olympicpool") || strings.Contains(out, "banana") {
		t.Fatalf("volume preset skip: code=%d out=%q err=%q", code, out, err)
	}
}

func TestCompareErrorsAndHelp(t *testing.T) {
	code, _, err := run("compare", "38in", "kg")
	if code == 0 || !strings.Contains(err, "kg: cannot convert") {
		t.Fatalf("incompatible target: code=%d err=%q", code, err)
	}
	code, _, err = run("compare", "38in", "kg", "--fun")
	if code == 0 || !strings.Contains(err, "kg: cannot convert") {
		t.Fatalf("explicit incompatible target with preset: code=%d err=%q", code, err)
	}
	code, _, err = run("compare", "1kg", "--fun")
	if code == 0 || !strings.Contains(err, "compatible preset") {
		t.Fatalf("incompatible preset: code=%d err=%q", code, err)
	}
	code, out, err := run("compare", "--help")
	if code != 0 || err != "" || !strings.Contains(out, "convunits compare <valueunit>") {
		t.Fatalf("help: code=%d out=%q err=%q", code, out, err)
	}
}

func TestCompareJSON(t *testing.T) {
	code, out, err := run("--json", "compare", "38in", "banana", "smoot", "Rj")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Command string
		Input   struct {
			Value float64
			Unit  string
		}
		Outputs []struct {
			Value       float64
			Unit        string
			Approximate bool
		}
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Command != "compare" || got.Input.Value != 38 || got.Input.Unit != "in" || len(got.Outputs) != 3 {
		t.Fatalf("%+v", got)
	}
	if got.Outputs[0].Unit != "banana" || !got.Outputs[0].Approximate || math.Abs(got.Outputs[0].Value-5.362222222) > 1e-9 {
		t.Fatalf("%+v", got.Outputs[0])
	}
}

func TestEvalCLI(t *testing.T) {
	tests := []struct {
		expression string
		want       string
	}{
		{"38in / Rj", "approximately 1.380612493e-08\n"},
		{"1Rsun / 38in", "approximately 720783257.4\n"},
		{"1olympicpool / 1cup", "approximately 10566882.09\n"},
		{"2*pi*1Re -> km", "approximately 40030.22888 km\n"},
		{"0.5 * 1500kg * (60mph)^2 -> kWh", "0.1498835712 kWh\n"},
		{"1kg * 9.80665m/s^2 -> N", "9.80665 N\n"},
		{"1m + 1ft -> m", "1.3048 m\n"},
		{"(10m)^2 -> m^2", "100 m^2\n"},
		{"1kg*c^2 -> J", "8.987551787e+16 J\n"},
		{"1kg * c^2 -> J", "8.987551787e+16 J\n"},
		{"1N*1m -> J", "1 J\n"},
		{"1N * 1m -> J", "1 J\n"},
		{"60mph*1h -> mi", "60 mi\n"},
		{"60mph * 1h -> mi", "60 mi\n"},
		{"1kg/m^3 -> kg/m^3", "1 kg/m^3\n"},
		{"1kg*m/s^2 -> N", "1 N\n"},
		{"1C -> A*s", "1 A*s\n"},
		{"1F -> A^2*s^4/(kg*m^2)", "1 A^2*s^4/(kg*m^2)\n"},
		{"1K + 2K -> K", "3 K\n"},
	}
	for _, tt := range tests {
		code, out, err := run("eval", tt.expression)
		if code != 0 || out != tt.want {
			t.Errorf("%q: code=%d out=%q err=%q", tt.expression, code, out, err)
		}
	}
}

func TestEvalErrorsAndHelp(t *testing.T) {
	tests := []struct {
		expression string
		contains   string
	}{
		{"1m + 1kg", "incompatible dimensions"},
		{"1m -> kg", "incompatible dimensions"},
		{"1 / 0", "division by zero"},
		{"1m / 0", "division by zero"},
		{"1m / 0s", "division by zero"},
		{"100C + 1K", "eval does not treat Celsius/Fahrenheit as ordinary scalar units"},
		{"1F -> C", "eval does not treat Celsius/Fahrenheit as ordinary scalar units"},
	}
	for _, tt := range tests {
		code, _, err := run("eval", tt.expression)
		if code == 0 || !strings.Contains(err, tt.contains) {
			t.Errorf("%q: code=%d err=%q", tt.expression, code, err)
		}
	}
	code, out, err := run("eval", "--help")
	if code != 0 || err != "" || !strings.Contains(out, "convunits eval '<expression>'") {
		t.Fatalf("eval help: code=%d out=%q err=%q", code, out, err)
	}
}

func TestEvalJSON(t *testing.T) {
	code, out, err := run("--json", "eval", "38in / Rj")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Command    string
		Expression string
		Output     struct {
			Value       float64
			Dimension   string
			Approximate bool
		}
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Command != "eval" || got.Expression != "38in / Rj" || got.Output.Dimension != "1" || !got.Output.Approximate || math.Abs(got.Output.Value-1.380612493e-08) > 1e-18 {
		t.Fatalf("%+v", got)
	}

	code, out, err = run("--json", "eval", "2*pi*1Re -> km")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var converted struct {
		Output struct {
			Value       float64
			Unit        string
			Approximate bool
		}
	}
	if err := json.Unmarshal([]byte(out), &converted); err != nil {
		t.Fatal(err)
	}
	if converted.Output.Unit != "km" || !converted.Output.Approximate || math.Abs(converted.Output.Value-40030.22888) > 1e-5 {
		t.Fatalf("%+v", converted.Output)
	}
}

func TestExplainCLI(t *testing.T) {
	tests := []struct {
		args     []string
		contains []string
	}{
		{[]string{"explain", "60mph", "m/s"}, []string{"60 mph -> m/s", "1 mph = 0.44704 m/s", "Result:\n  26.8224 m/s"}},
		{[]string{"explain", "10kg", "lb"}, []string{"10 kg -> lb", "1 lb = 0.45359237 kg", "22.04622622 lb"}},
		{[]string{"explain", "1N", "kg*m/s^2"}, []string{"1 N -> kg*m/s^2", "N has dimensions kg*m/s^2", "Result:\n  1 kg*m/s^2"}},
		{[]string{"explain", "38in", "Rj"}, []string{"38 in -> Rj", "1 Rj = 69911000 m approximately", "approximately 1.380612493e-08 Rj"}},
		{[]string{"explain", "2*pi*1Re -> km"}, []string{"2*pi*1Re -> km", "pi = 3.141592653589793", "1 Re = 6371008.8 m approximately", "approximately 40030.22888 km"}},
		{[]string{"explain", "0.5 * 1500kg * (60mph)^2 -> kWh"}, []string{"60 mph = 26.8224 m/s", "0.5 * 1500 kg", "1 kWh = 3600000 J", "0.1498835712 kWh"}},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 {
			t.Fatalf("%v: code=%d err=%q", tt.args, code, err)
		}
		for _, want := range tt.contains {
			if !strings.Contains(out, want) {
				t.Errorf("%v: output missing %q:\n%s", tt.args, want, out)
			}
		}
	}
}

func TestExplainJSON(t *testing.T) {
	code, out, err := run("--json", "explain", "60mph", "m/s")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var normal struct {
		Command string
		Input   struct {
			Value float64
			Unit  string
		}
		Output struct {
			Value       float64
			Unit        string
			Approximate bool
		}
		Steps      []string
		Dimensions struct {
			Input, Output string
		}
	}
	if err := json.Unmarshal([]byte(out), &normal); err != nil {
		t.Fatal(err)
	}
	if normal.Command != "explain" || normal.Input.Value != 60 || normal.Input.Unit != "mph" || normal.Output.Value != 26.8224 || normal.Output.Unit != "m/s" || normal.Output.Approximate {
		t.Fatalf("%+v", normal)
	}
	if normal.Dimensions.Input != "m/s" || len(normal.Steps) == 0 {
		t.Fatalf("%+v", normal)
	}

	code, out, err = run("--json", "explain", "2*pi*1Re -> km")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var eval struct {
		Command    string
		Expression string
		Output     struct {
			Value       float64
			Unit        string
			Approximate bool
		}
		Steps []string
	}
	if err := json.Unmarshal([]byte(out), &eval); err != nil {
		t.Fatal(err)
	}
	if eval.Command != "explain" || eval.Expression != "2*pi*1Re -> km" || eval.Output.Unit != "km" || !eval.Output.Approximate || math.Abs(eval.Output.Value-40030.22888) > 1e-5 || len(eval.Steps) == 0 {
		t.Fatalf("%+v", eval)
	}
}

func TestExplainErrorsAndHelp(t *testing.T) {
	tests := []struct {
		args     []string
		contains string
	}{
		{[]string{"explain", "1m", "kg"}, "incompatible dimensions"},
		{[]string{"explain", "1m + 1kg -> m"}, "incompatible dimensions"},
		{[]string{"explain", "recipe", "1cup", "flour", "g"}, "explain does not support recipe yet"},
		{[]string{"explain", "scale", "7", "pH", "mol/L"}, "explain does not support scale yet"},
	}
	for _, tt := range tests {
		code, _, err := run(tt.args...)
		if code == 0 || !strings.Contains(err, tt.contains) {
			t.Errorf("%v: code=%d err=%q", tt.args, code, err)
		}
	}
	code, out, err := run("explain", "--help")
	if code != 0 || err != "" || !strings.Contains(out, "convunits explain <valueunit>") {
		t.Fatalf("explain help: code=%d out=%q err=%q", code, out, err)
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

func TestRecipeCLI(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{[]string{"recipe", "1cup", "flour", "g"}, "approximately 120.1868241 g all-purpose flour\n"},
		{[]string{"recipe", "100g", "flour", "cup"}, "approximately 0.8320379602 cup all-purpose flour\n"},
		{[]string{"recipe", "2tbsp", "butter", "g"}, "approximately 28.36101485 g butter\n"},
		{[]string{"recipe", "1cup", "sugar", "g"}, "approximately 199.9170598 g granulated sugar\n"},
		{[]string{"recipe", "1cup", "honey", "oz"}, "approximately 11.85047432 oz honey\n"},
		{[]string{"recipe", "500ml", "water", "lb"}, "approximately 1.102311311 lb water\n"},
		{[]string{"recipe", "1cup", "rice", "g"}, "approximately 186.9047068 g uncooked rice\n"},
		{[]string{"recipe", "100g", "flour", "oz"}, "approximately 3.527396195 oz all-purpose flour\n"},
		{[]string{"recipe", "1cup", "flour", "tbsp"}, "approximately 16 tbsp all-purpose flour\n"},
		{[]string{"recipe", "1", "cup", "flour", "g"}, "approximately 120.1868241 g all-purpose flour\n"},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 || out != tt.want {
			t.Errorf("%v: code=%d out=%q err=%q", tt.args, code, out, err)
		}
	}
}

func TestRecipeErrorsAndListings(t *testing.T) {
	tests := []struct {
		args     []string
		contains string
	}{
		{[]string{"recipe", "1cup", "moon-dust", "g"}, "unknown ingredient"},
		{[]string{"recipe", "1m", "flour", "g"}, "not a mass or volume"},
		{[]string{"recipe", "1cup", "flour", "m"}, "not a mass or volume"},
	}
	for _, tt := range tests {
		code, _, err := run(tt.args...)
		if code == 0 || !strings.Contains(err, tt.contains) {
			t.Errorf("%v: code=%d err=%q", tt.args, code, err)
		}
	}
	code, out, err := run("recipe", "ingredients")
	if code != 0 || !strings.Contains(out, "all-purpose-flour") || !strings.Contains(out, "flour") || !strings.Contains(out, "kosher-salt") {
		t.Fatalf("ingredients: code=%d out=%q err=%q", code, out, err)
	}
	code, out, err = run("recipe", "ingredients", "baking")
	if code != 0 || !strings.Contains(out, "all-purpose-flour") || strings.Contains(out, "olive-oil") {
		t.Fatalf("ingredients baking: code=%d out=%q err=%q", code, out, err)
	}
	code, out, err = run("recipe", "--help")
	if code != 0 || err != "" || !strings.Contains(out, "convunits recipe <amount><unit>") {
		t.Fatalf("recipe help: code=%d out=%q err=%q", code, out, err)
	}
}

func TestRecipeJSON(t *testing.T) {
	code, out, err := run("--json", "recipe", "1cup", "flour", "g")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Command string
		Input   struct {
			Value      float64
			Unit       string
			Ingredient string
		}
		Output struct {
			Value       float64
			Unit        string
			Ingredient  string
			Approximate bool
		}
		Density struct {
			Value float64
			Unit  string
		}
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Command != "recipe" || got.Input.Value != 1 || got.Input.Unit != "cup" || got.Input.Ingredient != "flour" {
		t.Fatalf("%+v", got)
	}
	if got.Output.Unit != "g" || got.Output.Ingredient != "all-purpose flour" || !got.Output.Approximate || math.Abs(got.Output.Value-120.186824142) > 1e-9 {
		t.Fatalf("%+v", got.Output)
	}
	if got.Density.Value != 508 || got.Density.Unit != "kg/m^3" {
		t.Fatalf("%+v", got.Density)
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

func TestPaperJSON(t *testing.T) {
	for _, args := range [][]string{
		{"--json", "paper", "a4", "mm"},
		{"paper", "--json", "a4", "mm"},
	} {
		code, out, err := run(args...)
		if code != 0 {
			t.Fatalf("%v: code=%d err=%q", args, code, err)
		}
		var got struct {
			Command string
			Input   struct {
				Size string
			}
			Output struct {
				Width, Height float64
				Unit          string
			}
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("%v: %v\n%s", args, err, out)
		}
		if got.Command != "paper" || got.Input.Size != "a4" || got.Output.Width != 210 || got.Output.Height != 297 || got.Output.Unit != "mm" {
			t.Fatalf("%v: %+v", args, got)
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

func TestWireJSON(t *testing.T) {
	code, out, err := run("--json", "wire", "12awg", "mm")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Command string
		Input   struct {
			Gauge, System string
		}
		Output struct {
			Value       float64
			Unit        string
			Quantity    string
			Approximate bool
		}
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Command != "wire" || got.Input.Gauge != "12awg" || got.Input.System != "AWG" || got.Output.Unit != "mm" || got.Output.Quantity != "diameter" || !got.Output.Approximate || math.Abs(got.Output.Value-2.052525388) > 1e-9 {
		t.Fatalf("%+v", got)
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

func TestDrillJSON(t *testing.T) {
	code, out, err := run("--json", "drill", "#7", "mm")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Command string
		Input   struct {
			Size string
		}
		Output struct {
			Value       float64
			Unit        string
			Quantity    string
			Approximate bool
		}
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Command != "drill" || got.Input.Size != "#7" || got.Output.Unit != "mm" || got.Output.Quantity != "diameter" || !got.Output.Approximate || math.Abs(got.Output.Value-5.1054) > 1e-12 {
		t.Fatalf("%+v", got)
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

func TestSieveJSON(t *testing.T) {
	code, out, err := run("--json", "sieve", "#40", "mm")
	if code != 0 {
		t.Fatalf("code=%d err=%q", code, err)
	}
	var got struct {
		Command string
		Input   struct {
			Size string
		}
		Output struct {
			Value       float64
			Unit        string
			Quantity    string
			Approximate bool
		}
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Command != "sieve" || got.Input.Size != "#40" || got.Output.Unit != "mm" || got.Output.Quantity != "opening" || !got.Output.Approximate || math.Abs(got.Output.Value-0.425) > 1e-12 {
		t.Fatalf("%+v", got)
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
	if code != 0 || !strings.Contains(out, "escape-velocity") || !strings.Contains(out, "bmi") || !strings.Contains(out, "schwarzschild-radius") || !strings.Contains(out, "flow-rate") {
		t.Fatalf("code=%d out=%q err=%q", code, out, err)
	}
}

func TestFormulaJSON(t *testing.T) {
	tests := []struct {
		args      []string
		formula   string
		unit      string
		want      float64
		tolerance float64
		inputs    map[string]string
	}{
		{
			args:      []string{"--json", "formula", "bmi", "--mass", "180lb", "--height", "6ft", "bmi"},
			formula:   "bmi",
			unit:      "bmi",
			want:      24.41213818,
			tolerance: 1e-8,
			inputs:    map[string]string{"mass": "180lb", "height": "6ft"},
		},
		{
			args:      []string{"formula", "--json", "bmi", "--mass", "180lb", "--height", "6ft", "bmi"},
			formula:   "bmi",
			unit:      "bmi",
			want:      24.41213818,
			tolerance: 1e-8,
			inputs:    map[string]string{"mass": "180lb", "height": "6ft"},
		},
		{
			args:      []string{"--json", "formula", "schwarzschild-radius", "--mass", "1Mearth", "mm"},
			formula:   "schwarzschild-radius",
			unit:      "mm",
			want:      8.87,
			tolerance: .01,
			inputs:    map[string]string{"mass": "1Mearth"},
		},
		{
			args:      []string{"formula", "--json", "power", "--energy", "1kWh", "--time", "1h", "W"},
			formula:   "power",
			unit:      "W",
			want:      1000,
			tolerance: 1e-9,
			inputs:    map[string]string{"energy": "1kWh", "time": "1h"},
		},
	}
	for _, tt := range tests {
		code, out, err := run(tt.args...)
		if code != 0 {
			t.Fatalf("%v: code=%d err=%q", tt.args, code, err)
		}
		var got struct {
			Command string
			Formula string
			Inputs  map[string]string
			Output  struct {
				Value float64
				Unit  string
			}
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("%v: %v\n%s", tt.args, err, out)
		}
		if got.Command != "formula" || got.Formula != tt.formula || got.Output.Unit != tt.unit || math.Abs(got.Output.Value-tt.want) > tt.tolerance {
			t.Fatalf("%v: %+v", tt.args, got)
		}
		for key, want := range tt.inputs {
			if got.Inputs[key] != want {
				t.Fatalf("%v: input %s = %q, want %q", tt.args, key, got.Inputs[key], want)
			}
		}
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
