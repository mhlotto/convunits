package scales

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type Scale struct {
	Symbol, Name, Category string
	Aliases                []string
	Note, Base             string
	kind                   scaleKind
}

type scaleKind int

const (
	kindDB scaleKind = iota
	kindBel
	kindPowerRatio
	kindAmplitudeRatio
	kindPH
	kindHydrogenConcentration
	kindMagnitude
	kindBrightnessRatio
	kindBeaufort
	kindMetersPerSecond
	kindKilometersPerHour
	kindMilesPerHour
	kindKnot
	kindAWG
	kindDiameterMM
	kindDiameterIn
)

type Result struct {
	Value    *float64
	Min, Max *float64
	Unit     string
}

func Scalar(value float64, unit string) Result { return Result{Value: &value, Unit: unit} }

func BoundedRange(min, max float64, unit string) Result {
	return Result{Min: &min, Max: &max, Unit: unit}
}

type Registry struct {
	scales  []*Scale
	aliases map[string]*Scale
}

func NewRegistry() *Registry {
	r := &Registry{aliases: make(map[string]*Scale)}
	r.registerCatalog()
	return r
}

func (r *Registry) add(scale Scale) {
	p := &scale
	r.scales = append(r.scales, p)
	seen := make(map[string]bool)
	for _, alias := range append([]string{scale.Symbol, scale.Name}, scale.Aliases...) {
		if alias == "" || seen[alias] {
			continue
		}
		if existing := r.aliases[alias]; existing != nil {
			panic(fmt.Sprintf("scale alias %q conflicts with %s", alias, existing.Name))
		}
		r.aliases[alias] = p
		seen[alias] = true
	}
}

func (r *Registry) registerCatalog() {
	for _, scale := range []Scale{
		{Symbol: "dB", Name: "decibel", Category: "ratio", Base: "dB", kind: kindDB},
		{Symbol: "bel", Name: "bel", Category: "ratio", Base: "dB", kind: kindBel, Note: "1 bel = 10 dB"},
		{Symbol: "power-ratio", Name: "power ratio", Category: "ratio", Base: "dB", kind: kindPowerRatio, Note: "must be positive"},
		{Symbol: "amplitude-ratio", Name: "amplitude ratio", Category: "ratio", Base: "dB", kind: kindAmplitudeRatio, Note: "must be positive"},
		{Symbol: "pH", Name: "pH", Category: "chemistry", Base: "mol/L H+", kind: kindPH},
		{Symbol: "H+", Name: "hydrogen ion concentration", Category: "chemistry", Base: "mol/L H+", kind: kindHydrogenConcentration, Note: "specifically hydrogen ion concentration"},
		{Symbol: "mol/L", Name: "moles per liter H+", Category: "chemistry", Base: "mol/L H+", kind: kindHydrogenConcentration, Note: "in scale mode, specifically hydrogen ion concentration"},
		{Symbol: "Molar", Name: "molar H+ concentration", Category: "chemistry", Base: "mol/L H+", kind: kindHydrogenConcentration, Note: "in scale mode, specifically hydrogen ion concentration"},
		{Symbol: "mag", Name: "stellar magnitude difference", Category: "photometry", Base: "brightness ratio", kind: kindMagnitude, Note: "magnitude difference, not absolute calibration"},
		{Symbol: "brightness-ratio", Name: "brightness ratio", Category: "photometry", Base: "brightness ratio", kind: kindBrightnessRatio, Note: "must be positive"},
		{Symbol: "beaufort", Name: "Beaufort wind force", Category: "wind", Base: "m/s", kind: kindBeaufort, Note: "integer lookup scale from 0 through 12"},
		{Symbol: "m/s", Name: "meters per second", Category: "wind", Base: "m/s", kind: kindMetersPerSecond},
		{Symbol: "km/h", Name: "kilometers per hour", Category: "wind", Base: "m/s", kind: kindKilometersPerHour},
		{Symbol: "mph", Name: "miles per hour", Category: "wind", Base: "m/s", kind: kindMilesPerHour},
		{Symbol: "knot", Name: "knot", Category: "wind", Base: "m/s", kind: kindKnot, Aliases: []string{"kt"}},
		{Symbol: "awg", Name: "American wire gauge", Category: "wire", Base: "diameter-in", kind: kindAWG, Note: "numeric -3 (0000) through 40"},
		{Symbol: "diameter-mm", Name: "wire diameter millimeters", Category: "wire", Base: "diameter-in", kind: kindDiameterMM},
		{Symbol: "diameter-in", Name: "wire diameter inches", Category: "wire", Base: "diameter-in", kind: kindDiameterIn},
	} {
		r.add(scale)
	}
}

func (r *Registry) Lookup(name string) (*Scale, error) {
	if IsPaperSize(name) {
		return nil, fmt.Errorf("paper size %q cannot be used with scalar scale conversion; use convunits size", name)
	}
	scale := r.aliases[name]
	if scale == nil {
		return nil, fmt.Errorf("unknown scale %q", name)
	}
	return scale, nil
}

func (r *Registry) IsScale(name string) bool { return r.aliases[name] != nil }

func (r *Registry) Scales(category string) []*Scale {
	var out []*Scale
	for _, scale := range r.scales {
		if category == "" || strings.EqualFold(scale.Category, category) {
			out = append(out, scale)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })
	return out
}

func (r *Registry) Convert(value float64, inputName, outputName string) (Result, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Result{}, fmt.Errorf("scale value must be finite")
	}
	in, err := r.Lookup(inputName)
	if err != nil {
		return Result{}, fmt.Errorf("input scale: %w", err)
	}
	out, err := r.Lookup(outputName)
	if err != nil {
		return Result{}, fmt.Errorf("output scale: %w", err)
	}
	if in.Base != out.Base {
		return Result{}, fmt.Errorf("cannot convert %s to %s: incompatible scale families (%s and %s)", inputName, outputName, in.Category, out.Category)
	}

	switch in.Base {
	case "dB":
		return convertRatio(value, in, out)
	case "mol/L H+":
		return convertPH(value, in, out)
	case "brightness ratio":
		return convertMagnitude(value, in, out)
	case "m/s":
		return convertBeaufort(value, in, out)
	case "diameter-in":
		return convertAWG(value, in, out)
	default:
		return Result{}, fmt.Errorf("unsupported scale family %q", in.Base)
	}
}

func positive(value float64, name string) error {
	if value <= 0 {
		return fmt.Errorf("%s must be greater than zero", name)
	}
	return nil
}

func convertRatio(value float64, in, out *Scale) (Result, error) {
	var db float64
	switch in.kind {
	case kindDB:
		db = value
	case kindBel:
		db = value * 10
	case kindPowerRatio:
		if err := positive(value, "power ratio"); err != nil {
			return Result{}, err
		}
		db = 10 * math.Log10(value)
	case kindAmplitudeRatio:
		if err := positive(value, "amplitude ratio"); err != nil {
			return Result{}, err
		}
		db = 20 * math.Log10(value)
	}
	var result float64
	switch out.kind {
	case kindDB:
		result = db
	case kindBel:
		result = db / 10
	case kindPowerRatio:
		result = math.Pow(10, db/10)
	case kindAmplitudeRatio:
		result = math.Pow(10, db/20)
	}
	return Scalar(result, out.Symbol), nil
}

func convertPH(value float64, in, out *Scale) (Result, error) {
	concentration := value
	if in.kind == kindPH {
		concentration = math.Pow(10, -value)
	} else if err := positive(value, "hydrogen ion concentration"); err != nil {
		return Result{}, err
	}
	if out.kind == kindPH {
		return Scalar(-math.Log10(concentration), out.Symbol), nil
	}
	return Scalar(concentration, out.Symbol), nil
}

func convertMagnitude(value float64, in, out *Scale) (Result, error) {
	brightness := value
	if in.kind == kindMagnitude {
		brightness = math.Pow(10, -0.4*value)
	} else if err := positive(value, "brightness ratio"); err != nil {
		return Result{}, err
	}
	if out.kind == kindMagnitude {
		return Scalar(-2.5*math.Log10(brightness), out.Symbol), nil
	}
	return Scalar(brightness, out.Symbol), nil
}

type beaufortBand struct{ min, max float64 }

var beaufortBands = []beaufortBand{
	{0, 0.5}, {0.5, 1.5}, {1.6, 3.3}, {3.4, 5.5}, {5.5, 7.9}, {8.0, 10.7}, {10.8, 13.8},
	{13.9, 17.1}, {17.2, 20.7}, {20.8, 24.4}, {24.5, 28.4}, {28.5, 32.6}, {32.7, math.Inf(1)},
}

var beaufortLowerBounds = []float64{0, 0.5, 1.6, 3.4, 5.5, 8.0, 10.8, 13.9, 17.2, 20.8, 24.5, 28.5, 32.7}

func convertBeaufort(value float64, in, out *Scale) (Result, error) {
	if in.kind == kindBeaufort {
		if value < 0 || value > 12 || value != math.Trunc(value) {
			return Result{}, fmt.Errorf("invalid Beaufort number %g: expected an integer from 0 through 12", value)
		}
		if out.kind == kindBeaufort {
			return Scalar(value, out.Symbol), nil
		}
		band := beaufortBands[int(value)]
		if value == 0 {
			max := speedFromMPS(band.max, out.kind)
			return Result{Max: &max, Unit: out.Symbol}, nil
		}
		min := speedFromMPS(band.min, out.kind)
		if math.IsInf(band.max, 1) {
			return Result{Min: &min, Unit: out.Symbol}, nil
		}
		max := speedFromMPS(band.max, out.kind)
		return BoundedRange(min, max, out.Symbol), nil
	}

	mps := speedToMPS(value, in.kind)
	if mps < 0 {
		return Result{}, fmt.Errorf("wind speed must not be negative")
	}
	if out.kind != kindBeaufort {
		return Scalar(speedFromMPS(mps, out.kind), out.Symbol), nil
	}
	force := len(beaufortLowerBounds) - 1
	for i := 1; i < len(beaufortLowerBounds); i++ {
		if mps < beaufortLowerBounds[i] {
			force = i - 1
			break
		}
	}
	return Scalar(float64(force), out.Symbol), nil
}

func speedToMPS(value float64, kind scaleKind) float64 {
	switch kind {
	case kindKilometersPerHour:
		return value / 3.6
	case kindMilesPerHour:
		return value * 0.44704
	case kindKnot:
		return value * 1852 / 3600
	default:
		return value
	}
}

func speedFromMPS(value float64, kind scaleKind) float64 {
	switch kind {
	case kindKilometersPerHour:
		return value * 3.6
	case kindMilesPerHour:
		return value / 0.44704
	case kindKnot:
		return value * 3600 / 1852
	default:
		return value
	}
}

func convertAWG(value float64, in, out *Scale) (Result, error) {
	diameterIn := value
	if in.kind == kindAWG {
		if value < -3 || value > 40 || value != math.Trunc(value) {
			return Result{}, fmt.Errorf("invalid AWG value %g: expected an integer from -3 (0000) through 40", value)
		}
		diameterIn = 0.005 * math.Pow(92, (36-value)/39)
	} else {
		if err := positive(value, "wire diameter"); err != nil {
			return Result{}, err
		}
		if in.kind == kindDiameterMM {
			diameterIn = value / 25.4
		}
	}
	var result float64
	switch out.kind {
	case kindAWG:
		result = 36 - 39*math.Log(diameterIn/0.005)/math.Log(92)
	case kindDiameterMM:
		result = diameterIn * 25.4
	case kindDiameterIn:
		result = diameterIn
	}
	return Scalar(result, out.Symbol), nil
}
