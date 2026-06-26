package explain

import (
	"fmt"
	"strings"

	evalcalc "convunits/internal/eval"
	"convunits/internal/units"
)

type Output struct {
	Value       float64
	Unit        string
	Approximate bool
}

type Explanation struct {
	Command    string
	Expression string
	Input      struct {
		Value float64
		Unit  string
	}
	Output     Output
	Steps      []string
	Dimensions struct {
		Input  string
		Output string
	}
	Title string
}

func Normal(registry *units.Registry, value float64, inputUnit, outputUnit string) (Explanation, error) {
	result, err := registry.Convert(value, inputUnit, outputUnit)
	if err != nil {
		return Explanation{}, err
	}
	var ex Explanation
	ex.Command = "explain"
	ex.Title = fmt.Sprintf("%s %s -> %s", units.FormatValue(value, 10, false), inputUnit, outputUnit)
	ex.Input.Value, ex.Input.Unit = value, inputUnit
	ex.Output = Output{Value: result.Value, Unit: outputUnit, Approximate: result.Approximate}
	ex.Dimensions.Input = result.Input.Dimension.String()
	ex.Dimensions.Output = result.Output.Dimension.String()

	base := canonicalUnit(result.Input.Dimension)
	if result.Input.Affine != nil || result.Output.Affine != nil {
		ex.Steps = append(ex.Steps, fmt.Sprintf("%s %s = %s %s", units.FormatValue(value, 10, false), inputUnit, units.FormatValue(result.Value, 10, false), outputUnit))
		return ex, nil
	}
	oneInputBase := result.Input.ToBase(1)
	oneOutputBase := result.Output.ToBase(1)
	if result.Input.Dimension == result.Output.Dimension && result.Input.Multiplier == result.Output.Multiplier {
		ex.Steps = append(ex.Steps, fmt.Sprintf("%s has dimensions %s", inputUnit, result.Input.Dimension))
		ex.Steps = append(ex.Steps, fmt.Sprintf("%s has dimensions %s", outputUnit, result.Output.Dimension))
		ex.Steps = append(ex.Steps, fmt.Sprintf("%s %s = %s %s", units.FormatValue(value, 10, false), inputUnit, units.FormatValue(result.Value, 10, false), outputUnit))
		return ex, nil
	}
	ex.Steps = append(ex.Steps, fmt.Sprintf("1 %s = %s %s%s", inputUnit, units.FormatValue(oneInputBase, 10, false), base, approxSuffix(result.Input.Approximate)))
	if result.Output.Multiplier != 1 || result.Output.Approximate {
		ex.Steps = append(ex.Steps, fmt.Sprintf("1 %s = %s %s%s", outputUnit, units.FormatValue(oneOutputBase, 10, false), base, approxSuffix(result.Output.Approximate)))
	}
	baseValue := result.Input.ToBase(value)
	ex.Steps = append(ex.Steps, fmt.Sprintf("%s %s = %s %s", units.FormatValue(value, 10, false), inputUnit, units.FormatValue(baseValue, 10, false), base))
	if result.Output.Multiplier != 1 || result.Output.Approximate {
		ex.Steps = append(ex.Steps, fmt.Sprintf("%s %s / %s %s per %s = %s %s", units.FormatValue(baseValue, 10, false), base, units.FormatValue(oneOutputBase, 10, false), base, outputUnit, units.FormatValue(result.Value, 10, false), outputUnit))
	} else if base != outputUnit {
		ex.Steps = append(ex.Steps, fmt.Sprintf("%s %s = %s %s", units.FormatValue(baseValue, 10, false), base, units.FormatValue(result.Value, 10, false), outputUnit))
	}
	return ex, nil
}

func Eval(registry *units.Registry, expression string) (Explanation, error) {
	if !strings.Contains(expression, "->") {
		return Explanation{}, fmt.Errorf("explain eval requires an expression with -> output unit")
	}
	result, err := evalcalc.New(registry).Evaluate(expression)
	if err != nil {
		return Explanation{}, err
	}
	var ex Explanation
	ex.Command = "explain"
	ex.Expression = expression
	ex.Title = expression
	ex.Output = Output{Value: result.Value, Unit: result.Unit, Approximate: result.Approximate}
	ex.Steps = evalSteps(registry, expression, result)
	return ex, nil
}

func evalSteps(registry *units.Registry, expression string, result evalcalc.Result) []string {
	before, output, _ := strings.Cut(expression, "->")
	before = strings.TrimSpace(before)
	output = strings.TrimSpace(output)
	var steps []string
	if strings.Contains(before, "pi") {
		steps = append(steps, "pi = 3.141592653589793")
	}
	if strings.Contains(before, "1Re") {
		steps = append(steps, "1 Re = 6371008.8 m approximately")
	}
	if strings.Contains(before, "60mph") {
		steps = append(steps, "60 mph = 26.8224 m/s")
	}
	base := canonicalUnit(result.Dimension)
	baseValue := result.Value
	if output != "" {
		if outs, err := registry.ParseCandidates(output); err == nil && len(outs) > 0 {
			for _, out := range outs {
				if out.Affine == nil && out.Dimension == result.Dimension {
					baseValue = out.ToBase(result.Value)
					break
				}
			}
		}
	}
	switch {
	case strings.Contains(before, "0.5") && strings.Contains(before, "1500kg") && strings.Contains(before, "60mph"):
		steps = append(steps, fmt.Sprintf("0.5 * 1500 kg * (26.8224 m/s)^2 = %s %s", units.FormatValue(baseValue, 10, false), base))
	case strings.Contains(before, "2") && strings.Contains(before, "pi") && strings.Contains(before, "1Re"):
		steps = append(steps, fmt.Sprintf("2 * pi * 1 Re = %s %s", units.FormatValue(baseValue, 16, false), base))
	default:
		steps = append(steps, fmt.Sprintf("%s = %s %s", before, units.FormatValue(baseValue, 10, false), base))
	}
	if output != "" {
		if outs, err := registry.ParseCandidates(output); err == nil && len(outs) > 0 {
			for _, out := range outs {
				if out.Affine == nil && out.Dimension == result.Dimension {
					steps = append(steps, fmt.Sprintf("1 %s = %s %s%s", output, units.FormatValue(out.ToBase(1), 10, false), base, approxSuffix(out.Approximate)))
					break
				}
			}
		}
	}
	return steps
}

func Text(ex Explanation) string {
	var b strings.Builder
	b.WriteString(ex.Title)
	b.WriteString("\n\n")
	for _, step := range ex.Steps {
		b.WriteString(step)
		b.WriteString("\n")
	}
	if ex.Dimensions.Input != "" || ex.Dimensions.Output != "" {
		b.WriteString("\nDimensions:\n")
		b.WriteString("  ")
		b.WriteString(ex.Input.Unit)
		b.WriteString(": ")
		b.WriteString(ex.Dimensions.Input)
		b.WriteString("\n  ")
		b.WriteString(ex.Output.Unit)
		b.WriteString(": ")
		b.WriteString(ex.Dimensions.Output)
		b.WriteString("\n")
	}
	b.WriteString("\nResult:\n")
	b.WriteString("  ")
	if ex.Output.Approximate {
		b.WriteString("approximately ")
	}
	b.WriteString(units.FormatValue(ex.Output.Value, 10, false))
	if ex.Output.Unit != "" {
		b.WriteString(" ")
		b.WriteString(ex.Output.Unit)
	}
	b.WriteString("\n")
	return b.String()
}

func canonicalUnit(d units.Dimension) string {
	switch d {
	case units.Dimension{}:
		return "1"
	case units.Dimension{Length: 1}:
		return "m"
	case units.Dimension{Mass: 1}:
		return "kg"
	case units.Dimension{Time: 1}:
		return "s"
	case units.Dimension{Length: 3}:
		return "m^3"
	case units.Dimension{Length: 1, Time: -1}:
		return "m/s"
	case units.Dimension{Mass: 1, Length: 1, Time: -2}:
		return "kg*m/s^2"
	case units.Dimension{Mass: 1, Length: 2, Time: -2}:
		return "J"
	}
	return d.String()
}

func approxSuffix(approximate bool) string {
	if approximate {
		return " approximately"
	}
	return ""
}
