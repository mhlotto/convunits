package discovery

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	formulacalc "convunits/internal/formula"
	"convunits/internal/recipe"
	"convunits/internal/scales"
	"convunits/internal/units"
	"convunits/internal/weird/drill"
	"convunits/internal/weird/sieve"
	"convunits/internal/weird/wire"
)

type Result struct {
	Kind        string   `json:"kind"`
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Category    string   `json:"category,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Description string   `json:"description,omitempty"`
	Approximate bool     `json:"approximate,omitempty"`
	Dimension   string   `json:"dimension,omitempty"`
	Score       int      `json:"-"`
}

type AliasMatch struct {
	Kind          string   `json:"kind"`
	Key           string   `json:"key"`
	Canonical     string   `json:"canonical,omitempty"`
	Name          string   `json:"name"`
	Category      string   `json:"category,omitempty"`
	Aliases       []string `json:"aliases,omitempty"`
	Dimension     string   `json:"dimension,omitempty"`
	Description   string   `json:"description,omitempty"`
	Approximate   bool     `json:"approximate,omitempty"`
	DensityValue  float64  `json:"density_value,omitempty"`
	DensityUnit   string   `json:"density_unit,omitempty"`
	MatchedBy     string   `json:"matched_by,omitempty"`
	MatchedString string   `json:"matched_string,omitempty"`
	Score         int      `json:"-"`
}

type Catalog struct {
	registry *units.Registry
	results  []Result
}

func New(registry *units.Registry) *Catalog {
	c := &Catalog{registry: registry}
	c.results = c.build()
	return c
}

func (c *Catalog) Search(query, kind string, all bool) []Result {
	query = strings.TrimSpace(query)
	kind = strings.TrimSpace(kind)
	numeric := isNumericLookingQuery(query)
	var out []Result
	for _, result := range c.results {
		if kind != "" && !strings.EqualFold(result.Kind, kind) {
			continue
		}
		score := matchScore(query, result, all, numeric)
		if score == 0 {
			continue
		}
		result.Score = score
		out = append(out, result)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return strings.ToLower(out[i].Key) < strings.ToLower(out[j].Key)
	})
	if !all && len(out) > 20 {
		out = out[:20]
	}
	return out
}

func (c *Catalog) Aliases(query string) []AliasMatch {
	query = strings.TrimSpace(query)
	var out []AliasMatch
	for _, match := range c.allAliasMatches() {
		if aliasMatches(query, match) {
			out = append(out, markAliasMatch(query, match))
		}
	}
	sortAliasMatches(out)
	return out
}

func (c *Catalog) AllAliases() []AliasMatch {
	out := c.allAliasMatches()
	sortAliasMatches(out)
	return out
}

func (c *Catalog) build() []Result {
	var out []Result
	for _, unit := range c.registry.Units("") {
		description := unit.Note
		if description == "" {
			description = unit.SourceNote
		}
		if description == "" {
			description = "dimensions: " + unit.Dimension.String()
		}
		out = append(out, Result{
			Kind:        "unit",
			Key:         unit.Symbol,
			Name:        unit.Name,
			Category:    unit.Category,
			Aliases:     append([]string(nil), unit.Aliases...),
			Description: description,
			Approximate: unit.Approximate,
			Dimension:   unit.Dimension.String(),
		})
	}
	for _, scale := range scales.NewRegistry().Scales("") {
		description := scale.Note
		if description == "" && scale.Base != "" {
			description = "scale family: " + scale.Base
		}
		out = append(out, Result{
			Kind:        "scale",
			Key:         scale.Symbol,
			Name:        scale.Name,
			Category:    scale.Category,
			Aliases:     append([]string(nil), scale.Aliases...),
			Description: description,
		})
	}
	for _, ingredient := range recipe.Ingredients("") {
		description := ingredient.Note
		if description == "" {
			description = fmt.Sprintf("density: %g kg/m^3", ingredient.DensityKgPerM3)
		}
		out = append(out, Result{
			Kind:        "ingredient",
			Key:         ingredient.Key,
			Name:        ingredient.Name,
			Category:    ingredient.Category,
			Aliases:     append([]string(nil), ingredient.Aliases...),
			Description: description,
			Approximate: true,
		})
	}
	for _, definition := range formulacalc.Definitions() {
		out = append(out, Result{
			Kind:        "formula",
			Key:         definition.Name,
			Name:        definition.Name,
			Category:    definition.OutputDimension,
			Description: definition.Arguments,
			Approximate: true,
		})
	}
	for _, paper := range scales.PaperSizes("") {
		out = append(out, Result{
			Kind:        "paper",
			Key:         paper.Symbol,
			Name:        paper.Name,
			Category:    paper.Category,
			Description: fmt.Sprintf("%.1f x %.1f mm", paper.WidthM*1000, paper.HeightM*1000),
			Approximate: paper.Approximate,
		})
	}
	if entries, err := drill.Entries(""); err == nil {
		for _, entry := range entries {
			out = append(out, Result{
				Kind:        "drill",
				Key:         entry.Size,
				Name:        "drill size " + entry.Size,
				Category:    entry.Category,
				Description: fmt.Sprintf("%.4g in diameter", entry.DiameterIn),
				Approximate: true,
			})
		}
	}
	for gauge := -3; gauge <= 40; gauge++ {
		key := wireGaugeKey(gauge)
		aliases := []string{"awg" + strings.TrimSuffix(key, "awg")}
		description := "AWG wire gauge"
		if diameter, err := wire.DiameterMeters(gauge); err == nil {
			description = fmt.Sprintf("%.6g mm diameter", diameter*1000)
		}
		out = append(out, Result{
			Kind:        "wire",
			Key:         key,
			Name:        "AWG " + strings.TrimSuffix(key, "awg") + " wire gauge",
			Category:    "wire",
			Aliases:     aliases,
			Description: description,
			Approximate: true,
		})
	}
	for _, entry := range sieve.Entries() {
		out = append(out, Result{
			Kind:        "sieve",
			Key:         entry.Size,
			Name:        "sieve opening " + entry.Size,
			Category:    "sieve",
			Description: fmt.Sprintf("%.4g mm opening", entry.OpeningMM),
			Approximate: true,
		})
		if strings.HasPrefix(entry.Size, "#") {
			number := strings.TrimPrefix(entry.Size, "#")
			out = append(out, Result{
				Kind:        "sieve",
				Key:         "No. " + number,
				Name:        "sieve opening No. " + number,
				Category:    "sieve",
				Aliases:     []string{entry.Size, number + "mesh"},
				Description: fmt.Sprintf("%.4g mm opening", entry.OpeningMM),
				Approximate: true,
			})
		}
	}
	out = append(out, commandResults()...)
	return out
}

func (c *Catalog) allAliasMatches() []AliasMatch {
	var out []AliasMatch
	for _, unit := range c.registry.Units("") {
		description := unit.Note
		if description == "" {
			description = unit.SourceNote
		}
		out = append(out, AliasMatch{
			Kind:        "unit",
			Key:         unit.Symbol,
			Name:        unit.Name,
			Category:    unit.Category,
			Aliases:     append([]string(nil), unit.Aliases...),
			Dimension:   unit.Dimension.String(),
			Description: description,
			Approximate: unit.Approximate,
		})
	}
	for _, scale := range scales.NewRegistry().Scales("") {
		out = append(out, AliasMatch{
			Kind:        "scale",
			Key:         scale.Symbol,
			Name:        scale.Name,
			Category:    scale.Category,
			Aliases:     append([]string(nil), scale.Aliases...),
			Description: scale.Note,
		})
	}
	for _, ingredient := range recipe.Ingredients("") {
		out = append(out, AliasMatch{
			Kind:         "ingredient",
			Key:          ingredient.Key,
			Canonical:    ingredient.Key,
			Name:         ingredient.Name,
			Category:     ingredient.Category,
			Aliases:      append([]string(nil), ingredient.Aliases...),
			Description:  ingredient.Note,
			Approximate:  true,
			DensityValue: ingredient.DensityKgPerM3,
			DensityUnit:  "kg/m^3",
		})
	}
	for _, definition := range formulacalc.Definitions() {
		out = append(out, AliasMatch{
			Kind:        "formula",
			Key:         definition.Name,
			Name:        definition.Name,
			Category:    definition.OutputDimension,
			Description: definition.Arguments,
			Approximate: true,
		})
	}
	for _, paper := range scales.PaperSizes("") {
		out = append(out, AliasMatch{
			Kind:        "paper",
			Key:         paper.Symbol,
			Name:        paper.Name,
			Category:    paper.Category,
			Description: fmt.Sprintf("%.1f x %.1f mm", paper.WidthM*1000, paper.HeightM*1000),
		})
	}
	out = append(out, commandAliasMatches()...)
	return out
}

func matchScore(query string, result Result, all bool, numeric bool) int {
	q := strings.ToLower(query)
	if q == "" {
		return 0
	}
	if numeric && !all {
		return numericMatchScore(q, result)
	}
	best := scoreField(q, result.Key, 100+kindExactBoost(result.Kind), 90, 60)
	best = max(best, scoreField(q, result.Name, 95, 80, 50))
	for _, alias := range result.Aliases {
		best = max(best, scoreField(q, alias, 98, 82, 55))
	}
	best = max(best, scoreField(q, result.Category, 70, 45, 30))
	best = max(best, scoreField(q, result.Description, 65, 40, 25))
	best = max(best, scoreField(q, result.Dimension, 60, 35, 20))
	return best
}

func numericMatchScore(query string, result Result) int {
	best := scoreNumericField(query, result.Key, 100)
	for _, alias := range result.Aliases {
		best = max(best, scoreNumericField(query, alias, 98))
	}
	return best
}

func scoreNumericField(query, field string, exact int) int {
	if strings.EqualFold(field, query) {
		return exact
	}
	queryForms := normalizedLookupForms(query)
	fieldForms := normalizedLookupForms(field)
	for _, q := range queryForms {
		for _, f := range fieldForms {
			if q == f {
				return exact - 5
			}
		}
	}
	return 0
}

func scoreField(query, field string, exact, prefix, substring int) int {
	f := strings.ToLower(field)
	switch {
	case f == "":
		return 0
	case f == query:
		return exact
	case strings.HasPrefix(f, query):
		return prefix
	case strings.Contains(f, query):
		return substring
	default:
		return 0
	}
}

var (
	plainIntegerRE = regexp.MustCompile(`^[0-9]+$`)
	decimalRE      = regexp.MustCompile(`^[0-9]*\.[0-9]+$|^[0-9]+\.[0-9]*$`)
	fractionRE     = regexp.MustCompile(`^[0-9]+/[0-9]+$`)
	numberedRE     = regexp.MustCompile(`^#\s*[0-9]+$|^no\.\s*[0-9]+$`)
	meshRE         = regexp.MustCompile(`^[0-9]+mesh$|^mesh[0-9]+$`)
	awgRE          = regexp.MustCompile(`^[0-9]+awg$|^awg[0-9]+$`)
)

func isNumericLookingQuery(query string) bool {
	q := normalizeCompact(query)
	return plainIntegerRE.MatchString(q) ||
		decimalRE.MatchString(q) ||
		fractionRE.MatchString(q) ||
		numberedRE.MatchString(q) ||
		meshRE.MatchString(q) ||
		awgRE.MatchString(q)
}

func normalizedLookupForms(text string) []string {
	q := normalizeCompact(text)
	if q == "" {
		return nil
	}
	forms := []string{"raw:" + q}
	add := func(value string) {
		if value == "" {
			return
		}
		for _, existing := range forms {
			if existing == value {
				return
			}
		}
		forms = append(forms, value)
	}
	if strings.HasPrefix(q, "#") {
		number := strings.TrimPrefix(q, "#")
		add("numbered:" + number)
	}
	if strings.HasPrefix(q, "no.") {
		number := strings.TrimPrefix(q, "no.")
		add("numbered:" + number)
	}
	if strings.HasSuffix(q, "mesh") {
		number := strings.TrimSuffix(q, "mesh")
		add("numbered:" + number)
	}
	if strings.HasPrefix(q, "mesh") {
		number := strings.TrimPrefix(q, "mesh")
		add("numbered:" + number)
	}
	if strings.HasSuffix(q, "awg") {
		number := strings.TrimSuffix(q, "awg")
		add("wire:" + number)
	}
	if strings.HasPrefix(q, "awg") {
		number := strings.TrimPrefix(q, "awg")
		add("wire:" + number)
	}
	if plainIntegerRE.MatchString(q) {
		add("numbered:" + q)
	}
	return forms
}

func normalizeCompact(text string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(text), " ", ""))
}

func aliasMatches(query string, match AliasMatch) bool {
	q := strings.ToLower(query)
	if q == "" {
		return false
	}
	if strings.EqualFold(match.Key, query) || strings.EqualFold(match.Name, query) || strings.EqualFold(match.Canonical, query) {
		return true
	}
	for _, alias := range match.Aliases {
		if strings.EqualFold(alias, query) {
			return true
		}
	}
	return false
}

func markAliasMatch(query string, match AliasMatch) AliasMatch {
	switch {
	case strings.EqualFold(match.Key, query):
		match.MatchedBy, match.MatchedString = "key", match.Key
		match.Score = 100 + kindExactBoost(match.Kind)
	case strings.EqualFold(match.Canonical, query):
		match.MatchedBy, match.MatchedString = "canonical", match.Canonical
		match.Score = 95 + kindExactBoost(match.Kind)
	case strings.EqualFold(match.Name, query):
		match.MatchedBy, match.MatchedString = "name", match.Name
		match.Score = 80
	default:
		for _, alias := range match.Aliases {
			if strings.EqualFold(alias, query) {
				match.MatchedBy, match.MatchedString = "alias", alias
				match.Score = 90 + kindExactBoost(match.Kind)
				break
			}
		}
	}
	if match.Score == 0 {
		match.Score = kindSortRank(match.Kind)
	}
	return match
}

func sortAliasMatches(matches []AliasMatch) {
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		if matches[i].Kind != matches[j].Kind {
			return kindSortRank(matches[i].Kind) > kindSortRank(matches[j].Kind)
		}
		return strings.ToLower(matches[i].Key) < strings.ToLower(matches[j].Key)
	})
}

func kindExactBoost(kind string) int {
	switch kind {
	case "unit":
		return 20
	case "scale":
		return 5
	default:
		return 0
	}
}

func kindSortRank(kind string) int {
	switch kind {
	case "unit":
		return 100
	case "scale":
		return 90
	case "ingredient":
		return 80
	case "formula":
		return 70
	case "paper", "wire", "drill", "sieve":
		return 60
	case "command":
		return 50
	default:
		return 0
	}
}

func commandResults() []Result {
	var out []Result
	for _, command := range commands() {
		out = append(out, Result{
			Kind:        "command",
			Key:         command.key,
			Name:        command.name,
			Category:    "command",
			Aliases:     command.aliases,
			Description: command.description,
		})
	}
	return out
}

func commandAliasMatches() []AliasMatch {
	var out []AliasMatch
	for _, command := range commands() {
		out = append(out, AliasMatch{
			Kind:        "command",
			Key:         command.key,
			Name:        command.name,
			Category:    "command",
			Aliases:     command.aliases,
			Description: command.description,
		})
	}
	return out
}

func wireGaugeKey(gauge int) string {
	switch gauge {
	case -3:
		return "0000awg"
	case -2:
		return "000awg"
	case -1:
		return "00awg"
	default:
		return fmt.Sprintf("%dawg", gauge)
	}
}

type commandInfo struct {
	key, name, description string
	aliases                []string
}

func commands() []commandInfo {
	return []commandInfo{
		{"units", "units", "list normal dimensional units", nil},
		{"solve", "solve", "solve for one variable in common formulas", nil},
		{"scale", "scale", "nonlinear and lookup scale conversions", []string{"scales"}},
		{"shoe", "shoe", "approximate shoe size foot-length mappings", nil},
		{"paper", "paper", "paper size dimensions", []string{"size", "papers"}},
		{"size", "size", "alias for paper", []string{"paper"}},
		{"wire", "wire", "wire gauge diameter lookup", []string{"wires", "awg"}},
		{"drill", "drill", "drill bit size lookup", []string{"drills"}},
		{"sieve", "sieve", "sieve opening lookup", []string{"sieves", "mesh"}},
		{"formula", "formula", "formula mode calculations", []string{"formulas"}},
		{"compare", "compare", "present one quantity in compatible units", nil},
		{"recipe", "recipe", "approximate cooking ingredient conversions", []string{"ingredients"}},
		{"eval", "eval", "unit-aware calculator", nil},
		{"explain", "explain", "show how a conversion or eval result is derived", nil},
		{"search", "search", "search units, commands, formulas, scales, and lookups", nil},
		{"aliases", "aliases", "show aliases and metadata for catalog entries", nil},
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
