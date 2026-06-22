package units

import "testing"

func TestParser(t *testing.T) {
	r := NewRegistry()
	tests := []struct {
		expr string
		dim  Dimension
		mul  float64
	}{
		{"kg*(m/s^2)", Dimension{Length: 1, Mass: 1, Time: -2}, 1},
		{"km/h", Dimension{Length: 1, Time: -1}, 1000.0 / 3600},
		{"MB/s", Dimension{Information: 1, Time: -1}, 1e6},
		{"m^3", Dimension{Length: 3}, 1},
		{"s^-2", Dimension{Time: -2}, 1},
	}
	for _, tt := range tests {
		got, err := r.ParseCandidates(tt.expr)
		if err != nil {
			t.Errorf("%s: %v", tt.expr, err)
			continue
		}
		if len(got) != 1 || got[0].Dimension != tt.dim || got[0].Multiplier != tt.mul {
			t.Errorf("%s: %#v", tt.expr, got)
		}
	}
}

func TestUnknownUnit(t *testing.T) {
	if _, err := NewRegistry().ParseCandidates("wat"); err == nil {
		t.Fatal("expected error")
	}
}
