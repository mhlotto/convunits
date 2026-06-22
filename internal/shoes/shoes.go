package shoes

import (
	"fmt"
	"math"
	"sort"
)

var supported = map[string]bool{
	"us-men": true, "us-women": true, "uk-adult": true, "eu": true, "mondo": true, "jp": true,
}

func Systems() []string {
	out := make([]string, 0, len(supported))
	for system := range supported {
		out = append(out, system)
	}
	sort.Strings(out)
	return out
}

// FootLengthMeters returns an approximate foot length, not a fit recommendation.
func FootLengthMeters(system string, size float64) (float64, error) {
	if system == "us" {
		return 0, fmt.Errorf("ambiguous shoe system %q; use us-men or us-women", system)
	}
	if system == "uk" {
		return 0, fmt.Errorf("ambiguous shoe system %q; use uk-adult", system)
	}
	if system == "kids" || system == "us-kids" || system == "uk-kids" {
		return 0, fmt.Errorf("unsupported children's shoe system %q", system)
	}
	if !supported[system] {
		return 0, fmt.Errorf("unknown shoe system %q", system)
	}
	if math.IsNaN(size) || math.IsInf(size, 0) || size <= 0 {
		return 0, fmt.Errorf("invalid shoe size %g", size)
	}

	var meters float64
	switch system {
	case "us-men":
		if !halfStep(size) {
			return 0, fmt.Errorf("us-men size must use whole or half sizes")
		}
		meters = ((size + 22) / 3) * 0.0254
	case "us-women":
		if !halfStep(size) {
			return 0, fmt.Errorf("us-women size must use whole or half sizes")
		}
		meters = ((size + 21) / 3) * 0.0254
	case "uk-adult":
		if !halfStep(size) {
			return 0, fmt.Errorf("uk-adult size must use whole or half sizes")
		}
		meters = ((size + 23) / 3) * 0.0254
	case "eu":
		if !halfStep(size) {
			return 0, fmt.Errorf("eu size must use whole or half sizes")
		}
		meters = (size*2.0/3 - 1.5) / 100
	case "mondo":
		meters = size / 1000
	case "jp":
		meters = size / 100
	}
	if meters <= 0 {
		return 0, fmt.Errorf("shoe size %g produces a non-positive foot length", size)
	}
	return meters, nil
}

func halfStep(value float64) bool { return math.Abs(value*2-math.Round(value*2)) < 1e-9 }
