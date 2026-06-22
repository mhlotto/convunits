package wire

import (
	"math"
	"strings"
	"testing"
)

func TestGaugeParsingAndDiameter(t *testing.T) {
	tests := []struct {
		text  string
		gauge int
	}{
		{"12", 12}, {"12awg", 12}, {"awg12", 12}, {"0000awg", -3}, {"000awg", -2}, {"00awg", -1}, {"0awg", 0}, {"-3awg", -3},
	}
	for _, tt := range tests {
		gauge, err := ParseGauge(tt.text)
		if err != nil || gauge != tt.gauge {
			t.Errorf("%s: gauge=%d err=%v", tt.text, gauge, err)
		}
	}
	d12, _ := DiameterMeters(12)
	if math.Abs(d12*1000-2.052525388) > 1e-9 {
		t.Fatalf("12 AWG: %g mm", d12*1000)
	}
	d0000, _ := DiameterMeters(-3)
	if math.Abs(d0000*1000-11.684) > 1e-9 {
		t.Fatalf("0000 AWG: %g mm", d0000*1000)
	}
	d40, _ := DiameterMeters(40)
	if math.Abs(d40*1000-0.07987108513) > 1e-10 {
		t.Fatalf("40 AWG: %g mm", d40*1000)
	}
}

func TestInvalidGauge(t *testing.T) {
	for _, value := range []string{"wat", "41", "00000awg"} {
		if _, err := ParseGauge(value); err == nil || !strings.Contains(err.Error(), "invalid AWG") {
			t.Errorf("%s: %v", value, err)
		}
	}
}
