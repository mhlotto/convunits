package units

import (
	"fmt"
	"sort"
	"strings"
)

type Registry struct {
	units   []*Unit
	aliases map[string][]*Unit
}

func NewRegistry() *Registry {
	r := &Registry{aliases: make(map[string][]*Unit)}
	r.registerCatalog()
	return r
}

func (r *Registry) Add(u Unit) {
	if u.Multiplier == 0 {
		u.Multiplier = 1
	}
	p := &u
	r.units = append(r.units, p)
	seen := make(map[string]bool)
	for _, a := range append([]string{u.Symbol, u.Name}, u.Aliases...) {
		if a != "" && !seen[a] {
			r.aliases[a] = append(r.aliases[a], p)
			seen[a] = true
		}
	}
}

func (r *Registry) LookupAll(name string) ([]*Unit, error) {
	u := r.aliases[name]
	if len(u) == 0 {
		return nil, UnknownUnitError{name}
	}
	return u, nil
}

func (r *Registry) lookupCompound(name string) (*Unit, error) {
	all, err := r.LookupAll(name)
	if err != nil {
		return nil, err
	}
	var usable []*Unit
	for _, u := range all {
		if !u.Affine {
			usable = append(usable, u)
		}
	}
	if len(usable) == 1 {
		return usable[0], nil
	}
	if len(usable) == 0 {
		return nil, fmt.Errorf("affine temperature unit %q cannot be used in a compound expression", name)
	}
	return nil, fmt.Errorf("ambiguous unit alias %q", name)
}

func (r *Registry) Units(category string) []*Unit {
	var out []*Unit
	for _, u := range r.units {
		if category == "" || strings.EqualFold(u.Category, category) {
			out = append(out, u)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Symbol < out[j].Symbol })
	return out
}

func d(l, m, t, a, temp, amount, lum, angle, info int) Dimension {
	return Dimension{l, m, t, a, temp, amount, lum, angle, info}
}

func (r *Registry) add(symbol, name, category string, dim Dimension, mul float64, aliases ...string) {
	r.Add(Unit{Symbol: symbol, Name: name, Category: category, Dimension: dim, Multiplier: mul, Aliases: aliases})
}

func (r *Registry) registerCatalog() {
	L, M, Ti := d(1, 0, 0, 0, 0, 0, 0, 0, 0), d(0, 1, 0, 0, 0, 0, 0, 0, 0), d(0, 0, 1, 0, 0, 0, 0, 0, 0)
	I, Temp := d(0, 0, 0, 1, 0, 0, 0, 0, 0), d(0, 0, 0, 0, 1, 0, 0, 0, 0)
	force, energy := d(1, 1, -2, 0, 0, 0, 0, 0, 0), d(2, 1, -2, 0, 0, 0, 0, 0, 0)
	power, pressure := d(2, 1, -3, 0, 0, 0, 0, 0, 0), d(-1, 1, -2, 0, 0, 0, 0, 0, 0)
	for _, x := range []struct {
		s, n string
		v    float64
		a    []string
	}{
		{"m", "meter", 1, []string{"meters", "metre", "metres"}}, {"km", "kilometer", 1e3, []string{"kilometers", "kilometre", "kilometres"}},
		{"cm", "centimeter", 1e-2, []string{"centimeters"}}, {"mm", "millimeter", 1e-3, []string{"millimeters"}}, {"um", "micrometer", 1e-6, []string{"micrometers", "µm"}}, {"nm", "nanometer", 1e-9, []string{"nanometers"}},
		{"in", "inch", 0.0254, []string{"inches"}}, {"ft", "foot", 0.3048, []string{"feet"}}, {"yd", "yard", 0.9144, []string{"yards"}}, {"mi", "mile", 1609.344, []string{"miles"}},
		{"nmi", "nautical mile", 1852, []string{"nauticalmile", "nautical-mile"}}, {"au", "astronomical unit", 149597870700, []string{"astronomicalunit"}}, {"ly", "light-year", 9.4607304725808e15, []string{"lightyear"}}, {"pc", "parsec", 3.0856775814913673e16, nil},
	} {
		r.add(x.s, x.n, "length", L, x.v, x.a...)
	}
	for _, x := range []struct {
		s, n string
		v    float64
		a    []string
	}{
		{"kg", "kilogram", 1, []string{"kilograms"}}, {"g", "gram", 1e-3, []string{"grams"}}, {"mg", "milligram", 1e-6, []string{"milligrams"}}, {"ug", "microgram", 1e-9, []string{"micrograms", "µg"}}, {"t", "tonne", 1e3, []string{"metricton"}},
		{"oz", "ounce", 0.028349523125, []string{"ounces"}}, {"lb", "pound", 0.45359237, []string{"pounds"}}, {"st", "stone", 6.35029318, nil}, {"ton", "short ton", 907.18474, []string{"shortton"}}, {"slug", "slug", 14.5939029372, nil},
	} {
		r.add(x.s, x.n, "mass", M, x.v, x.a...)
	}
	for _, x := range []struct {
		s, n string
		v    float64
		a    []string
	}{
		{"ns", "nanosecond", 1e-9, nil}, {"us", "microsecond", 1e-6, []string{"µs"}}, {"ms", "millisecond", 1e-3, nil}, {"s", "second", 1, []string{"sec", "seconds"}}, {"min", "minute", 60, []string{"minutes"}}, {"h", "hour", 3600, []string{"hr", "hours"}}, {"d", "day", 86400, []string{"days"}}, {"wk", "week", 604800, []string{"weeks"}}, {"mo", "month", 2629746, []string{"months"}}, {"y", "year", 31557600, []string{"yr", "years", "julianyear"}},
	} {
		r.add(x.s, x.n, "time", Ti, x.v, x.a...)
	}
	// Area, volume, and convenient compound aliases.
	r.add("acre", "acre", "area", L.Mul(2), 4046.8564224)
	r.add("ha", "hectare", "area", L.Mul(2), 10000, "hectares")
	r.add("L", "liter", "volume", L.Mul(3), 1e-3, "l", "liters", "litre", "litres")
	r.add("mL", "milliliter", "volume", L.Mul(3), 1e-6, "ml", "milliliters")
	r.add("gal", "US gallon", "volume", L.Mul(3), 0.003785411784, "gallon", "gallons", "USgallon")
	r.add("qt", "quart", "volume", L.Mul(3), 0.000946352946, "quarts")
	r.add("pt", "pint", "volume", L.Mul(3), 0.000473176473, "pints")
	r.add("cup", "cup", "volume", L.Mul(3), 0.0002365882365, "cups")
	r.add("floz", "fluid ounce", "volume", L.Mul(3), 0.0000295735295625, "fluidounce")
	r.add("tbsp", "tablespoon", "volume", L.Mul(3), 0.00001478676478125)
	r.add("tsp", "teaspoon", "volume", L.Mul(3), 0.00000492892159375)
	r.add("teaspoon-metric", "metric teaspoon", "volume", L.Mul(3), 5e-6, "metricteaspoon")
	r.add("mph", "mile per hour", "speed", L.Sub(Ti), 1609.344/3600)
	r.add("kph", "kilometer per hour", "speed", L.Sub(Ti), 1000.0/3600, "kmh")
	r.add("knot", "knot", "speed", L.Sub(Ti), 1852.0/3600, "kt")
	r.add("fps", "foot per second", "speed", L.Sub(Ti), 0.3048)
	r.add("g0", "standard gravity", "acceleration", L.Sub(Ti.Mul(2)), 9.80665, "gee")
	for _, x := range []struct {
		s, n string
		v    float64
		a    []string
	}{
		{"N", "newton", 1, nil}, {"kN", "kilonewton", 1e3, nil}, {"lbf", "pound-force", 4.4482216152605, []string{"poundforce"}}, {"dyn", "dyne", 1e-5, nil},
	} {
		r.add(x.s, x.n, "force", force, x.v, x.a...)
	}
	for _, x := range []struct {
		s, n string
		v    float64
		a    []string
	}{
		{"J", "joule", 1, nil}, {"kJ", "kilojoule", 1e3, nil}, {"cal", "calorie", 4.184, nil}, {"kcal", "kilocalorie", 4184, nil}, {"Wh", "watt-hour", 3600, nil}, {"kWh", "kilowatt-hour", 3.6e6, nil}, {"eV", "electronvolt", 1.602176634e-19, nil}, {"BTU", "British thermal unit", 1055.05585262, []string{"btu"}}, {"therm", "therm", 105505585.262, nil},
	} {
		r.add(x.s, x.n, "energy", energy, x.v, x.a...)
	}
	for _, x := range []struct {
		s, n string
		v    float64
	}{{"W", "watt", 1}, {"kW", "kilowatt", 1e3}, {"MW", "megawatt", 1e6}, {"hp", "horsepower", 745.6998715822702}} {
		r.add(x.s, x.n, "power", power, x.v)
	}
	for _, x := range []struct {
		s, n string
		v    float64
	}{{"Pa", "pascal", 1}, {"kPa", "kilopascal", 1e3}, {"MPa", "megapascal", 1e6}, {"bar", "bar", 1e5}, {"mbar", "millibar", 100}, {"atm", "atmosphere", 101325}, {"psi", "pounds per square inch", 6894.757293168}, {"torr", "torr", 101325.0 / 760}, {"mmHg", "millimeter of mercury", 133.322387415}} {
		r.add(x.s, x.n, "pressure", pressure, x.v)
	}
	for _, x := range []struct {
		s, n string
		v    float64
	}{{"Hz", "hertz", 1}, {"kHz", "kilohertz", 1e3}, {"MHz", "megahertz", 1e6}, {"GHz", "gigahertz", 1e9}, {"rpm", "revolutions per minute", 1.0 / 60}} {
		r.add(x.s, x.n, "frequency", Ti.Mul(-1), x.v)
	}
	angle := d(0, 0, 0, 0, 0, 0, 0, 1, 0)
	r.add("rad", "radian", "angle", angle, 1, "radians")
	r.add("deg", "degree", "angle", angle, 3.141592653589793/180, "degrees")
	r.add("grad", "gradian", "angle", angle, 3.141592653589793/200)
	r.add("turn", "turn", "angle", angle, 2*3.141592653589793, "rev")
	info := d(0, 0, 0, 0, 0, 0, 0, 0, 1)
	for _, x := range []struct {
		s, n string
		v    float64
	}{{"b", "bit", 0.125}, {"B", "byte", 1}, {"kb", "kilobit", 125}, {"kB", "kilobyte", 1e3}, {"Mb", "megabit", 125e3}, {"MB", "megabyte", 1e6}, {"Gb", "gigabit", 125e6}, {"GB", "gigabyte", 1e9}, {"Tb", "terabit", 125e9}, {"TB", "terabyte", 1e12}, {"KiB", "kibibyte", 1024}, {"MiB", "mebibyte", 1048576}, {"GiB", "gibibyte", 1073741824}, {"TiB", "tebibyte", 1099511627776}} {
		r.add(x.s, x.n, "information", info, x.v)
	}
	r.add("kbps", "kilobits per second", "data rate", info.Sub(Ti), 125, "Kbps")
	r.add("Mbps", "megabits per second", "data rate", info.Sub(Ti), 125e3)
	r.add("Gbps", "gigabits per second", "data rate", info.Sub(Ti), 125e6)
	r.add("gpm", "gallons per minute", "flow", L.Mul(3).Sub(Ti), 0.003785411784/60)
	r.add("mpg", "miles per gallon", "fuel economy", L.Mul(-2), 1609.344/0.003785411784)
	r.add("L/100km", "liters per 100 kilometers", "fuel consumption", L.Mul(2), litersPer100KilometersScale, "l/100km")
	// Temperature. C and F intentionally share symbols with coulomb/farad; conversion resolves by dimension.
	r.Add(Unit{Symbol: "K", Name: "kelvin", Category: "temperature", Dimension: Temp, Multiplier: 1, Aliases: []string{"kelvins"}})
	r.Add(Unit{Symbol: "C", Name: "Celsius", Category: "temperature", Dimension: Temp, Multiplier: 1, Offset: 273.15, Affine: true, Aliases: []string{"celsius", "°C"}})
	r.Add(Unit{Symbol: "F", Name: "Fahrenheit", Category: "temperature", Dimension: Temp, Multiplier: 5.0 / 9, Offset: 273.15 - 32*5.0/9, Affine: true, Aliases: []string{"fahrenheit", "°F"}})
	r.Add(Unit{Symbol: "R", Name: "Rankine", Category: "temperature", Dimension: Temp, Multiplier: 5.0 / 9, Aliases: []string{"rankine", "°R"}})
	r.add("A", "ampere", "electricity", I, 1, "amp", "amperes")
	r.add("mA", "milliampere", "electricity", I, 1e-3)
	volt := d(2, 1, -3, -1, 0, 0, 0, 0, 0)
	ohm := d(2, 1, -3, -2, 0, 0, 0, 0, 0)
	r.add("V", "volt", "electricity", volt, 1, "volts")
	r.add("Ohm", "ohm", "electricity", ohm, 1, "ohms", "Ω")
	r.add("C", "coulomb", "electricity", I.Add(Ti), 1, "coulombs")
	r.add("F", "farad", "electricity", d(-2, -1, 4, 2, 0, 0, 0, 0, 0), 1, "farads")
	r.add("H", "henry", "electricity", d(2, 1, -2, -2, 0, 0, 0, 0, 0), 1)
	r.add("Wb", "weber", "electricity", d(2, 1, -2, -1, 0, 0, 0, 0, 0), 1)
	r.add("T", "tesla", "electricity", d(0, 1, -2, -1, 0, 0, 0, 0, 0), 1)
	r.add("mol", "mole", "substance", d(0, 0, 0, 0, 0, 1, 0, 0, 0), 1, "moles")
	r.add("mmol", "millimole", "substance", d(0, 0, 0, 0, 0, 1, 0, 0, 0), 1e-3)
	r.add("cd", "candela", "luminous", d(0, 0, 0, 0, 0, 0, 1, 0, 0), 1)
	r.add("lm", "lumen", "luminous", d(0, 0, 0, 0, 0, 0, 1, 0, 0), 1)
	r.add("lx", "lux", "luminous", d(-2, 0, 0, 0, 0, 0, 1, 0, 0), 1)
	for _, u := range r.units {
		if u.Symbol == "mo" || u.Symbol == "y" {
			u.Approximate = true
		}
	}
	// Generate unambiguous SI-prefixed forms for common metric base units.
	for _, symbol := range []string{"m", "g", "s", "A", "V", "Ohm", "Hz", "J", "W", "Pa", "mol"} {
		r.addSIPrefixes(symbol)
	}
	r.registerExpandedCatalog()
}

func (r *Registry) addSIPrefixes(base string) {
	all := r.aliases[base]
	if len(all) != 1 {
		return
	}
	prefixes := []struct {
		symbol, name string
		multiplier   float64
	}{
		{"p", "pico", 1e-12}, {"n", "nano", 1e-9}, {"u", "micro", 1e-6},
		{"m", "milli", 1e-3}, {"c", "centi", 1e-2}, {"d", "deci", 1e-1},
		{"da", "deca", 1e1}, {"h", "hecto", 1e2}, {"k", "kilo", 1e3},
		{"M", "mega", 1e6}, {"G", "giga", 1e9}, {"T", "tera", 1e12},
	}
	u := all[0]
	for _, p := range prefixes {
		symbol := p.symbol + base
		if len(r.aliases[symbol]) != 0 {
			continue
		}
		r.add(symbol, p.name+u.Name, u.Category, u.Dimension, p.multiplier*u.Multiplier)
	}
}
