package wire

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ParseGauge(text string) (int, error) {
	s := strings.ToLower(strings.TrimSpace(text))
	if strings.HasPrefix(s, "awg") {
		s = strings.TrimPrefix(s, "awg")
	}
	if strings.HasSuffix(s, "awg") {
		s = strings.TrimSuffix(s, "awg")
	}
	switch s {
	case "0000":
		return -3, nil
	case "000":
		return -2, nil
	case "00":
		return -1, nil
	case "0":
		return 0, nil
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("invalid AWG gauge %q: zero gauges use 0, 00, 000, or 0000", text)
	}
	gauge, err := strconv.Atoi(s)
	if err != nil || gauge < -3 || gauge > 40 {
		return 0, fmt.Errorf("invalid AWG gauge %q: expected -3 through 40, 0000awg through 0awg, awgN, or Nawg", text)
	}
	return gauge, nil
}

func DiameterMeters(gauge int) (float64, error) {
	if gauge < -3 || gauge > 40 {
		return 0, fmt.Errorf("invalid AWG gauge %d: expected -3 through 40", gauge)
	}
	diameterIn := 0.005 * math.Pow(92, float64(36-gauge)/39)
	return diameterIn * 0.0254, nil
}

func SyntaxHelp() string {
	return "AWG gauges -3 through 40; accepted forms: 12, 12awg, awg12, 0000awg, 000awg, 00awg, 0awg"
}
