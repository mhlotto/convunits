package recipe

import (
	"fmt"
	"sort"
	"strings"
)

type Ingredient struct {
	Key            string
	Name           string
	Category       string
	Aliases        []string
	DensityKgPerM3 float64
	Note           string
}

var ingredients = []Ingredient{
	{Key: "water", Name: "water", Category: "liquids", DensityKgPerM3: 1000},
	{Key: "milk", Name: "milk", Category: "liquids", DensityKgPerM3: 1030},
	{Key: "oil", Name: "oil", Category: "liquids", DensityKgPerM3: 920},
	{Key: "olive-oil", Name: "olive oil", Category: "liquids", Aliases: []string{"evoo"}, DensityKgPerM3: 910},
	{Key: "honey", Name: "honey", Category: "liquids", DensityKgPerM3: 1420},
	{Key: "maple-syrup", Name: "maple syrup", Category: "liquids", DensityKgPerM3: 1320},
	{Key: "all-purpose-flour", Name: "all-purpose flour", Category: "baking", Aliases: []string{"flour"}, DensityKgPerM3: 508, Note: "approx 120 g per US cup"},
	{Key: "bread-flour", Name: "bread flour", Category: "baking", DensityKgPerM3: 550},
	{Key: "cake-flour", Name: "cake flour", Category: "baking", DensityKgPerM3: 480},
	{Key: "granulated-sugar", Name: "granulated sugar", Category: "baking", Aliases: []string{"sugar"}, DensityKgPerM3: 845, Note: "approx 200 g per US cup"},
	{Key: "brown-sugar", Name: "brown sugar", Category: "baking", DensityKgPerM3: 930, Note: "packed brown sugar approximation"},
	{Key: "powdered-sugar", Name: "powdered sugar", Category: "baking", DensityKgPerM3: 480},
	{Key: "cocoa-powder", Name: "cocoa powder", Category: "baking", DensityKgPerM3: 420},
	{Key: "baking-powder", Name: "baking powder", Category: "baking", DensityKgPerM3: 900},
	{Key: "baking-soda", Name: "baking soda", Category: "baking", DensityKgPerM3: 920},
	{Key: "salt", Name: "salt", Category: "baking", DensityKgPerM3: 1217},
	{Key: "kosher-salt", Name: "kosher salt", Category: "baking", DensityKgPerM3: 530, Note: "brand/grain size varies heavily"},
	{Key: "butter", Name: "butter", Category: "fats", DensityKgPerM3: 959, Note: "approx 227 g per US cup"},
	{Key: "peanut-butter", Name: "peanut butter", Category: "fats", DensityKgPerM3: 1080},
	{Key: "uncooked-rice", Name: "uncooked rice", Category: "grains", Aliases: []string{"rice"}, DensityKgPerM3: 790},
	{Key: "rolled-oats", Name: "rolled oats", Category: "grains", Aliases: []string{"oats"}, DensityKgPerM3: 360},
	{Key: "breadcrumbs", Name: "breadcrumbs", Category: "grains", DensityKgPerM3: 450},
	{Key: "cornmeal", Name: "cornmeal", Category: "grains", DensityKgPerM3: 720},
	{Key: "chopped-onion", Name: "chopped onion", Category: "produce", DensityKgPerM3: 530},
	{Key: "diced-tomato", Name: "diced tomato", Category: "produce", DensityKgPerM3: 650},
	{Key: "shredded-cheese", Name: "shredded cheese", Category: "produce", DensityKgPerM3: 450},
	{Key: "grated-parmesan", Name: "grated parmesan", Category: "produce", DensityKgPerM3: 1000},
}

func Lookup(name string) (Ingredient, error) {
	var matches []Ingredient
	for _, ingredient := range ingredients {
		if ingredient.Key == name {
			matches = append(matches, ingredient)
		}
		for _, alias := range ingredient.Aliases {
			if alias == name {
				matches = append(matches, ingredient)
			}
		}
	}
	if len(matches) == 0 {
		return Ingredient{}, fmt.Errorf("unknown ingredient %q", name)
	}
	if len(matches) > 1 {
		return Ingredient{}, fmt.Errorf("ambiguous ingredient alias %q", name)
	}
	return matches[0], nil
}

func Ingredients(category string) []Ingredient {
	var out []Ingredient
	for _, ingredient := range ingredients {
		if category == "" || strings.EqualFold(ingredient.Category, category) {
			out = append(out, ingredient)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}
