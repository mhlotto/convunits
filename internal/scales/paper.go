package scales

import (
	"fmt"
	"sort"
)

type PaperSize struct {
	Symbol, Name, Category string
	WidthM, HeightM        float64
	Approximate            bool
}

var paperSizes = buildPaperSizes()

func buildPaperSizes() map[string]PaperSize {
	out := make(map[string]PaperSize)
	addMM := func(symbol, name, category string, width, height float64) {
		out[symbol] = PaperSize{symbol, name, category, width / 1000, height / 1000, false}
	}
	for i, dims := range [][2]float64{{841, 1189}, {594, 841}, {420, 594}, {297, 420}, {210, 297}, {148, 210}, {105, 148}, {74, 105}, {52, 74}, {37, 52}, {26, 37}} {
		symbol := fmt.Sprintf("a%d", i)
		addMM(symbol, "ISO A"+fmt.Sprint(i), "iso", dims[0], dims[1])
	}
	for i, dims := range [][2]float64{{1000, 1414}, {707, 1000}, {500, 707}, {353, 500}, {250, 353}, {176, 250}, {125, 176}, {88, 125}, {62, 88}, {44, 62}, {31, 44}} {
		symbol := fmt.Sprintf("b%d", i)
		addMM(symbol, "ISO B"+fmt.Sprint(i), "iso", dims[0], dims[1])
	}
	for i, dims := range [][2]float64{{917, 1297}, {648, 917}, {458, 648}, {324, 458}, {229, 324}, {162, 229}, {114, 162}, {81, 114}, {57, 81}, {40, 57}, {28, 40}} {
		symbol := fmt.Sprintf("c%d", i)
		addMM(symbol, "ISO C"+fmt.Sprint(i), "iso", dims[0], dims[1])
	}
	addIn := func(symbol, name, category string, width, height float64) {
		out[symbol] = PaperSize{symbol, name, category, width * 0.0254, height * 0.0254, false}
	}
	addIn("letter", "US Letter", "us", 8.5, 11)
	addIn("legal", "US Legal", "us", 8.5, 14)
	addIn("tabloid", "US Tabloid", "us", 11, 17)
	addIn("ledger", "US Ledger", "us", 17, 11)
	addIn("executive", "US Executive", "us", 7.25, 10.5)
	addIn("index3x5", "3 x 5 index card", "us", 3, 5)
	addIn("index4x6", "4 x 6 index card", "us", 4, 6)
	addIn("index5x8", "5 x 8 index card", "us", 5, 8)
	addIn("photo4x6", "4 x 6 photo", "photo", 4, 6)
	addIn("photo5x7", "5 x 7 photo", "photo", 5, 7)
	addIn("photo8x10", "8 x 10 photo", "photo", 8, 10)
	return out
}

func IsPaperSize(name string) bool { _, ok := paperSizes[name]; return ok }

func LookupPaperSize(name string) (PaperSize, error) {
	size, ok := paperSizes[name]
	if !ok {
		return PaperSize{}, fmt.Errorf("unknown paper size %q", name)
	}
	return size, nil
}

func PaperSizes(category string) []PaperSize {
	var out []PaperSize
	for _, size := range paperSizes {
		if category == "" || size.Category == category {
			out = append(out, size)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })
	return out
}
