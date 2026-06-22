package drill

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var numberInches = map[int]float64{
	80: .0135, 79: .0145, 78: .0160, 77: .0180, 76: .0200, 75: .0210, 74: .0225, 73: .0240, 72: .0250, 71: .0260,
	70: .0280, 69: .0292, 68: .0310, 67: .0320, 66: .0330, 65: .0350, 64: .0360, 63: .0370, 62: .0380, 61: .0390,
	60: .0400, 59: .0410, 58: .0420, 57: .0430, 56: .0465, 55: .0520, 54: .0550, 53: .0595, 52: .0635, 51: .0670,
	50: .0700, 49: .0730, 48: .0760, 47: .0785, 46: .0810, 45: .0820, 44: .0860, 43: .0890, 42: .0935, 41: .0960,
	40: .0980, 39: .0995, 38: .1015, 37: .1040, 36: .1065, 35: .1100, 34: .1110, 33: .1130, 32: .1160, 31: .1200,
	30: .1285, 29: .1360, 28: .1405, 27: .1440, 26: .1470, 25: .1495, 24: .1520, 23: .1540, 22: .1570, 21: .1590,
	20: .1610, 19: .1660, 18: .1695, 17: .1730, 16: .1770, 15: .1800, 14: .1820, 13: .1850, 12: .1890, 11: .1910,
	10: .1935, 9: .1960, 8: .1990, 7: .2010, 6: .2040, 5: .2055, 4: .2090, 3: .2130, 2: .2210, 1: .2280,
}

var letterInches = map[string]float64{
	"A": .2340, "B": .2380, "C": .2420, "D": .2460, "E": .2500, "F": .2570, "G": .2610, "H": .2660, "I": .2720,
	"J": .2770, "K": .2810, "L": .2900, "M": .2950, "N": .3020, "O": .3160, "P": .3230, "Q": .3320, "R": .3390,
	"S": .3480, "T": .3580, "U": .3680, "V": .3770, "W": .3860, "X": .3970, "Y": .4040, "Z": .4130,
}

func DiameterMeters(text string) (float64, error) {
	s := strings.TrimSpace(text)
	if strings.HasSuffix(strings.ToLower(s), "mm") {
		value, err := strconv.ParseFloat(strings.TrimSpace(s[:len(s)-2]), 64)
		if err != nil || value <= 0 {
			return 0, fmt.Errorf("invalid metric drill size %q", text)
		}
		return value / 1000, nil
	}
	if strings.Contains(s, "/") {
		parts := strings.Split(s, "/")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid fractional drill size %q", text)
		}
		numerator, e1 := strconv.Atoi(parts[0])
		denominator, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil || numerator <= 0 || denominator <= 0 || 64%denominator != 0 || numerator*64/denominator > 64 {
			return 0, fmt.Errorf("fractional drill size %q must be from 1/64 through 1 inch in 1/64 increments", text)
		}
		return (float64(numerator) / float64(denominator)) * 0.0254, nil
	}
	if strings.HasPrefix(s, "#") {
		number, err := strconv.Atoi(strings.TrimPrefix(s, "#"))
		inches, ok := numberInches[number]
		if err != nil || !ok {
			return 0, fmt.Errorf("unknown number drill size %q", text)
		}
		return inches * 0.0254, nil
	}
	letter := strings.ToUpper(s)
	if len(letter) == 1 {
		if inches, ok := letterInches[letter]; ok {
			return inches * 0.0254, nil
		}
	}
	return 0, fmt.Errorf("unknown drill size %q; use a fraction, #number, letter, or metric value such as 6.8mm", text)
}

type Entry struct {
	Size, Category string
	DiameterIn     float64
}

func Entries(category string) ([]Entry, error) {
	if category != "" && category != "number" && category != "letter" && category != "fractional" {
		return nil, fmt.Errorf("unknown drill category %q", category)
	}
	var out []Entry
	if category == "" || category == "number" {
		var numbers []int
		for number := range numberInches {
			numbers = append(numbers, number)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(numbers)))
		for _, number := range numbers {
			out = append(out, Entry{fmt.Sprintf("#%d", number), "number", numberInches[number]})
		}
	}
	if category == "" || category == "letter" {
		for letter := 'A'; letter <= 'Z'; letter++ {
			s := string(letter)
			out = append(out, Entry{s, "letter", letterInches[s]})
		}
	}
	if category == "" || category == "fractional" {
		for n := 1; n <= 64; n++ {
			out = append(out, Entry{simplify64(n), "fractional", float64(n) / 64})
		}
	}
	return out, nil
}

func simplify64(n int) string {
	g := gcd(n, 64)
	return fmt.Sprintf("%d/%d", n/g, 64/g)
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
