package units

import "fmt"

const (
	speedOfLight = 299792458.0
	julianYear   = 31557600.0
)

// addUnique is used by the expanded catalog so new definitions cannot
// accidentally introduce aliases that change existing parsing behavior.
func (r *Registry) addUnique(u Unit) {
	for _, alias := range append([]string{u.Symbol, u.Name}, u.Aliases...) {
		if alias == "" {
			continue
		}
		if existing := r.aliases[alias]; len(existing) != 0 {
			panic(fmt.Sprintf("expanded unit alias %q conflicts with %s", alias, existing[0].Name))
		}
	}
	r.Add(u)
}

func (r *Registry) registerExpandedCatalog() {
	length := Dimension{Length: 1}
	mass := Dimension{Mass: 1}
	time := Dimension{Time: 1}
	area := Dimension{Length: 2}
	volume := Dimension{Length: 3}
	speed := Dimension{Length: 1, Time: -1}
	energy := Dimension{Mass: 1, Length: 2, Time: -2}
	pressure := Dimension{Mass: 1, Length: -1, Time: -2}
	frequency := Dimension{Time: -1}
	information := Dimension{Information: 1}
	dose := Dimension{Length: 2, Time: -2}
	concentration := Dimension{Amount: 1, Length: -3}
	massConcentration := Dimension{Mass: 1, Length: -3}

	r.registerAstronomicalUnits(length, mass, time, speed)
	r.registerHistoricalUnits(length, area)
	r.registerGeologicUnits(time, pressure)
	r.registerAtomicUnits(length, mass, energy)
	r.registerRadiationUnits(dose, frequency)
	r.registerChemistryUnits(concentration, massConcentration)
	r.registerTypographyUnits(length)
	r.registerComputingUnits(information, frequency)
	r.registerSignalUnits(time, frequency)
	r.registerHumanScaleUnits(length, volume)
}

func (r *Registry) registerAstronomicalUnits(length, mass, time, speed Dimension) {
	for _, u := range []Unit{
		{Symbol: "ls", Name: "light-second", Category: "astronomical-length", Dimension: length, Multiplier: speedOfLight, Aliases: []string{"lightsecond"}},
		// lm remains lumen for backward compatibility.
		{Symbol: "lightmin", Name: "light-minute", Category: "astronomical-length", Dimension: length, Multiplier: speedOfLight * 60, Aliases: []string{"lightminute"}, Note: "symbol lm is reserved for lumen"},
		{Symbol: "lh", Name: "light-hour", Category: "astronomical-length", Dimension: length, Multiplier: speedOfLight * 3600, Aliases: []string{"lighthour"}},
		{Symbol: "ld", Name: "light-day", Category: "astronomical-length", Dimension: length, Multiplier: speedOfLight * 86400, Aliases: []string{"lightday"}},
		{Symbol: "Re", Name: "Earth mean radius", Category: "astronomical-length", Dimension: length, Multiplier: 6371008.8, Aliases: []string{"earthradius"}, Approximate: true, Note: "mean planetary radius"},
		{Symbol: "Rj", Name: "Jupiter mean radius", Category: "astronomical-length", Dimension: length, Multiplier: 69911000, Aliases: []string{"jupiterradius"}, Approximate: true, Note: "mean planetary radius"},
		{Symbol: "Rsun", Name: "solar radius", Category: "astronomical-length", Dimension: length, Multiplier: 695700000, Aliases: []string{"solarradius"}, Approximate: true, Note: "nominal solar radius"},
		{Symbol: "LD", Name: "average lunar distance", Category: "astronomical-length", Dimension: length, Multiplier: 384400000, Aliases: []string{"lunardistance"}, Approximate: true, Note: "average Earth-Moon distance"},
		{Symbol: "Mearth", Name: "Earth mass", Category: "astronomical-mass", Dimension: mass, Multiplier: 5.9722e24, Aliases: []string{"earthmass"}, Approximate: true, Note: "nominal astronomical value"},
		{Symbol: "Mjup", Name: "Jupiter mass", Category: "astronomical-mass", Dimension: mass, Multiplier: 1.89813e27, Aliases: []string{"jupitermass"}, Approximate: true, Note: "nominal astronomical value"},
		{Symbol: "Msun", Name: "solar mass", Category: "astronomical-mass", Dimension: mass, Multiplier: 1.98847e30, Aliases: []string{"solarmass"}, Approximate: true, Note: "nominal astronomical value"},
		{Symbol: "siderealday", Name: "sidereal day", Category: "astronomical-time", Dimension: time, Multiplier: 86164.0905},
		{Symbol: "siderealyear", Name: "sidereal year", Category: "astronomical-time", Dimension: time, Multiplier: 31558149.7635456},
		{Symbol: "galacticyear", Name: "galactic year", Category: "astronomical-time", Dimension: time, Multiplier: 225e6 * julianYear, Aliases: []string{"cosmicyear"}, Approximate: true, Note: "using 225 million Julian years"},
		{Symbol: "c0", Name: "speed of light", Category: "speed", Dimension: speed, Multiplier: speedOfLight, Aliases: []string{"lightspeed"}},
		{Symbol: "earthorbitalvelocity", Name: "Earth orbital velocity", Category: "speed", Dimension: speed, Multiplier: 29780, Approximate: true, Note: "mean orbital speed"},
		{Symbol: "escapeearth", Name: "Earth escape velocity", Category: "speed", Dimension: speed, Multiplier: 11186, Approximate: true, Note: "near-surface nominal value"},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerHistoricalUnits(length, area Dimension) {
	for _, u := range []Unit{
		{Symbol: "cubit", Name: "common cubit", Category: "ancient-length", Dimension: length, Multiplier: 0.4572, Aliases: []string{"cubits", "commoncubit"}, Approximate: true, Note: "varies by culture; using 18 inches"},
		{Symbol: "royalcubit", Name: "royal Egyptian cubit", Category: "ancient-length", Dimension: length, Multiplier: 0.5236, Aliases: []string{"egyptiancubit"}, Approximate: true, Note: "using 20.62 inches"},
		{Symbol: "span", Name: "span", Category: "ancient-length", Dimension: length, Multiplier: 0.2286, Approximate: true, Note: "varies historically; using 9 inches"},
		{Symbol: "handbreadth", Name: "handbreadth", Category: "ancient-length", Dimension: length, Multiplier: 0.0762, Aliases: []string{"palm"}, Approximate: true, Note: "varies historically; using 3 inches"},
		{Symbol: "fingerbreadth", Name: "fingerbreadth", Category: "ancient-length", Dimension: length, Multiplier: 0.01905, Aliases: []string{"digit"}, Approximate: true, Note: "varies historically; using 0.75 inches"},
		{Symbol: "pace", Name: "pace", Category: "ancient-length", Dimension: length, Multiplier: 0.762, Approximate: true, Note: "using 2.5 feet"},
		{Symbol: "romanfoot", Name: "Roman foot", Category: "ancient-length", Dimension: length, Multiplier: 0.296, Approximate: true, Note: "varies historically; using common modern approximation"},
		{Symbol: "romanmile", Name: "Roman mile", Category: "ancient-length", Dimension: length, Multiplier: 1479.5, Approximate: true, Note: "varies historically; using common modern approximation"},
		{Symbol: "stadion", Name: "stadion", Category: "ancient-length", Dimension: length, Multiplier: 185, Aliases: []string{"stadium", "stade"}, Approximate: true, Note: "varies historically; using 185 meters"},
		{Symbol: "parasang", Name: "parasang", Category: "ancient-length", Dimension: length, Multiplier: 5556, Approximate: true, Note: "varies historically; using three miles"},
		{Symbol: "league", Name: "league", Category: "historical-length", Dimension: length, Multiplier: 4828.032, Approximate: true, Note: "using three statute miles"},
		{Symbol: "fathom", Name: "fathom", Category: "nautical-length", Dimension: length, Multiplier: 1.8288, Approximate: true, Note: "traditional maritime unit; using six feet"},
		{Symbol: "cable", Name: "cable length", Category: "nautical-length", Dimension: length, Multiplier: 185.2, Approximate: true, Note: "using one tenth nautical mile"},
		{Symbol: "shot", Name: "anchor-chain shot", Category: "nautical-length", Dimension: length, Multiplier: 27.432, Approximate: true, Note: "traditional 15-fathom shot"},
		{Symbol: "nleague", Name: "nautical league", Category: "nautical-length", Dimension: length, Multiplier: 5556, Aliases: []string{"league-nautical"}, Approximate: true, Note: "using three nautical miles"},
		{Symbol: "chain", Name: "survey chain", Category: "historical-length", Dimension: length, Multiplier: 20.1168, Approximate: true, Note: "traditional survey unit"},
		{Symbol: "rod", Name: "survey rod", Category: "historical-length", Dimension: length, Multiplier: 5.0292, Approximate: true, Note: "traditional survey unit"},
		{Symbol: "furlong", Name: "furlong", Category: "historical-length", Dimension: length, Multiplier: 201.168, Approximate: true, Note: "traditional survey unit"},
		{Symbol: "hand", Name: "hand", Category: "historical-length", Dimension: length, Multiplier: 0.1016, Approximate: true, Note: "traditional horse-height unit; using four inches"},
		{Symbol: "rood", Name: "rood", Category: "area", Dimension: area, Multiplier: 1011.7141056},
		{Symbol: "section", Name: "survey section", Category: "area", Dimension: area, Multiplier: 2589988.110336},
		{Symbol: "township", Name: "survey township", Category: "area", Dimension: area, Multiplier: 93239571.972096},
		{Symbol: "barn", Name: "barn", Category: "physics-area", Dimension: area, Multiplier: 1e-28, Note: "nuclear cross-section unit"},
		{Symbol: "shed", Name: "shed", Category: "physics-area", Dimension: area, Multiplier: 1e-52, Note: "humorous rare cross-section unit"},
		{Symbol: "outhouse", Name: "outhouse", Category: "physics-area", Dimension: area, Multiplier: 1e-34, Note: "humorous rare cross-section unit"},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerGeologicUnits(time, pressure Dimension) {
	for _, u := range []Unit{
		{Symbol: "ka", Name: "kiloannum", Category: "geologic-time", Dimension: time, Multiplier: 1e3 * julianYear},
		{Symbol: "Ma", Name: "megaannum", Category: "geologic-time", Dimension: time, Multiplier: 1e6 * julianYear},
		{Symbol: "Ga", Name: "gigaannum", Category: "geologic-time", Dimension: time, Multiplier: 1e9 * julianYear},
		{Symbol: "eon", Name: "eon", Category: "geologic-time", Dimension: time, Multiplier: 1e9 * julianYear, Approximate: true, Note: "using one billion Julian years"},
		{Symbol: "Mbar", Name: "megabar", Category: "pressure", Dimension: pressure, Multiplier: 1e11},
		{Symbol: "kbar", Name: "kilobar", Category: "pressure", Dimension: pressure, Multiplier: 1e8},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerAtomicUnits(length, mass, energy Dimension) {
	for _, u := range []Unit{
		{Symbol: "u", Name: "unified atomic mass unit", Category: "atomic-mass", Dimension: mass, Multiplier: 1.66053906660e-27, Aliases: []string{"amu"}},
		{Symbol: "Da", Name: "dalton", Category: "atomic-mass", Dimension: mass, Multiplier: 1.66053906660e-27},
		{Symbol: "me", Name: "electron mass", Category: "atomic-mass", Dimension: mass, Multiplier: 9.1093837139e-31},
		{Symbol: "mp", Name: "proton mass", Category: "atomic-mass", Dimension: mass, Multiplier: 1.67262192595e-27},
		{Symbol: "mn", Name: "neutron mass", Category: "atomic-mass", Dimension: mass, Multiplier: 1.67492750056e-27},
		{Symbol: "angstrom", Name: "angstrom", Category: "atomic-length", Dimension: length, Multiplier: 1e-10, Aliases: []string{"Å", "Aang"}},
		{Symbol: "bohr", Name: "Bohr radius", Category: "atomic-length", Dimension: length, Multiplier: 5.29177210544e-11, Aliases: []string{"a0"}},
		{Symbol: "fermi", Name: "fermi", Category: "atomic-length", Dimension: length, Multiplier: 1e-15, Aliases: []string{"fm"}},
		{Symbol: "hartree", Name: "Hartree energy", Category: "atomic-energy", Dimension: energy, Multiplier: 4.3597447222060e-18, Aliases: []string{"Eh"}},
		{Symbol: "rydberg", Name: "Rydberg energy", Category: "atomic-energy", Dimension: energy, Multiplier: 2.1798723611030e-18, Aliases: []string{"Ry"}},
		{Symbol: "erg", Name: "erg", Category: "energy", Dimension: energy, Multiplier: 1e-7},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerRadiationUnits(dose, frequency Dimension) {
	for _, u := range []Unit{
		{Symbol: "Gy", Name: "gray", Category: "radiation", Dimension: dose, Multiplier: 1, Note: "absorbed dose (J/kg)"},
		{Symbol: "rad-dose", Name: "radiation absorbed dose", Category: "radiation", Dimension: dose, Multiplier: 0.01, Aliases: []string{"rd"}, Note: "0.01 Gy; rad is reserved for radians"},
		{Symbol: "Sv", Name: "sievert", Category: "radiation", Dimension: dose, Multiplier: 1, Note: "equivalent/effective dose; radiation weighting applies in real use"},
		{Symbol: "rem", Name: "roentgen equivalent man", Category: "radiation", Dimension: dose, Multiplier: 0.01, Note: "0.01 Sv equivalent dose"},
		{Symbol: "Bq", Name: "becquerel", Category: "radioactivity", Dimension: frequency, Multiplier: 1},
		{Symbol: "Ci", Name: "curie", Category: "radioactivity", Dimension: frequency, Multiplier: 3.7e10},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerChemistryUnits(concentration, massConcentration Dimension) {
	for _, u := range []Unit{
		{Symbol: "Molar", Name: "molar concentration", Category: "concentration", Dimension: concentration, Multiplier: 1000, Aliases: []string{"M"}},
		{Symbol: "mMolar", Name: "millimolar concentration", Category: "concentration", Dimension: concentration, Multiplier: 1, Aliases: []string{"mM"}},
		{Symbol: "uMolar", Name: "micromolar concentration", Category: "concentration", Dimension: concentration, Multiplier: 1e-3, Aliases: []string{"uM"}},
		{Symbol: "nMolar", Name: "nanomolar concentration", Category: "concentration", Dimension: concentration, Multiplier: 1e-6, Aliases: []string{"nM"}},
		{Symbol: "ppm-water", Name: "parts per million in dilute water", Category: "mass-concentration", Dimension: massConcentration, Multiplier: 1e-3, Approximate: true, Note: "approximates 1 mg/L for dilute water only"},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerTypographyUnits(length Dimension) {
	point := 0.00035277777777777776
	didot := 0.0003759715104
	for _, u := range []Unit{
		{Symbol: "point", Name: "typographic point", Category: "typography-length", Dimension: length, Multiplier: point, Aliases: []string{"typographicpoint"}, Note: "pt remains the pint symbol"},
		{Symbol: "pica", Name: "pica", Category: "typography-length", Dimension: length, Multiplier: 0.004233333333333333},
		{Symbol: "twip", Name: "twentieth of a point", Category: "typography-length", Dimension: length, Multiplier: point / 20},
		{Symbol: "csspx", Name: "CSS pixel", Category: "screen-length", Dimension: length, Multiplier: 0.0254 / 96},
		{Symbol: "didot", Name: "Didot point", Category: "typography-length", Dimension: length, Multiplier: didot, Approximate: true, Note: "historically variable; using 0.3759715104 mm"},
		{Symbol: "cicero", Name: "cicero", Category: "typography-length", Dimension: length, Multiplier: 12 * didot, Approximate: true, Note: "using 12 Didot points"},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerComputingUnits(information, frequency Dimension) {
	for _, u := range []Unit{
		{Symbol: "PB", Name: "petabyte", Category: "information", Dimension: information, Multiplier: 1e15},
		{Symbol: "EB", Name: "exabyte", Category: "information", Dimension: information, Multiplier: 1e18},
		{Symbol: "Pb", Name: "petabit", Category: "information", Dimension: information, Multiplier: 1e15 / 8},
		{Symbol: "Eb", Name: "exabit", Category: "information", Dimension: information, Multiplier: 1e18 / 8},
		{Symbol: "PiB", Name: "pebibyte", Category: "information", Dimension: information, Multiplier: 1 << 50},
		{Symbol: "EiB", Name: "exbibyte", Category: "information", Dimension: information, Multiplier: 1 << 60},
		{Symbol: "nibble", Name: "nibble", Category: "information", Dimension: information, Multiplier: 0.5},
		{Symbol: "word16", Name: "16-bit word", Category: "information", Dimension: information, Multiplier: 2},
		{Symbol: "word32", Name: "32-bit word", Category: "information", Dimension: information, Multiplier: 4},
		{Symbol: "word64", Name: "64-bit word", Category: "information", Dimension: information, Multiplier: 8},
		{Symbol: "block512", Name: "512-byte block", Category: "information", Dimension: information, Multiplier: 512},
		{Symbol: "page4k", Name: "4-KiB page", Category: "information", Dimension: information, Multiplier: 4096},
		{Symbol: "baud", Name: "baud", Category: "symbol-rate", Dimension: frequency, Multiplier: 1, Note: "symbols per second, not information per second"},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerSignalUnits(time, frequency Dimension) {
	for _, u := range []Unit{
		{Symbol: "bpm", Name: "beats per minute", Category: "frequency", Dimension: frequency, Multiplier: 1.0 / 60},
		{Symbol: "frame24", Name: "24 fps frame", Category: "media-time", Dimension: time, Multiplier: 1.0 / 24},
		{Symbol: "frame25", Name: "25 fps frame", Category: "media-time", Dimension: time, Multiplier: 1.0 / 25},
		{Symbol: "frame30", Name: "30 fps frame", Category: "media-time", Dimension: time, Multiplier: 1.0 / 30},
		{Symbol: "frame60", Name: "60 fps frame", Category: "media-time", Dimension: time, Multiplier: 1.0 / 60},
		{Symbol: "sample44100", Name: "44.1 kHz sample period", Category: "media-time", Dimension: time, Multiplier: 1.0 / 44100},
		{Symbol: "sample48000", Name: "48 kHz sample period", Category: "media-time", Dimension: time, Multiplier: 1.0 / 48000},
	} {
		r.addUnique(u)
	}
}

func (r *Registry) registerHumanScaleUnits(length, volume Dimension) {
	for _, u := range []Unit{
		{Symbol: "banana", Name: "banana length", Category: "human-scale", Dimension: length, Multiplier: 0.18, Approximate: true, Note: "bananas vary; using 18 centimeters"},
		{Symbol: "smoot", Name: "smoot", Category: "human-scale", Dimension: length, Multiplier: 1.7018, Approximate: true, Note: "nominal Oliver Smoot height"},
		{Symbol: "footballfield", Name: "US football field", Category: "human-scale", Dimension: length, Multiplier: 91.44, Aliases: []string{"footballfield-us"}, Approximate: true, Note: "100 yards, excluding end zones"},
		{Symbol: "marathon", Name: "marathon distance", Category: "human-scale", Dimension: length, Multiplier: 42195, Approximate: true, Note: "official race distance"},
		{Symbol: "earthcircumference", Name: "Earth equatorial circumference", Category: "human-scale", Dimension: length, Multiplier: 40075017, Approximate: true, Note: "equatorial circumference"},
		{Symbol: "olympicpool", Name: "Olympic pool volume", Category: "human-scale", Dimension: volume, Multiplier: 2500, Approximate: true, Note: "50 m x 25 m x 2 m assumed depth"},
	} {
		r.addUnique(u)
	}
}
