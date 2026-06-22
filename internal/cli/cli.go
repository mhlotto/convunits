package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"

	formulacalc "convunits/internal/formula"
	"convunits/internal/scales"
	"convunits/internal/shoes"
	"convunits/internal/solve"
	"convunits/internal/units"
	"convunits/internal/weird/drill"
	"convunits/internal/weird/sieve"
	"convunits/internal/weird/wire"
)

type CLI struct {
	Registry *units.Registry
	Out, Err io.Writer
}

func New(out, err io.Writer) *CLI { return &CLI{Registry: units.NewRegistry(), Out: out, Err: err} }

func (c *CLI) Run(args []string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		c.help()
		return 0
	}
	if args[0] == "units" {
		return c.listUnits(args[1:])
	}
	if args[0] == "solve" {
		return c.runSolve(args[1:])
	}
	if args[0] == "scale" {
		return c.runScale(args[1:])
	}
	if args[0] == "scales" {
		return c.listScales(args[1:])
	}
	if args[0] == "size" || args[0] == "scale-size" || args[0] == "paper" {
		return c.runSize(args[1:])
	}
	if args[0] == "papers" {
		return c.listPapers(args[1:])
	}
	if args[0] == "wire" {
		return c.runWire(args[1:])
	}
	if args[0] == "wires" {
		fmt.Fprintln(c.Out, wire.SyntaxHelp())
		return 0
	}
	if args[0] == "drill" {
		return c.runDrill(args[1:])
	}
	if args[0] == "drills" {
		return c.listDrills(args[1:])
	}
	if args[0] == "sieve" {
		return c.runSieve(args[1:])
	}
	if args[0] == "sieves" {
		return c.listSieves(args[1:])
	}
	if args[0] == "formula" {
		return c.runFormula(args[1:])
	}
	if args[0] == "formulas" {
		return c.listFormulas(args[1:])
	}
	if args[0] == "shoe" {
		return c.runShoe(args[1:])
	}
	fs := flag.NewFlagSet("convunits", flag.ContinueOnError)
	fs.SetOutput(c.Err)
	precision := fs.Int("precision", 10, "significant digits")
	scientific := fs.Bool("scientific", false, "use scientific notation")
	asJSON := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	value, input, output, err := parseOperands(fs.Args())
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 2
	}
	result, err := c.Registry.Convert(value, input, output)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	if *asJSON {
		payload := struct {
			Input struct {
				Value float64 `json:"value"`
				Unit  string  `json:"unit"`
			} `json:"input"`
			Output struct {
				Value float64 `json:"value"`
				Unit  string  `json:"unit"`
			} `json:"output"`
			Dimension string `json:"dimension"`
		}{}
		payload.Input.Value, payload.Input.Unit = value, input
		payload.Output.Value, payload.Output.Unit = result.Value, output
		payload.Dimension = result.Category
		enc := json.NewEncoder(c.Out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		return 0
	}
	fmt.Fprintf(c.Out, "%s %s\n", units.FormatValue(result.Value, *precision, *scientific), output)
	return 0
}

func (c *CLI) runScale(args []string) int {
	precision, scientific, asJSON, positional, err := parseScaleOptions(args)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 2
	}
	if len(positional) != 3 {
		fmt.Fprintln(c.Err, "error: usage: convunits scale [options] VALUE INPUT-SCALE OUTPUT-SCALE")
		return 2
	}
	value, err := strconv.ParseFloat(positional[0], 64)
	if err != nil {
		fmt.Fprintf(c.Err, "error: invalid scale value %q\n", positional[0])
		return 2
	}
	result, err := scales.NewRegistry().Convert(value, positional[1], positional[2])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	if asJSON {
		payload := struct {
			Input struct {
				Value float64 `json:"value"`
				Scale string  `json:"scale"`
			} `json:"input"`
			Output struct {
				Value *float64 `json:"value,omitempty"`
				Min   *float64 `json:"min,omitempty"`
				Max   *float64 `json:"max,omitempty"`
				Unit  string   `json:"unit"`
			} `json:"output"`
		}{}
		payload.Input.Value, payload.Input.Scale = value, positional[1]
		payload.Output.Value, payload.Output.Min, payload.Output.Max, payload.Output.Unit = result.Value, result.Min, result.Max, result.Unit
		enc := json.NewEncoder(c.Out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		return 0
	}
	fmt.Fprintln(c.Out, formatScaleResult(result, precision, scientific))
	return 0
}

func parseScaleOptions(args []string) (int, bool, bool, []string, error) {
	precision, scientific, asJSON := 10, false, false
	var positional []string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--json":
			asJSON = true
		case args[i] == "--scientific":
			scientific = true
		case args[i] == "--precision":
			i++
			if i >= len(args) {
				return 0, false, false, nil, fmt.Errorf("--precision requires an integer")
			}
			value, err := strconv.Atoi(args[i])
			if err != nil || value <= 0 {
				return 0, false, false, nil, fmt.Errorf("--precision must be a positive integer")
			}
			precision = value
		case strings.HasPrefix(args[i], "--precision="):
			value, err := strconv.Atoi(strings.TrimPrefix(args[i], "--precision="))
			if err != nil || value <= 0 {
				return 0, false, false, nil, fmt.Errorf("--precision must be a positive integer")
			}
			precision = value
		case strings.HasPrefix(args[i], "-"):
			if _, err := strconv.ParseFloat(args[i], 64); err == nil {
				positional = append(positional, args[i])
				continue
			}
			return 0, false, false, nil, fmt.Errorf("unknown scale option %q", args[i])
		default:
			positional = append(positional, args[i])
		}
	}
	return precision, scientific, asJSON, positional, nil
}

func formatScaleResult(result scales.Result, precision int, scientific bool) string {
	format := func(value float64) string { return units.FormatValue(value, precision, scientific) }
	switch {
	case result.Value != nil:
		return format(*result.Value) + " " + result.Unit
	case result.Min == nil:
		return "<" + format(*result.Max) + " " + result.Unit
	case result.Max == nil:
		return ">=" + format(*result.Min) + " " + result.Unit
	default:
		return format(*result.Min) + "-" + format(*result.Max) + " " + result.Unit
	}
}

func (c *CLI) listScales(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits scales [category]")
		return 2
	}
	category := ""
	if len(args) == 1 {
		category = args[0]
	}
	list := scales.NewRegistry().Scales(category)
	if len(list) == 0 {
		fmt.Fprintf(c.Err, "error: unknown or empty scale category %q\n", category)
		return 1
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, scale := range list {
		aliases := strings.Join(scale.Aliases, ", ")
		if aliases == "" {
			aliases = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", scale.Symbol, scale.Name, scale.Category, scale.Base, aliases, scale.Note)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) runSize(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits size PAPER-SIZE OUTPUT-LENGTH-UNIT")
		return 2
	}
	scaleRegistry := scales.NewRegistry()
	if scaleRegistry.IsScale(args[0]) {
		fmt.Fprintf(c.Err, "error: scalar scale %q cannot be used as a paper size; use convunits scale\n", args[0])
		return 1
	}
	size, err := scales.LookupPaperSize(args[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	width, err := c.Registry.Convert(size.WidthM, "m", args[1])
	if err != nil {
		fmt.Fprintln(c.Err, "error: paper output unit must be length:", err)
		return 1
	}
	height, err := c.Registry.Convert(size.HeightM, "m", args[1])
	if err != nil {
		fmt.Fprintln(c.Err, "error: paper output unit must be length:", err)
		return 1
	}
	prefix := ""
	if size.Approximate {
		prefix = "approximately "
	}
	fmt.Fprintf(c.Out, "%s%s x %s %s\n", prefix, units.FormatValue(width.Value, 10, false), units.FormatValue(height.Value, 10, false), args[1])
	return 0
}

func (c *CLI) listPapers(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits papers [iso|us|photo]")
		return 2
	}
	category := ""
	if len(args) == 1 {
		category = args[0]
	}
	list := scales.PaperSizes(category)
	if len(list) == 0 {
		fmt.Fprintf(c.Err, "error: unknown or empty paper category %q\n", category)
		return 1
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, size := range list {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s x %s mm\n", size.Symbol, size.Name, size.Category,
			units.FormatValue(size.WidthM*1000, 10, false), units.FormatValue(size.HeightM*1000, 10, false))
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) runWire(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits wire GAUGE OUTPUT-LENGTH-UNIT")
		return 2
	}
	gauge, err := wire.ParseGauge(args[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	meters, err := wire.DiameterMeters(gauge)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return c.printLengthResult(meters, args[1], "diameter", true)
}

func (c *CLI) runDrill(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits drill SIZE OUTPUT-LENGTH-UNIT")
		return 2
	}
	meters, err := drill.DiameterMeters(args[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return c.printLengthResult(meters, args[1], "diameter", true)
}

func (c *CLI) runSieve(args []string) int {
	if len(args) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits sieve SIZE OUTPUT-LENGTH-UNIT")
		return 2
	}
	meters, err := sieve.OpeningMeters(args[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return c.printLengthResult(meters, args[1], "opening", true)
}

func (c *CLI) printLengthResult(meters float64, outputUnit, label string, approximate bool) int {
	converted, err := c.Registry.Convert(meters, "m", outputUnit)
	if err != nil {
		fmt.Fprintf(c.Err, "error: output unit for %s must be length: %v\n", label, err)
		return 1
	}
	prefix := ""
	if approximate {
		prefix = "approximately "
	}
	fmt.Fprintf(c.Out, "%s%s %s %s\n", prefix, units.FormatValue(converted.Value, 10, false), outputUnit, label)
	return 0
}

func (c *CLI) listDrills(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits drills [number|letter|fractional]")
		return 2
	}
	category := ""
	if len(args) == 1 {
		category = args[0]
	}
	entries, err := drill.Entries(category)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, entry := range entries {
		fmt.Fprintf(w, "%s\t%s\t%g in\n", entry.Size, entry.Category, entry.DiameterIn)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) listSieves(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(c.Err, "error: usage: convunits sieves")
		return 2
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, entry := range sieve.Entries() {
		fmt.Fprintf(w, "%s\t%g mm nominal opening\n", entry.Size, entry.OpeningMM)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) runFormula(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits formula NAME [--ARG VALUEUNIT ...] OUTPUT-UNIT")
		return 2
	}
	name := args[0]
	rawInputs := make(map[string]string)
	var positional []string
	allowed := map[string]bool{"mass": true, "radius": true, "height": true, "length": true, "gravity": true}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			positional = append(positional, arg)
			continue
		}
		keyValue := strings.TrimPrefix(arg, "--")
		key, value, hasValue := strings.Cut(keyValue, "=")
		if !allowed[key] {
			fmt.Fprintf(c.Err, "error: unknown formula argument --%s\n", key)
			return 2
		}
		if _, exists := rawInputs[key]; exists {
			fmt.Fprintf(c.Err, "error: formula argument --%s supplied more than once\n", key)
			return 2
		}
		if !hasValue {
			i++
			if i >= len(args) {
				fmt.Fprintf(c.Err, "error: formula argument --%s requires VALUEUNIT\n", key)
				return 2
			}
			value = args[i]
		}
		rawInputs[key] = value
	}
	if len(positional) != 1 {
		fmt.Fprintln(c.Err, "error: formula requires exactly one output unit")
		return 2
	}
	inputs := make(map[string]formulacalc.Input)
	for key, raw := range rawInputs {
		value, unit, err := splitValueUnit(raw)
		if err != nil {
			fmt.Fprintf(c.Err, "error: --%s: %v\n", key, err)
			return 2
		}
		inputs[key] = formulacalc.Input{Value: value, Unit: unit}
	}
	result, err := formulacalc.New(c.Registry).Compute(name, inputs, positional[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	prefix := ""
	if result.Approximate {
		prefix = "approximately "
	}
	fmt.Fprintf(c.Out, "%s%s %s\n", prefix, units.FormatValue(result.Value, 10, false), result.Unit)
	return 0
}

func (c *CLI) listFormulas(args []string) int {
	if len(args) != 0 {
		fmt.Fprintln(c.Err, "error: usage: convunits formulas")
		return 2
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, definition := range formulacalc.Definitions() {
		fmt.Fprintf(w, "%s\t%s\t%s\n", definition.Name, definition.Arguments, definition.OutputDimension)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) runShoe(args []string) int {
	if len(args) == 1 && args[0] == "systems" {
		for _, system := range shoes.Systems() {
			fmt.Fprintln(c.Out, system)
		}
		return 0
	}
	if len(args) != 3 {
		fmt.Fprintln(c.Err, "error: usage: convunits shoe SYSTEM SIZE OUTPUT-LENGTH-UNIT\n       convunits shoe systems")
		return 2
	}
	size, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		fmt.Fprintf(c.Err, "error: invalid shoe size %q\n", args[1])
		return 2
	}
	meters, err := shoes.FootLengthMeters(args[0], size)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	converted, err := c.Registry.Convert(meters, "m", args[2])
	if err != nil {
		fmt.Fprintln(c.Err, "error: shoe output unit must be length:", err)
		return 1
	}
	fmt.Fprintf(c.Out, "approximately %s %s foot length\n", units.FormatValue(converted.Value, 10, false), args[2])
	return 0
}

func (c *CLI) runSolve(args []string) int {
	precision := 10
	scientific := false
	asJSON := false
	givens := make(map[string]solve.Quantity)
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--given":
			i++
			if i >= len(args) {
				fmt.Fprintln(c.Err, "error: --given requires NAME=VALUEUNIT")
				return 2
			}
			if err := addGiven(givens, args[i]); err != nil {
				fmt.Fprintln(c.Err, "error:", err)
				return 2
			}
		case strings.HasPrefix(arg, "--given="):
			if err := addGiven(givens, strings.TrimPrefix(arg, "--given=")); err != nil {
				fmt.Fprintln(c.Err, "error:", err)
				return 2
			}
		case arg == "--precision":
			i++
			if i >= len(args) {
				fmt.Fprintln(c.Err, "error: --precision requires an integer")
				return 2
			}
			value, err := strconv.Atoi(args[i])
			if err != nil || value <= 0 {
				fmt.Fprintln(c.Err, "error: --precision must be a positive integer")
				return 2
			}
			precision = value
		case strings.HasPrefix(arg, "--precision="):
			value, err := strconv.Atoi(strings.TrimPrefix(arg, "--precision="))
			if err != nil || value <= 0 {
				fmt.Fprintln(c.Err, "error: --precision must be a positive integer")
				return 2
			}
			precision = value
		case arg == "--scientific":
			scientific = true
		case arg == "--json":
			asJSON = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(c.Err, "error: unknown solve option %q\n", arg)
			return 2
		default:
			positional = append(positional, arg)
		}
	}

	input, outputUnit, err := parseSolveOperands(positional)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 2
	}
	result, err := solve.New(c.Registry).Solve(input, outputUnit, givens)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	if asJSON {
		payload := struct {
			Variable string  `json:"variable"`
			Minimum  float64 `json:"minimum"`
			Maximum  float64 `json:"maximum"`
			Unit     string  `json:"unit"`
		}{result.Variable, result.Range.Min, result.Range.Max, result.Unit}
		enc := json.NewEncoder(c.Out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(payload); err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		return 0
	}
	if result.Range.Scalar() {
		fmt.Fprintf(c.Out, "%s %s\n", units.FormatValue(result.Range.Min, precision, scientific), result.Unit)
	} else {
		fmt.Fprintf(c.Out, "%s-%s %s\n",
			units.FormatValue(result.Range.Min, precision, scientific),
			units.FormatValue(result.Range.Max, precision, scientific), result.Unit)
	}
	return 0
}

func addGiven(givens map[string]solve.Quantity, text string) error {
	name, value, ok := strings.Cut(text, "=")
	if !ok || name == "" || value == "" {
		return fmt.Errorf("invalid --given %q; expected NAME=VALUEUNIT", text)
	}
	if _, exists := givens[name]; exists {
		return fmt.Errorf("given %s was supplied more than once", name)
	}
	quantity, err := parseQuantity(value)
	if err != nil {
		return fmt.Errorf("given %s: %w", name, err)
	}
	givens[name] = quantity
	return nil
}

func parseSolveOperands(args []string) (solve.Quantity, string, error) {
	if len(args) == 2 {
		quantity, err := parseQuantity(args[0])
		return quantity, args[1], err
	}
	if len(args) == 3 {
		rangeValue, err := parseNumericRange(args[0])
		if err != nil {
			return solve.Quantity{}, "", err
		}
		return solve.Quantity{Range: rangeValue, Unit: args[1]}, args[2], nil
	}
	return solve.Quantity{}, "", fmt.Errorf("usage: convunits solve VALUEUNIT OUTPUTUNIT --given NAME=VALUEUNIT ...")
}

func parseQuantity(text string) (solve.Quantity, error) {
	if before, after, ok := strings.Cut(text, ".."); ok {
		min, err := strconv.ParseFloat(before, 64)
		if err != nil {
			return solve.Quantity{}, fmt.Errorf("invalid range minimum %q", before)
		}
		max, unit, err := splitValueUnit(after)
		if err != nil {
			return solve.Quantity{}, err
		}
		interval, err := solve.NewInterval(min, max)
		return solve.Quantity{Range: interval, Unit: unit}, err
	}
	value, unit, err := splitValueUnit(text)
	if err != nil {
		return solve.Quantity{}, err
	}
	interval, err := solve.NewInterval(value, value)
	return solve.Quantity{Range: interval, Unit: unit}, err
}

func parseNumericRange(text string) (solve.Interval, error) {
	if before, after, ok := strings.Cut(text, ".."); ok {
		min, err := strconv.ParseFloat(before, 64)
		if err != nil {
			return solve.Interval{}, fmt.Errorf("invalid range minimum %q", before)
		}
		max, err := strconv.ParseFloat(after, 64)
		if err != nil {
			return solve.Interval{}, fmt.Errorf("invalid range maximum %q", after)
		}
		return solve.NewInterval(min, max)
	}
	value, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return solve.Interval{}, fmt.Errorf("invalid numeric value %q", text)
	}
	return solve.NewInterval(value, value)
}

func parseOperands(args []string) (float64, string, string, error) {
	if len(args) == 2 {
		value, unit, err := splitValueUnit(args[0])
		if err != nil {
			return 0, "", "", err
		}
		return value, unit, args[1], nil
	}
	if len(args) == 3 {
		v, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return 0, "", "", fmt.Errorf("invalid numeric value %q", args[0])
		}
		return v, args[1], args[2], nil
	}
	return 0, "", "", fmt.Errorf("usage: convunits [options] <value><input-unit> <output-unit>\n       convunits [options] <value> <input-unit> <output-unit>")
}

func splitValueUnit(s string) (float64, string, error) {
	for i := 1; i <= len(s); i++ {
		if v, err := strconv.ParseFloat(s[:i], 64); err == nil && i < len(s) {
			// Continue while the next byte may still be numeric; keep the longest valid prefix.
			bestV, bestI := v, i
			for j := i + 1; j <= len(s); j++ {
				if x, e := strconv.ParseFloat(s[:j], 64); e == nil && j < len(s) {
					bestV, bestI = x, j
				}
			}
			return bestV, s[bestI:], nil
		}
	}
	return 0, "", fmt.Errorf("invalid value and unit %q", s)
}

func (c *CLI) listUnits(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits units [category]")
		return 2
	}
	category := ""
	if len(args) == 1 {
		category = args[0]
	}
	list := c.Registry.Units(category)
	if len(list) == 0 {
		fmt.Fprintf(c.Err, "error: unknown or empty unit category %q\n", category)
		return 1
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, u := range list {
		aliases := strings.Join(u.Aliases, ", ")
		if aliases == "" {
			aliases = "-"
		}
		details := ""
		if u.Approximate {
			details = "approximate"
			if u.Note != "" {
				details += ": " + u.Note
			}
		} else if u.Note != "" {
			details = u.Note
		}
		if u.SourceNote != "" {
			if details != "" {
				details += "; "
			}
			details += "source: " + u.SourceNote
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", u.Symbol, u.Name, u.Category, u.Dimension, aliases, details)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) help() {
	fmt.Fprint(c.Out, `convunits converts scalar, compound, and derived units.

Usage:
  convunits [--precision N] [--scientific] [--json] <value><input-unit> <output-unit>
  convunits [options] <value> <input-unit> <output-unit>
  convunits units [category]
  convunits solve VALUEUNIT OUTPUTUNIT --given NAME=VALUEUNIT ...
  convunits scale VALUE INPUT-SCALE OUTPUT-SCALE
  convunits scales [category]
  convunits size PAPER-SIZE OUTPUT-LENGTH-UNIT
  convunits shoe SYSTEM SIZE OUTPUT-LENGTH-UNIT
  convunits shoe systems
  convunits paper PAPER-SIZE OUTPUT-LENGTH-UNIT
  convunits papers [iso|us|photo]
  convunits wire GAUGE OUTPUT-LENGTH-UNIT
  convunits wires
  convunits drill SIZE OUTPUT-LENGTH-UNIT
  convunits drills [number|letter|fractional]
  convunits sieve SIZE OUTPUT-LENGTH-UNIT
  convunits sieves
  convunits formula NAME [--ARG VALUEUNIT ...] OUTPUT-UNIT
  convunits formulas

Examples:
  convunits 10kg lb
  convunits 60 mph km/h
  convunits 1N 'kg*m/s^2'
  convunits 100F C
  convunits 30mpg L/100km
  convunits solve 10N s --given mass=2kg --given distance=5m
  convunits scale 5 beaufort m/s
  convunits size a4 mm
  convunits shoe eu 43 cm
  convunits wire 12awg mm
  convunits drill '#7' mm
  convunits sieve '#40' mm
  convunits formula escape-velocity --mass 1Mearth --radius 1Re km/s

Unit expressions support *, /, integer powers, and parentheses. Parsing is case-sensitive.
The catalog below lists symbols, long names, categories, dimensions, and aliases.

`)
	_ = c.listUnits(nil)
}
