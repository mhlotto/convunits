# convunits

## Overview

`convunits` is a dependency-free Go CLI for dimensional unit conversion plus a
set of deliberately separate commands for lookups, recipes, formulas, scales,
comparison, evaluation, and explanations.

Normal conversions are dimensionally strict: length converts to length, force
converts to force, and incompatible dimensions are rejected.

## Build

```sh
go test ./...
go build ./cmd/convunits
```

Run during development:

```sh
go run ./cmd/convunits 10kg lb
```

Install locally:

```sh
go install ./cmd/convunits
```

## Command map

```sh
convunits 10kg lb
convunits 60mph km/h
convunits compare 38in banana Rj
convunits eval '1kg*c^2 -> J'
convunits explain 60mph m/s
convunits scale 7 pH mol/L
convunits recipe 1cup flour g
convunits formula escape-velocity --mass 1Mearth --radius 1Re km/s
```

Use `--help` with `compare`, `recipe`, `eval`, `explain`, `formula`, and
`scale` for command-specific help. Lookup commands also accept concise help,
for example `convunits wire --help`.

## Basic conversions

```sh
convunits 10kg lb
convunits 10 kg lb
convunits 60mph km/h
convunits 100C F
convunits --precision 4 10kg lb
convunits --scientific 1e9m km
convunits units
convunits units length
```

Parsing is case-sensitive: `b` is bit, `B` is byte, `Mb` is megabit, and `MB`
is megabyte.

## Compound and derived units

Unit expressions support multiplication, division, integer powers, and
parentheses. Quote expressions containing `*` in shells:

```sh
convunits 1N 'kg*m/s^2'
convunits 1Pa 'N/m^2'
convunits 1W J/s
convunits 1 'g/cm^3' 'kg/m^3'
convunits 1 MB/s Mb/s
convunits 30mpg L/100km
```

## Weird but dimensional units

Some unusual units are still dimensional and live in the normal registry:

```sh
convunits 1Rsun km
convunits 1cubit in
convunits 1smoot ft
convunits 1banana cm
convunits 1olympicpool gal
```

Approximate entries are marked in listings:

```sh
convunits units astronomical-length
convunits units ancient-length
convunits units human-scale
```

## Compare mode

Compare expresses one quantity in one or more compatible units. It uses the
normal converter and remains dimensionally strict:

```sh
convunits compare 38in banana smoot Rj
convunits compare 1olympicpool cup gal
convunits compare 38in --astronomical
convunits compare 60mph m/s km/h ft/s
```

Presets include `--fun`, `--human`, `--astronomical`, `--ancient`, and `--all`.

## Unit-aware eval

`eval` is a small unit-aware calculator, not a programming language:

```sh
convunits eval '38in / Rj'
convunits eval '38in -> Rj'
convunits eval '1kg*c^2 -> J'
convunits eval '0.5 * 1500kg * (60mph)^2 -> kWh'
```

It supports numbers, unit-attached numbers, `+`, `-`, `*`, `/`, `^`,
parentheses, unary minus, and constants `pi`, `c`, `G`, and `g0`.

## Explain mode

Explain shows how a normal conversion or eval expression with `->` is derived:

```sh
convunits explain 60mph m/s
convunits explain 38in Rj
convunits explain '2*pi*1Re -> km'
```

It is explanatory, not a symbolic proof engine. Recipe, scale, shoe, paper,
wire, drill, sieve, formula, and compare explanations are not implemented yet.

## Nonlinear scales

Nonlinear, logarithmic, ordinal, and lookup scales use `scale`:

```sh
convunits scale 7 pH mol/L
convunits scale 60 dB power-ratio
convunits scale 5 beaufort mph
convunits scale 5 mag brightness-ratio
convunits scale 12 awg diameter-mm
convunits scales
```

## Recipe conversions

Recipe mode converts between mass and volume for known ingredients using
approximate bulk density data. Densities are ingredient-specific and are not
normal units.

```sh
convunits recipe 1cup flour g
convunits recipe 100g flour cup
convunits recipe 2tbsp butter g
convunits recipe 1cup honey oz
convunits recipe ingredients baking
```

Cooking conversions vary by packing, brand, humidity, grind, ingredient form,
and measurement method.

## Shoe sizes

Shoe commands estimate approximate foot length, not fit:

```sh
convunits shoe us-men 10 yd
convunits shoe eu 43 cm
convunits shoe systems
```

## Paper sizes

Paper sizes are two-dimensional lookups:

```sh
convunits paper a4 mm
convunits paper letter in
convunits size a4 mm
convunits papers iso
```

`size` and `scale-size` are aliases for `paper`.

## Wire gauges

```sh
convunits wire 12awg mm
convunits wire 0000awg in
convunits wires
```

## Drill bits

```sh
convunits drill '#7' mm
convunits drill '1/4' mm
convunits drill A in
convunits drills number
```

## Sieve openings

```sh
convunits sieve 'No. 200' um
convunits sieve '#40' mm
convunits sieve 4mesh mm
convunits sieves
```

## Formula mode

Formula mode parses and dimension-checks named inputs before computing:

```sh
convunits formula escape-velocity --mass 1Mearth --radius 1Re km/s
convunits formula schwarzschild-radius --mass 1Msun km
convunits formula bmi --mass 180lb --height 6ft bmi
convunits formula density --mass 1kg --volume 1L kg/m^3
convunits formulas
```

BMI is calculated only; no medical interpretation is provided.

## JSON output

`--json` works for normal conversion, scale, shoe, paper, wire, drill, sieve,
formula, compare, recipe, eval, and explain.

Examples:

```sh
convunits --json 10kg lb
convunits --json compare 38in banana smoot Rj
convunits --json explain 60mph m/s
```

Text output is unchanged unless JSON is requested.

## Design boundaries

- Normal conversions are dimensionally strict.
- `compare` is dimensionally strict but presentation-oriented.
- `eval` is a unit-aware calculator, not a programming language.
- `explain` is explanatory, not a symbolic proof engine.
- Celsius/Fahrenheit are affine and should use normal conversion mode.
- pH, dB, Beaufort, and magnitude are scale conversions.
- Recipe conversions are approximate and ingredient-specific.
- Shoe sizes are approximate foot-length mappings, not fit recommendations.
- Ancient and human-scale units use documented approximations.
- Currency is unsupported.
- Redshift is not treated as distance.
- Medical interpretation of BMI is not provided.

## Unsupported by design

- Currency and exchange rates.
- Redshift-as-distance.
- Mohs hardness conversion.
- Generic pH-as-concentration conversion outside the supported pH scale model.
- Recipe densities as normal units.
- Full programming-language behavior in `eval`.
