package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"convunits/internal/discovery"
	evalcalc "convunits/internal/eval"
	"convunits/internal/explain"
	formulacalc "convunits/internal/formula"
	"convunits/internal/recipe"
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
	globalJSON := false
	if args[0] == "--json" {
		globalJSON = true
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintln(c.Err, "error: usage: convunits [options] <value><input-unit> <output-unit>")
		return 2
	}
	if args[0] == "units" {
		return c.listUnits(args[1:])
	}
	if args[0] == "search" {
		return c.runSearch(args[1:], globalJSON)
	}
	if args[0] == "aliases" {
		return c.runAliases(args[1:], globalJSON)
	}
	if args[0] == "compare" {
		return c.runCompare(args[1:], globalJSON)
	}
	if args[0] == "eval" {
		return c.runEval(args[1:], globalJSON)
	}
	if args[0] == "explain" {
		return c.runExplain(args[1:], globalJSON)
	}
	if args[0] == "solve" {
		return c.runSolve(args[1:], globalJSON)
	}
	if args[0] == "scale" {
		return c.runScale(args[1:], globalJSON)
	}
	if args[0] == "scales" {
		return c.listScales(args[1:])
	}
	if args[0] == "recipe" {
		return c.runRecipe(args[1:], globalJSON)
	}
	if args[0] == "size" || args[0] == "scale-size" || args[0] == "paper" {
		return c.runSize(args[1:], globalJSON)
	}
	if args[0] == "papers" {
		return c.listPapers(args[1:])
	}
	if args[0] == "wire" {
		return c.runWire(args[1:], globalJSON)
	}
	if args[0] == "wires" {
		fmt.Fprintln(c.Out, wire.SyntaxHelp())
		return 0
	}
	if args[0] == "drill" {
		return c.runDrill(args[1:], globalJSON)
	}
	if args[0] == "drills" {
		return c.listDrills(args[1:])
	}
	if args[0] == "sieve" {
		return c.runSieve(args[1:], globalJSON)
	}
	if args[0] == "sieves" {
		return c.listSieves(args[1:])
	}
	if args[0] == "formula" {
		return c.runFormula(args[1:], globalJSON)
	}
	if args[0] == "formulas" {
		return c.listFormulas(args[1:])
	}
	if args[0] == "shoe" {
		return c.runShoe(args[1:], globalJSON)
	}
	fs := flag.NewFlagSet("convunits", flag.ContinueOnError)
	fs.SetOutput(c.Err)
	precision := fs.Int("precision", 10, "significant digits")
	scientific := fs.Bool("scientific", false, "use scientific notation")
	asJSON := fs.Bool("json", globalJSON, "emit JSON")
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

func (c *CLI) runScale(args []string, globalJSON bool) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		c.scaleHelp()
		return 0
	}
	precision, scientific, asJSON, positional, err := parseScaleOptions(args)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 2
	}
	asJSON = asJSON || globalJSON
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

func (c *CLI) scaleHelp() {
	fmt.Fprint(c.Out, `convunits scale converts nonlinear, logarithmic, ordinal, and lookup scales.

Usage:
  convunits scale [--precision N] [--scientific] [--json] VALUE INPUT-SCALE OUTPUT-SCALE
  convunits scales [category]

Examples:
  convunits scale 7 pH mol/L
  convunits scale 60 dB power-ratio
  convunits scale 5 beaufort mph
  convunits scale 12 awg diameter-mm

Scale conversions are separate from the normal dimensional unit engine.
`)
}

func (c *CLI) runSearch(args []string, globalJSON bool) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		c.searchHelp()
		return 0
	}
	asJSON := globalJSON
	all := false
	kind := ""
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--json":
			asJSON = true
		case arg == "--all":
			all = true
		case arg == "--kind":
			if i+1 >= len(args) {
				fmt.Fprintln(c.Err, "error: --kind requires a value")
				return 2
			}
			i++
			kind = args[i]
		case strings.HasPrefix(arg, "--kind="):
			kind = strings.TrimPrefix(arg, "--kind=")
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(c.Err, "error: unknown search option %q\n", arg)
			return 2
		default:
			positional = append(positional, arg)
		}
	}
	if len(positional) != 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits search QUERY [--all] [--kind KIND] [--json]")
		return 2
	}
	query := positional[0]
	results := discovery.New(c.Registry).Search(query, kind, all)
	if asJSON {
		payload := struct {
			Command string             `json:"command"`
			Query   string             `json:"query"`
			Results []discovery.Result `json:"results"`
		}{"search", query, results}
		return c.encodeJSON(payload)
	}
	if len(results) == 0 {
		fmt.Fprintf(c.Err, "error: no search results for %q\n", query)
		return 1
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, result := range results {
		details := formatDiscoveryDetails(result.Aliases, result.Description, result.Approximate, result.Dimension)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", result.Kind, result.Key, result.Name, result.Category, details)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) runAliases(args []string, globalJSON bool) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		c.aliasesHelp()
		return 0
	}
	asJSON := globalJSON
	all := false
	var positional []string
	for _, arg := range args {
		switch arg {
		case "--json":
			asJSON = true
		case "--all":
			all = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(c.Err, "error: unknown aliases option %q\n", arg)
				return 2
			}
			positional = append(positional, arg)
		}
	}
	if all {
		if len(positional) != 0 {
			fmt.Fprintln(c.Err, "error: usage: convunits aliases --all [--json]")
			return 2
		}
		matches := discovery.New(c.Registry).AllAliases()
		if asJSON {
			payload := struct {
				Command string                 `json:"command"`
				Matches []discovery.AliasMatch `json:"matches"`
			}{"aliases", matches}
			return c.encodeJSON(payload)
		}
		return c.printAllAliases(matches)
	}
	if len(positional) != 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits aliases [--json] UNIT-OR-ALIAS")
		return 2
	}
	query := positional[0]
	matches := discovery.New(c.Registry).Aliases(query)
	if len(matches) == 0 {
		fmt.Fprintf(c.Err, "error: no aliases or catalog entry found for %q\n", query)
		return 1
	}
	if asJSON {
		payload := struct {
			Command string                 `json:"command"`
			Query   string                 `json:"query"`
			Matches []discovery.AliasMatch `json:"matches"`
		}{"aliases", query, matches}
		return c.encodeJSON(payload)
	}
	for i, match := range matches {
		if i > 0 {
			fmt.Fprintln(c.Out)
		}
		c.printAliasMatch(query, match)
	}
	return 0
}

func (c *CLI) printAllAliases(matches []discovery.AliasMatch) int {
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, match := range matches {
		aliases := strings.Join(match.Aliases, ", ")
		if aliases == "" {
			aliases = "none"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\taliases: %s\n", match.Kind, match.Key, match.Name, match.Category, aliases)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) printAliasMatch(query string, match discovery.AliasMatch) {
	fmt.Fprintln(c.Out, query)
	fmt.Fprintf(c.Out, "  kind: %s\n", match.Kind)
	if match.Canonical != "" && match.Canonical != query {
		fmt.Fprintf(c.Out, "  canonical: %s\n", match.Canonical)
	}
	fmt.Fprintf(c.Out, "  name: %s\n", match.Name)
	if match.Category != "" {
		fmt.Fprintf(c.Out, "  category: %s\n", match.Category)
	}
	aliases := strings.Join(match.Aliases, ", ")
	if aliases == "" {
		aliases = "none"
	}
	fmt.Fprintf(c.Out, "  aliases: %s\n", aliases)
	if match.Dimension != "" {
		fmt.Fprintf(c.Out, "  dimensions: %s\n", match.Dimension)
	}
	if match.Approximate {
		text := match.Description
		if text == "" {
			text = "yes"
		}
		fmt.Fprintf(c.Out, "  approximate: %s\n", text)
	}
	if match.DensityUnit != "" {
		fmt.Fprintf(c.Out, "  density: %g %s\n", match.DensityValue, match.DensityUnit)
	}
	if match.Description != "" && !match.Approximate {
		fmt.Fprintf(c.Out, "  note: %s\n", match.Description)
	}
	if match.Description != "" && match.Approximate && match.Kind == "ingredient" {
		fmt.Fprintf(c.Out, "  note: %s\n", match.Description)
	}
	if match.MatchedBy != "" && match.MatchedBy != "key" {
		fmt.Fprintf(c.Out, "  matched: %s %q\n", match.MatchedBy, match.MatchedString)
	}
}

func (c *CLI) searchHelp() {
	fmt.Fprint(c.Out, `convunits search finds units, aliases, scales, formulas, ingredients, lookup entries, and commands.

Usage:
  convunits search QUERY [--all] [--kind KIND] [--json]
  convunits --json search QUERY

Examples:
  convunits search jupiter
  convunits search flour
  convunits search beaufort
  convunits search '#40'
  convunits search schwarzschild --kind formula

Default output is limited to 20 results. Use --all to show every match.
`)
}

func (c *CLI) aliasesHelp() {
	fmt.Fprint(c.Out, `convunits aliases shows canonical names, aliases, dimensions, and notes.

Usage:
  convunits aliases UNIT-OR-ALIAS
  convunits aliases --all
  convunits aliases --json UNIT-OR-ALIAS
  convunits --json aliases UNIT-OR-ALIAS

Examples:
  convunits aliases mph
  convunits aliases Rj
  convunits aliases flour
  convunits aliases --all
`)
}

func formatDiscoveryDetails(aliases []string, description string, approximate bool, dimension string) string {
	var parts []string
	if len(aliases) > 0 {
		parts = append(parts, "aliases: "+strings.Join(aliases, ", "))
	}
	if description != "" {
		parts = append(parts, description)
	}
	if dimension != "" && !strings.Contains(description, "dimensions:") {
		parts = append(parts, "dimensions: "+dimension)
	}
	if approximate {
		parts = append(parts, "approximate")
	}
	return strings.Join(parts, "; ")
}

type compareOutput struct {
	value       float64
	unit        string
	approximate bool
}

func (c *CLI) runCompare(args []string, globalJSON bool) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		c.compareHelp()
		return 0
	}
	asJSON := globalJSON
	noLimit := false
	var presets []string
	var positional []string
	for _, arg := range args {
		switch arg {
		case "--json":
			asJSON = true
		case "--no-limit":
			noLimit = true
		case "--fun", "--human", "--astronomical", "--ancient", "--all":
			presets = append(presets, strings.TrimPrefix(arg, "--"))
		default:
			if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(c.Err, "error: unknown compare option %q\n", arg)
				return 2
			}
			positional = append(positional, arg)
		}
	}
	value, inputUnit, targets, err := parseCompareOperands(positional)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 2
	}
	explicitTargets := make(map[string]bool)
	for _, target := range targets {
		explicitTargets[target] = true
	}
	if err := c.rejectCompareAffine(inputUnit, "input"); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	for _, preset := range presets {
		targets = append(targets, c.comparePresetTargets(value, inputUnit, preset, noLimit)...)
	}
	targets = uniqueStrings(targets)
	if len(targets) == 0 {
		fmt.Fprintln(c.Err, "error: compare requires at least one target unit or compatible preset")
		return 2
	}
	var outputs []compareOutput
	var failed []string
	for _, target := range targets {
		if err := c.rejectCompareAffine(target, "target"); err != nil {
			if explicitTargets[target] {
				failed = append(failed, err.Error())
			}
			continue
		}
		result, err := c.Registry.Convert(value, inputUnit, target)
		if err != nil {
			if explicitTargets[target] {
				failed = append(failed, fmt.Sprintf("%s: %v", target, err))
			}
			continue
		}
		outputs = append(outputs, compareOutput{value: result.Value, unit: target, approximate: result.Approximate})
	}
	if len(failed) > 0 {
		fmt.Fprintln(c.Err, "error: incompatible compare target:", strings.Join(failed, "; "))
		return 1
	}
	if len(outputs) == 0 {
		fmt.Fprintln(c.Err, "error: no compare preset units are compatible with input")
		return 1
	}
	if asJSON {
		payload := struct {
			Command string `json:"command"`
			Input   struct {
				Value float64 `json:"value"`
				Unit  string  `json:"unit"`
			} `json:"input"`
			Outputs []struct {
				Value       float64 `json:"value"`
				Unit        string  `json:"unit"`
				Approximate bool    `json:"approximate"`
			} `json:"outputs"`
		}{Command: "compare"}
		payload.Input.Value, payload.Input.Unit = value, inputUnit
		for _, output := range outputs {
			payload.Outputs = append(payload.Outputs, struct {
				Value       float64 `json:"value"`
				Unit        string  `json:"unit"`
				Approximate bool    `json:"approximate"`
			}{output.value, output.unit, output.approximate})
		}
		return c.encodeJSON(payload)
	}
	if len(outputs) == 1 {
		prefix := ""
		if outputs[0].approximate {
			prefix = "approximately "
		}
		fmt.Fprintf(c.Out, "%s %s = %s%s %s\n", units.FormatValue(value, 10, false), inputUnit, prefix, units.FormatValue(outputs[0].value, 10, false), outputs[0].unit)
		return 0
	}
	fmt.Fprintf(c.Out, "%s %s is:\n", units.FormatValue(value, 10, false), inputUnit)
	for _, output := range outputs {
		prefix := ""
		if output.approximate {
			prefix = "approximately "
		}
		fmt.Fprintf(c.Out, "  %s%s %s\n", prefix, units.FormatValue(output.value, 10, false), output.unit)
	}
	return 0
}

func parseCompareOperands(args []string) (float64, string, []string, error) {
	if len(args) < 1 {
		return 0, "", nil, fmt.Errorf("usage: convunits compare <valueunit> <target-unit>...")
	}
	if len(args) >= 2 {
		if value, err := strconv.ParseFloat(args[0], 64); err == nil {
			return value, args[1], args[2:], nil
		}
	}
	value, unit, err := splitValueUnit(args[0])
	if err != nil {
		return 0, "", nil, err
	}
	return value, unit, args[1:], nil
}

func (c *CLI) comparePresetTargets(value float64, inputUnit, preset string, noLimit bool) []string {
	switch preset {
	case "fun":
		targets := []string{"banana", "smoot", "footballfield", "marathon", "olympicpool", "earthcircumference"}
		return c.compatibleCompareTargets(value, inputUnit, targets, 0)
	case "human":
		return c.compatibleCompareTargets(value, inputUnit, []string{"banana", "hand", "cubit", "pace", "span", "smoot", "footballfield", "marathon"}, 0)
	case "astronomical":
		return c.compatibleCompareTargets(value, inputUnit, []string{"Re", "Rj", "Rsun", "LD", "au", "ls", "lightmin", "lh", "ld", "ly", "pc"}, 0)
	case "ancient":
		return c.compatibleCompareTargets(value, inputUnit, []string{"cubit", "royalcubit", "span", "handbreadth", "fingerbreadth", "pace", "romanfoot", "romanmile", "stadion", "parasang"}, 0)
	case "all":
		limit := 30
		if noLimit {
			limit = 0
		}
		return c.compatibleCompareTargets(value, inputUnit, c.allCompareTargets(), limit)
	default:
		return nil
	}
}

func (c *CLI) allCompareTargets() []string {
	unitsList := c.Registry.Units("")
	sort.SliceStable(unitsList, func(i, j int) bool {
		pi, pj := compareCategoryPriority(unitsList[i].Category), compareCategoryPriority(unitsList[j].Category)
		if pi != pj {
			return pi < pj
		}
		return unitsList[i].Symbol < unitsList[j].Symbol
	})
	var out []string
	for _, unit := range unitsList {
		if unit.Affine {
			continue
		}
		out = append(out, unit.Symbol)
	}
	return out
}

func compareCategoryPriority(category string) int {
	switch category {
	case "human-scale", "ancient-length", "historical-length", "nautical-length", "astronomical-length", "astronomical-mass", "astronomical-time":
		return 0
	case "length", "mass", "time", "area", "volume", "speed", "acceleration", "force", "energy", "power", "pressure", "flow":
		return 1
	default:
		return 2
	}
}

func (c *CLI) compatibleCompareTargets(value float64, inputUnit string, targets []string, limit int) []string {
	var out []string
	for _, target := range targets {
		if _, err := c.Registry.Convert(value, inputUnit, target); err == nil {
			out = append(out, target)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out
}

func (c *CLI) rejectCompareAffine(unit, role string) error {
	candidates, err := c.Registry.ParseCandidates(unit)
	if err != nil {
		return fmt.Errorf("%s unit: %w", role, err)
	}
	for _, candidate := range candidates {
		if candidate.Affine != nil {
			return fmt.Errorf("compare does not treat Celsius/Fahrenheit as ordinary scalar units")
		}
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}

func (c *CLI) compareHelp() {
	fmt.Fprint(c.Out, `convunits compare expresses one quantity in compatible target units.

Usage:
  convunits compare <valueunit> <target-unit>...
  convunits compare <value> <input-unit> <target-unit>...
  convunits compare <valueunit> --fun
  convunits compare <valueunit> --human
  convunits compare <valueunit> --astronomical
  convunits compare <valueunit> --ancient
  convunits compare <valueunit> --all [--no-limit]

Examples:
  convunits compare 38in banana smoot cubit Rj
  convunits compare 38 in banana smoot Rj
  convunits compare 1Rsun Rj Re LD
  convunits compare 60mph m/s km/h ft/s
  convunits --json compare 38in banana smoot Rj

Compare mode uses the normal unit parser/converter and remains dimensionally strict.
Preset units that do not match the input dimension are skipped.
`)
}

func (c *CLI) runEval(args []string, globalJSON bool) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		c.evalHelp()
		return 0
	}
	asJSON := globalJSON
	var parts []string
	for _, arg := range args {
		if arg == "--json" {
			asJSON = true
			continue
		}
		parts = append(parts, arg)
	}
	if len(parts) != 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits eval '<expression>'")
		return 2
	}
	expression := parts[0]
	result, err := evalcalc.New(c.Registry).Evaluate(expression)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	if asJSON {
		payload := struct {
			Command    string `json:"command"`
			Expression string `json:"expression"`
			Output     struct {
				Value       float64 `json:"value"`
				Unit        string  `json:"unit,omitempty"`
				Dimension   string  `json:"dimension,omitempty"`
				Approximate bool    `json:"approximate"`
			} `json:"output"`
		}{Command: "eval", Expression: expression}
		payload.Output.Value = result.Value
		payload.Output.Unit = result.Unit
		if result.Unit == "" {
			payload.Output.Dimension = result.Dimension.String()
		}
		payload.Output.Approximate = result.Approximate
		return c.encodeJSON(payload)
	}
	prefix := ""
	if result.Approximate {
		prefix = "approximately "
	}
	if result.Unit != "" {
		fmt.Fprintf(c.Out, "%s%s %s\n", prefix, units.FormatValue(result.Value, 10, false), result.Unit)
		return 0
	}
	if result.Dimension == (units.Dimension{}) {
		fmt.Fprintf(c.Out, "%s%s\n", prefix, units.FormatValue(result.Value, 10, false))
		return 0
	}
	fmt.Fprintf(c.Out, "%s%s %s\n", prefix, units.FormatValue(result.Value, 10, false), result.Dimension)
	return 0
}

func (c *CLI) evalHelp() {
	fmt.Fprint(c.Out, `convunits eval is a small unit-aware calculator.

Usage:
  convunits eval '<expression>'
  convunits eval '<expression> -> <output-unit>'

Examples:
  convunits eval '38in / Rj'
  convunits eval '1olympicpool / 1cup'
  convunits eval '2 * pi * 1Re -> km'
  convunits eval '0.5 * 1500kg * (60mph)^2 -> kWh'
  convunits eval '1kg * 9.80665m/s^2 -> N'
  convunits --json eval '38in / Rj'

Supported features are numbers, unit-attached numbers, +, -, *, /, ^,
parentheses, unary minus, and constants pi, c, G, and g0. Eval is intentionally
not a programming language: no variables, functions, loops, conditionals, or assignment.
Recipe ingredient conversions remain under "convunits recipe".
`)
}

func (c *CLI) runExplain(args []string, globalJSON bool) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		c.explainHelp()
		return 0
	}
	asJSON := globalJSON
	var positional []string
	for _, arg := range args {
		if arg == "--json" {
			asJSON = true
			continue
		}
		positional = append(positional, arg)
	}
	if len(positional) > 0 && unsupportedExplainCommand(positional[0]) {
		fmt.Fprintf(c.Err, "error: explain does not support %s yet\n", positional[0])
		return 1
	}
	var ex explain.Explanation
	var err error
	if len(positional) == 1 && strings.Contains(positional[0], "->") {
		ex, err = explain.Eval(c.Registry, positional[0])
	} else {
		var value float64
		var inputUnit, outputUnit string
		value, inputUnit, outputUnit, err = parseOperands(positional)
		if err == nil {
			ex, err = explain.Normal(c.Registry, value, inputUnit, outputUnit)
		}
	}
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	if asJSON {
		type explainInput struct {
			Value float64 `json:"value"`
			Unit  string  `json:"unit"`
		}
		payload := struct {
			Command    string        `json:"command"`
			Expression string        `json:"expression,omitempty"`
			Input      *explainInput `json:"input,omitempty"`
			Output     struct {
				Value       float64 `json:"value"`
				Unit        string  `json:"unit"`
				Approximate bool    `json:"approximate"`
			} `json:"output"`
			Steps      []string `json:"steps"`
			Dimensions struct {
				Input  string `json:"input,omitempty"`
				Output string `json:"output,omitempty"`
			} `json:"dimensions,omitempty"`
		}{Command: "explain", Expression: ex.Expression, Steps: ex.Steps}
		if ex.Input.Unit != "" {
			payload.Input = &explainInput{Value: ex.Input.Value, Unit: ex.Input.Unit}
		}
		payload.Output.Value, payload.Output.Unit, payload.Output.Approximate = ex.Output.Value, ex.Output.Unit, ex.Output.Approximate
		payload.Dimensions.Input, payload.Dimensions.Output = ex.Dimensions.Input, ex.Dimensions.Output
		return c.encodeJSON(payload)
	}
	fmt.Fprint(c.Out, explain.Text(ex))
	return 0
}

func unsupportedExplainCommand(command string) bool {
	switch command {
	case "recipe", "scale", "shoe", "paper", "size", "scale-size", "wire", "drill", "sieve", "formula", "compare":
		return true
	default:
		return false
	}
}

func (c *CLI) explainHelp() {
	fmt.Fprint(c.Out, `convunits explain shows how a conversion or eval expression is derived.

Usage:
  convunits explain <valueunit> <output-unit>
  convunits explain <value> <input-unit> <output-unit>
  convunits explain '<eval-expression> -> <output-unit>'

Examples:
  convunits explain 60mph m/s
  convunits explain 1N 'kg*m/s^2'
  convunits explain '2*pi*1Re -> km'
  convunits explain '0.5 * 1500kg * (60mph)^2 -> kWh'
  convunits --json explain 60mph m/s

Explain currently supports normal conversions and eval expressions with ->.
`)
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

func parseJSONOnly(args []string) (bool, []string) {
	asJSON := false
	var positional []string
	for _, arg := range args {
		if arg == "--json" {
			asJSON = true
			continue
		}
		positional = append(positional, arg)
	}
	return asJSON, positional
}

func (c *CLI) encodeJSON(payload any) int {
	enc := json.NewEncoder(c.Out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) runRecipe(args []string, globalJSON bool) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		c.recipeHelp()
		return 0
	}
	if args[0] == "ingredients" {
		return c.listRecipeIngredients(args[1:])
	}
	asJSON := globalJSON
	var positional []string
	for _, arg := range args {
		if arg == "--json" {
			asJSON = true
			continue
		}
		if strings.HasPrefix(arg, "--") {
			fmt.Fprintf(c.Err, "error: unknown recipe option %q\n", arg)
			return 2
		}
		positional = append(positional, arg)
	}
	value, inputUnit, ingredientName, outputUnit, err := parseRecipeOperands(positional)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 2
	}
	ingredient, err := recipe.Lookup(ingredientName)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	inputKind, err := c.recipeUnitKind(inputUnit)
	if err != nil {
		fmt.Fprintln(c.Err, "error: input unit:", err)
		return 1
	}
	outputKind, err := c.recipeUnitKind(outputUnit)
	if err != nil {
		fmt.Fprintln(c.Err, "error: output unit:", err)
		return 1
	}
	if ingredient.DensityKgPerM3 <= 0 && inputKind != outputKind {
		fmt.Fprintf(c.Err, "error: density missing for ingredient %q\n", ingredient.Key)
		return 1
	}
	var outputValue float64
	switch {
	case inputKind == outputKind:
		converted, err := c.Registry.Convert(value, inputUnit, outputUnit)
		if err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		outputValue = converted.Value
	case inputKind == "volume" && outputKind == "mass":
		volume, err := c.Registry.Convert(value, inputUnit, "m^3")
		if err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		converted, err := c.Registry.Convert(volume.Value*ingredient.DensityKgPerM3, "kg", outputUnit)
		if err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		outputValue = converted.Value
	case inputKind == "mass" && outputKind == "volume":
		mass, err := c.Registry.Convert(value, inputUnit, "kg")
		if err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		converted, err := c.Registry.Convert(mass.Value/ingredient.DensityKgPerM3, "m^3", outputUnit)
		if err != nil {
			fmt.Fprintln(c.Err, "error:", err)
			return 1
		}
		outputValue = converted.Value
	}
	if asJSON {
		payload := struct {
			Command string `json:"command"`
			Input   struct {
				Value      float64 `json:"value"`
				Unit       string  `json:"unit"`
				Ingredient string  `json:"ingredient"`
			} `json:"input"`
			Output struct {
				Value       float64 `json:"value"`
				Unit        string  `json:"unit"`
				Ingredient  string  `json:"ingredient"`
				Approximate bool    `json:"approximate"`
			} `json:"output"`
			Density struct {
				Value float64 `json:"value"`
				Unit  string  `json:"unit"`
			} `json:"density"`
		}{Command: "recipe"}
		payload.Input.Value, payload.Input.Unit, payload.Input.Ingredient = value, inputUnit, ingredientName
		payload.Output.Value, payload.Output.Unit, payload.Output.Ingredient, payload.Output.Approximate = outputValue, outputUnit, ingredient.Name, true
		payload.Density.Value, payload.Density.Unit = ingredient.DensityKgPerM3, "kg/m^3"
		return c.encodeJSON(payload)
	}
	fmt.Fprintf(c.Out, "approximately %s %s %s\n", units.FormatValue(outputValue, 10, false), outputUnit, ingredient.Name)
	return 0
}

func parseRecipeOperands(args []string) (float64, string, string, string, error) {
	if len(args) == 3 {
		value, unit, err := splitValueUnit(args[0])
		if err != nil {
			return 0, "", "", "", err
		}
		return value, unit, args[1], args[2], nil
	}
	if len(args) == 4 {
		value, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return 0, "", "", "", fmt.Errorf("invalid recipe amount %q", args[0])
		}
		return value, args[1], args[2], args[3], nil
	}
	return 0, "", "", "", fmt.Errorf("usage: convunits recipe <amount><unit> <ingredient> <output-unit>")
}

func (c *CLI) recipeUnitKind(unit string) (string, error) {
	if _, err := c.Registry.Convert(1, unit, "kg"); err == nil {
		return "mass", nil
	}
	if _, err := c.Registry.Convert(1, unit, "m^3"); err == nil {
		return "volume", nil
	}
	if _, err := c.Registry.ParseCandidates(unit); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%q is not a mass or volume unit", unit)
}

func (c *CLI) listRecipeIngredients(args []string) int {
	if len(args) > 1 {
		fmt.Fprintln(c.Err, "error: usage: convunits recipe ingredients [category]")
		return 2
	}
	category := ""
	if len(args) == 1 {
		category = args[0]
	}
	list := recipe.Ingredients(category)
	if len(list) == 0 {
		fmt.Fprintf(c.Err, "error: unknown or empty recipe ingredient category %q\n", category)
		return 1
	}
	w := tabwriter.NewWriter(c.Out, 0, 4, 2, ' ', 0)
	for _, ingredient := range list {
		aliases := strings.Join(ingredient.Aliases, ", ")
		if aliases == "" {
			aliases = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", ingredient.Key, ingredient.Name, ingredient.Category, aliases, ingredient.Note)
	}
	if err := w.Flush(); err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return 0
}

func (c *CLI) recipeHelp() {
	fmt.Fprint(c.Out, `convunits recipe converts approximate cooking ingredient amounts.

Usage:
  convunits recipe <amount><unit> <ingredient> <output-unit>
  convunits recipe <amount> <unit> <ingredient> <output-unit>
  convunits recipe ingredients [category]

Examples:
  convunits recipe 1cup flour g
  convunits recipe 2tbsp butter g
  convunits recipe 100g sugar cup
  convunits recipe 500ml water lb
  convunits --json recipe 1cup flour g

Recipe conversions are approximate and ingredient-specific. Density data is not part
of the normal unit registry.
`)
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

func (c *CLI) runSize(args []string, globalJSON bool) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		c.paperHelp()
		return 0
	}
	asJSON, positional := parseJSONOnly(args)
	asJSON = asJSON || globalJSON
	if len(positional) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits size PAPER-SIZE OUTPUT-LENGTH-UNIT")
		return 2
	}
	scaleRegistry := scales.NewRegistry()
	if scaleRegistry.IsScale(positional[0]) {
		fmt.Fprintf(c.Err, "error: scalar scale %q cannot be used as a paper size; use convunits scale\n", positional[0])
		return 1
	}
	size, err := scales.LookupPaperSize(positional[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	width, err := c.Registry.Convert(size.WidthM, "m", positional[1])
	if err != nil {
		fmt.Fprintln(c.Err, "error: paper output unit must be length:", err)
		return 1
	}
	height, err := c.Registry.Convert(size.HeightM, "m", positional[1])
	if err != nil {
		fmt.Fprintln(c.Err, "error: paper output unit must be length:", err)
		return 1
	}
	if asJSON {
		payload := struct {
			Command string `json:"command"`
			Input   struct {
				Size string `json:"size"`
			} `json:"input"`
			Output struct {
				Width  float64 `json:"width"`
				Height float64 `json:"height"`
				Unit   string  `json:"unit"`
			} `json:"output"`
		}{Command: "paper"}
		payload.Input.Size = positional[0]
		payload.Output.Width, payload.Output.Height, payload.Output.Unit = width.Value, height.Value, positional[1]
		return c.encodeJSON(payload)
	}
	prefix := ""
	if size.Approximate {
		prefix = "approximately "
	}
	fmt.Fprintf(c.Out, "%s%s x %s %s\n", prefix, units.FormatValue(width.Value, 10, false), units.FormatValue(height.Value, 10, false), positional[1])
	return 0
}

func (c *CLI) paperHelp() {
	fmt.Fprint(c.Out, `convunits paper looks up paper dimensions.

Usage:
  convunits paper PAPER-SIZE OUTPUT-LENGTH-UNIT
  convunits size PAPER-SIZE OUTPUT-LENGTH-UNIT
  convunits papers [iso|us|photo]

Examples:
  convunits paper a4 mm
  convunits paper letter in
  convunits --json paper a4 mm
`)
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

func (c *CLI) runWire(args []string, globalJSON bool) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		c.wireHelp()
		return 0
	}
	asJSON, positional := parseJSONOnly(args)
	asJSON = asJSON || globalJSON
	if len(positional) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits wire GAUGE OUTPUT-LENGTH-UNIT")
		return 2
	}
	gauge, err := wire.ParseGauge(positional[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	meters, err := wire.DiameterMeters(gauge)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return c.printLengthResult(meters, positional[1], "wire", positional[0], "diameter", true, asJSON)
}

func (c *CLI) wireHelp() {
	fmt.Fprint(c.Out, `convunits wire converts AWG gauge to approximate conductor diameter.

Usage:
  convunits wire GAUGE OUTPUT-LENGTH-UNIT
  convunits wires

Examples:
  convunits wire 12awg mm
  convunits wire 0000awg in
  convunits --json wire 12awg mm
`)
}

func (c *CLI) runDrill(args []string, globalJSON bool) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		c.drillHelp()
		return 0
	}
	asJSON, positional := parseJSONOnly(args)
	asJSON = asJSON || globalJSON
	if len(positional) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits drill SIZE OUTPUT-LENGTH-UNIT")
		return 2
	}
	meters, err := drill.DiameterMeters(positional[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return c.printLengthResult(meters, positional[1], "drill", positional[0], "diameter", true, asJSON)
}

func (c *CLI) drillHelp() {
	fmt.Fprint(c.Out, `convunits drill looks up drill bit diameter.

Usage:
  convunits drill SIZE OUTPUT-LENGTH-UNIT
  convunits drills [number|letter|fractional]

Examples:
  convunits drill '#7' mm
  convunits drill '1/4' mm
  convunits --json drill '#7' mm
`)
}

func (c *CLI) runSieve(args []string, globalJSON bool) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		c.sieveHelp()
		return 0
	}
	asJSON, positional := parseJSONOnly(args)
	asJSON = asJSON || globalJSON
	if len(positional) != 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits sieve SIZE OUTPUT-LENGTH-UNIT")
		return 2
	}
	meters, err := sieve.OpeningMeters(positional[0])
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	return c.printLengthResult(meters, positional[1], "sieve", positional[0], "opening", true, asJSON)
}

func (c *CLI) sieveHelp() {
	fmt.Fprint(c.Out, `convunits sieve looks up approximate nominal sieve openings.

Usage:
  convunits sieve SIZE OUTPUT-LENGTH-UNIT
  convunits sieves

Examples:
  convunits sieve 'No. 200' um
  convunits sieve '#40' mm
  convunits --json sieve 'No. 200' um
`)
}

func (c *CLI) printLengthResult(meters float64, outputUnit, command, input, label string, approximate, asJSON bool) int {
	converted, err := c.Registry.Convert(meters, "m", outputUnit)
	if err != nil {
		fmt.Fprintf(c.Err, "error: output unit for %s must be length: %v\n", label, err)
		return 1
	}
	if asJSON {
		payload := struct {
			Command string `json:"command"`
			Input   struct {
				Size   string `json:"size,omitempty"`
				Gauge  string `json:"gauge,omitempty"`
				System string `json:"system,omitempty"`
			} `json:"input"`
			Output struct {
				Value       float64 `json:"value"`
				Unit        string  `json:"unit"`
				Quantity    string  `json:"quantity"`
				Approximate bool    `json:"approximate"`
			} `json:"output"`
		}{Command: command}
		if command == "wire" {
			payload.Input.Gauge = input
			payload.Input.System = "AWG"
		} else {
			payload.Input.Size = input
		}
		payload.Output.Value = converted.Value
		payload.Output.Unit = outputUnit
		payload.Output.Quantity = label
		payload.Output.Approximate = approximate
		return c.encodeJSON(payload)
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

func (c *CLI) runFormula(args []string, globalJSON bool) int {
	asJSON := globalJSON
	if len(args) > 0 && args[0] == "--json" {
		asJSON = true
		args = args[1:]
	}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		c.formulaHelp()
		return 0
	}
	if len(args) < 2 {
		fmt.Fprintln(c.Err, "error: usage: convunits formula NAME [--ARG VALUEUNIT ...] OUTPUT-UNIT")
		return 2
	}
	name := args[0]
	rawInputs := make(map[string]string)
	var positional []string
	allowed := map[string]bool{
		"area": true, "distance": true, "energy": true, "force": true, "gravity": true,
		"height": true, "length": true, "mass": true, "mass1": true, "mass2": true,
		"power": true, "radius": true, "speed": true, "time": true, "volume": true,
	}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			positional = append(positional, arg)
			continue
		}
		if arg == "--json" {
			asJSON = true
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
	if asJSON {
		payload := struct {
			Command string            `json:"command"`
			Formula string            `json:"formula"`
			Inputs  map[string]string `json:"inputs"`
			Output  struct {
				Value float64 `json:"value"`
				Unit  string  `json:"unit"`
			} `json:"output"`
		}{Command: "formula", Formula: name, Inputs: rawInputs}
		payload.Output.Value, payload.Output.Unit = result.Value, result.Unit
		return c.encodeJSON(payload)
	}
	prefix := ""
	if result.Approximate {
		prefix = "approximately "
	}
	fmt.Fprintf(c.Out, "%s%s %s\n", prefix, units.FormatValue(result.Value, 10, false), result.Unit)
	return 0
}

func (c *CLI) formulaHelp() {
	fmt.Fprint(c.Out, `convunits formula computes named formulas with unit-checked inputs.

Usage:
  convunits formula NAME [--ARG VALUEUNIT ...] OUTPUT-UNIT
  convunits formulas

Examples:
  convunits formula escape-velocity --mass 1Mearth --radius 1Re km/s
  convunits formula schwarzschild-radius --mass 1Msun km
  convunits formula bmi --mass 180lb --height 6ft bmi

Use "convunits formulas" to list available formula names and arguments.
`)
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

func (c *CLI) runShoe(args []string, globalJSON bool) int {
	asJSON, positional := parseJSONOnly(args)
	asJSON = asJSON || globalJSON
	if len(positional) == 1 && (positional[0] == "--help" || positional[0] == "-h") {
		c.shoeHelp()
		return 0
	}
	if len(positional) == 1 && positional[0] == "systems" {
		for _, system := range shoes.Systems() {
			fmt.Fprintln(c.Out, system)
		}
		return 0
	}
	if len(positional) != 3 {
		fmt.Fprintln(c.Err, "error: usage: convunits shoe SYSTEM SIZE OUTPUT-LENGTH-UNIT\n       convunits shoe systems")
		return 2
	}
	size, err := strconv.ParseFloat(positional[1], 64)
	if err != nil {
		fmt.Fprintf(c.Err, "error: invalid shoe size %q\n", positional[1])
		return 2
	}
	meters, err := shoes.FootLengthMeters(positional[0], size)
	if err != nil {
		fmt.Fprintln(c.Err, "error:", err)
		return 1
	}
	converted, err := c.Registry.Convert(meters, "m", positional[2])
	if err != nil {
		fmt.Fprintln(c.Err, "error: shoe output unit must be length:", err)
		return 1
	}
	if asJSON {
		payload := struct {
			Command string `json:"command"`
			Input   struct {
				System string  `json:"system"`
				Size   float64 `json:"size"`
			} `json:"input"`
			Output struct {
				Value       float64 `json:"value"`
				Unit        string  `json:"unit"`
				Quantity    string  `json:"quantity"`
				Approximate bool    `json:"approximate"`
			} `json:"output"`
		}{Command: "shoe"}
		payload.Input.System, payload.Input.Size = positional[0], size
		payload.Output.Value, payload.Output.Unit, payload.Output.Quantity, payload.Output.Approximate = converted.Value, positional[2], "foot length", true
		return c.encodeJSON(payload)
	}
	fmt.Fprintf(c.Out, "approximately %s %s foot length\n", units.FormatValue(converted.Value, 10, false), positional[2])
	return 0
}

func (c *CLI) shoeHelp() {
	fmt.Fprint(c.Out, `convunits shoe estimates foot length from a shoe-size system.

Usage:
  convunits shoe SYSTEM SIZE OUTPUT-LENGTH-UNIT
  convunits shoe systems

Examples:
  convunits shoe us-men 10 yd
  convunits shoe eu 43 cm
  convunits --json shoe us-men 10 yd

Shoe conversions are approximate foot-length mappings, not fit recommendations.
`)
}

func (c *CLI) runSolve(args []string, globalJSON bool) int {
	precision := 10
	scientific := false
	asJSON := globalJSON
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
  convunits search QUERY
  convunits aliases UNIT-OR-ALIAS
  convunits compare <valueunit> <target-unit>...
  convunits eval '<expression>'
  convunits explain <valueunit> <output-unit>
  convunits solve VALUEUNIT OUTPUTUNIT --given NAME=VALUEUNIT ...
  convunits scale VALUE INPUT-SCALE OUTPUT-SCALE
  convunits scales [category]
  convunits recipe <amount><unit> <ingredient> <output-unit>
  convunits recipe ingredients [category]
  convunits size PAPER-SIZE OUTPUT-LENGTH-UNIT   alias for paper
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
  convunits 60mph km/h
  convunits 1N 'kg*m/s^2'
  convunits 1Rsun km
  convunits search jupiter
  convunits aliases Rj
  convunits compare 38in banana smoot Rj
  convunits eval '2*pi*1Re -> km'
  convunits explain 60mph m/s
  convunits 100F C
  convunits 30mpg L/100km
  convunits solve 10N s --given mass=2kg --given distance=5m
  convunits scale 5 beaufort m/s
  convunits recipe 1cup flour g
  convunits paper a4 mm
  convunits shoe us-men 10 yd
  convunits wire 12awg mm
  convunits drill '#7' mm
  convunits sieve 'No. 200' um
  convunits formula escape-velocity --mass 1Mearth --radius 1Re km/s
  convunits --json 10kg lb

Unit expressions support *, /, integer powers, and parentheses. Parsing is case-sensitive.
Use "convunits search", "convunits units", "convunits scales", and "convunits formulas" for listings.
See README.md and SUPPORTED_UNITS.md for detailed examples and limitations.
`)
}
