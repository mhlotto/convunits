package units

import (
	"math"
	"strings"
	"testing"
)

func TestExpandedCatalogConversions(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		value    float64
		from, to string
		want     float64
		relTol   float64
	}{
		{1, "ls", "m", 299792458, 0},
		{1, "au", "ls", 149597870700.0 / 299792458.0, 1e-14},
		{1, "Rsun", "km", 695700, 0},
		{1, "Mearth", "kg", 5.9722e24, 0},
		{1, "cubit", "in", 18, 1e-14},
		{1, "royalcubit", "m", 0.5236, 0},
		{1, "stadion", "m", 185, 0},
		{1, "barn", "m^2", 1e-28, 0},
		{1, "Gy", "J/kg", 1, 0},
		{1, "Ci", "Bq", 3.7e10, 0},
		{1, "angstrom", "nm", 0.1, 1e-14},
		{1, "Da", "kg", 1.66053906660e-27, 0},
		{1, "hartree", "eV", 4.3597447222060e-18 / 1.602176634e-19, 1e-14},
		{1, "Molar", "mol/L", 1, 0},
		{1, "point", "in", 1.0 / 72, 1e-14},
		{1, "csspx", "in", 1.0 / 96, 1e-14},
		{1, "PiB", "B", 1125899906842624, 0},
		{1, "nibble", "b", 4, 0},
		{120, "bpm", "Hz", 2, 0},
		{1, "smoot", "m", 1.7018, 0},
		{1, "marathon", "km", 42.195, 1e-14},
		{1, "olympicpool", "L", 2500000, 0},
		{1, "teaspoon-metric", "mL", 5, 1e-14},
		{8, "Kbps", "kB/s", 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.from+"_to_"+tt.to, func(t *testing.T) {
			got, err := r.Convert(tt.value, tt.from, tt.to)
			if err != nil {
				t.Fatal(err)
			}
			tolerance := tt.relTol * math.Abs(tt.want)
			if tolerance == 0 {
				tolerance = math.SmallestNonzeroFloat64
			}
			if math.Abs(got.Value-tt.want) > tolerance {
				t.Fatalf("got %.17g, want %.17g (tolerance %g)", got.Value, tt.want, tolerance)
			}
		})
	}
}

func TestNonlinearAndContextualScalesRemainUnsupported(t *testing.T) {
	r := NewRegistry()
	for _, name := range []string{"beaufort", "pH", "redshift", "redshift-distance-placeholder", "magnitude", "semitone", "octave", "epoch"} {
		t.Run(name, func(t *testing.T) {
			_, err := r.Convert(1, name, "m")
			if err == nil || !strings.Contains(err.Error(), "unknown unit") {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestCollisionSensitiveSymbols(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		value    float64
		from, to string
		want     float64
	}{
		{100, "C", "F", 212},
		{1, "C", "A*s", 1},
		{1, "F", "A^2*s^4/(kg*m^2)", 1},
		{1, "rad", "deg", 180 / math.Pi},
		{1, "rad-dose", "Gy", 0.01},
		{1, "lm", "cd", 1},
		{1, "lightmin", "km", speedOfLight * 60 / 1000},
		{1, "pt", "L", 0.473176473},
		{1, "point", "in", 1.0 / 72},
		{1, "Molar", "mol/L", 1},
		{1, "M", "mol/L", 1},
		{1, "u", "kg", 1.66053906660e-27},
		{1, "um", "m", 1e-6},
		{1, "T", "kg/(A*s^2)", 1},
		{1, "TB", "B", 1e12},
	}
	for _, tt := range tests {
		got, err := r.Convert(tt.value, tt.from, tt.to)
		if err != nil {
			t.Fatalf("%s to %s: %v", tt.from, tt.to, err)
		}
		if math.Abs(got.Value-tt.want) > math.Max(1e-12, math.Abs(tt.want)*1e-12) {
			t.Fatalf("%s to %s: got %g, want %g", tt.from, tt.to, got.Value, tt.want)
		}
	}
}

func TestUnresolvedAmbiguousUnitsReturnClearErrors(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		from, to string
		want     string
	}{
		{"C", "C", "ambiguous conversion"},
		{"F", "F", "ambiguous conversion"},
		{"C", "s", "ambiguous input unit \"C\""},
		{"kg", "F", "ambiguous output unit \"F\""},
	}
	for _, tt := range tests {
		_, err := r.Convert(1, tt.from, tt.to)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s to %s: got %v, want error containing %q", tt.from, tt.to, err, tt.want)
		}
	}
}

func TestOnlyCAndFHaveDuplicateSymbols(t *testing.T) {
	r := NewRegistry()
	bySymbol := make(map[string][]string)
	for _, unit := range r.units {
		bySymbol[unit.Symbol] = append(bySymbol[unit.Symbol], unit.Name)
	}
	for symbol, names := range bySymbol {
		switch symbol {
		case "C", "F":
			if len(names) != 2 {
				t.Errorf("intentional duplicate symbol %q has definitions %v", symbol, names)
			}
		default:
			if len(names) != 1 {
				t.Errorf("duplicate symbol %q: %v", symbol, names)
			}
		}
	}
}

func TestExpandedCatalogHasNoNewAmbiguousAliases(t *testing.T) {
	r := NewRegistry()
	for alias, entries := range r.aliases {
		if alias == "C" || alias == "F" {
			if len(entries) != 2 {
				t.Errorf("intentional alias %q has %d entries, want 2", alias, len(entries))
			}
			continue
		}
		if len(entries) != 1 {
			names := make([]string, 0, len(entries))
			for _, entry := range entries {
				names = append(names, entry.Name)
			}
			t.Errorf("alias %q is ambiguous: %v", alias, names)
		}
	}
}

func TestApproximateMetadata(t *testing.T) {
	r := NewRegistry()
	units, err := r.LookupAll("cubit")
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 1 || !units[0].Approximate || !strings.Contains(units[0].Note, "18 inches") {
		t.Fatalf("unexpected cubit metadata: %+v", units)
	}
}

func TestExpandedCatalogSymbolsAndAliasesAreRegistered(t *testing.T) {
	r := NewRegistry()
	names := []string{
		"ld", "lightday", "lh", "lighthour", "lightmin", "lightminute", "ls", "lightsecond",
		"earthradius", "Re", "jupiterradius", "Rj", "solarradius", "Rsun", "lunardistance", "LD",
		"Mearth", "earthmass", "Mjup", "jupitermass", "Msun", "solarmass",
		"siderealday", "siderealyear", "julianyear", "galacticyear", "cosmicyear",
		"c0", "lightspeed", "earthorbitalvelocity", "escapeearth",
		"cubit", "royalcubit", "span", "handbreadth", "fingerbreadth", "pace", "romanfoot", "romanmile",
		"stadion", "parasang", "league", "fathom", "cable", "shot", "nleague", "chain", "rod", "furlong", "hand",
		"rood", "section", "township", "barn", "shed", "outhouse",
		"ka", "Ma", "Ga", "eon", "GPa", "Mbar", "kbar",
		"u", "Da", "me", "mp", "mn", "angstrom", "Å", "Aang", "bohr", "a0", "fermi", "fm",
		"hartree", "Eh", "rydberg", "Ry", "erg",
		"Gy", "rad-dose", "rd", "Sv", "rem", "Bq", "Ci",
		"Molar", "M", "mMolar", "mM", "uMolar", "uM", "nMolar", "nM", "ppm-water",
		"point", "typographicpoint", "pica", "twip", "csspx", "didot", "cicero",
		"PiB", "EiB", "PB", "EB", "Pb", "Eb", "nibble", "word16", "word32", "word64", "block512", "page4k", "baud",
		"bpm", "frame24", "frame25", "frame30", "frame60", "sample44100", "sample48000",
		"banana", "smoot", "footballfield", "footballfield-us", "marathon", "earthcircumference", "olympicpool",
		"teaspoon-metric", "metricteaspoon", "Kbps",
	}
	for _, name := range names {
		if _, err := r.LookupAll(name); err != nil {
			t.Errorf("%q is not registered: %v", name, err)
		}
	}
}
