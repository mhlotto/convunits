package sieve

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var meshMM = map[int]float64{
	4: 4.75, 8: 2.36, 10: 2.00, 16: 1.18, 20: 0.850, 30: 0.600, 40: 0.425,
	50: 0.300, 60: 0.250, 80: 0.180, 100: 0.150, 140: 0.106, 170: 0.090,
	200: 0.075, 270: 0.053, 325: 0.045, 400: 0.038,
}

var inchMM = map[string]float64{
	"3in": 75, "2in": 50, "1.5in": 37.5, "1in": 25, "3/4in": 19,
	"1/2in": 12.5, "3/8in": 9.5,
}

func OpeningMeters(text string) (float64, error) {
	s := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(text), " ", ""))
	if mm, ok := inchMM[s]; ok {
		return mm / 1000, nil
	}
	var numberText string
	switch {
	case strings.HasPrefix(s, "#"):
		numberText = strings.TrimPrefix(s, "#")
	case strings.HasPrefix(s, "no."):
		numberText = strings.TrimPrefix(s, "no.")
	case strings.HasSuffix(s, "mesh"):
		numberText = strings.TrimSuffix(s, "mesh")
	case strings.HasPrefix(s, "mesh"):
		numberText = strings.TrimPrefix(s, "mesh")
	default:
		return 0, fmt.Errorf("unknown sieve size %q", text)
	}
	number, err := strconv.Atoi(numberText)
	if err != nil {
		return 0, fmt.Errorf("invalid sieve size %q", text)
	}
	mm, ok := meshMM[number]
	if !ok {
		return 0, fmt.Errorf("unsupported sieve mesh #%d", number)
	}
	return mm / 1000, nil
}

type Entry struct {
	Size      string
	OpeningMM float64
}

func Entries() []Entry {
	var out []Entry
	for size, mm := range inchMM {
		out = append(out, Entry{size, mm})
	}
	var meshes []int
	for mesh := range meshMM {
		meshes = append(meshes, mesh)
	}
	sort.Ints(meshes)
	for _, mesh := range meshes {
		out = append(out, Entry{fmt.Sprintf("#%d", mesh), meshMM[mesh]})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].OpeningMM > out[j].OpeningMM })
	return out
}
