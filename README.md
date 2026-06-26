# convunits

## Overview

`convunits` is a dependency-free Go CLI for dimensional unit conversion and a
small set of intentionally separate "weird" conversions. Normal conversions
use dimensional analysis: length converts to length, force converts to force,
and incompatible dimensions are rejected.

The tool supports:

- scalar, compound, and derived unit conversion
- unit catalog listing
- compare mode for expressing one quantity in several compatible units
- experimental unit-aware calculator expressions
- concise explanations for normal conversions and eval expressions
- a small force-relation solve mode
- nonlinear scale conversions
- approximate recipe ingredient mass/volume conversions
- approximate shoe-size foot-length mappings
- paper-size lookups
- wire gauges, drill sizes, and sieve openings
- named formula calculations
- JSON output for normal conversions, scale/solve mode, formulas, paper, wire,
  drill, and sieve commands

See [SUPPORTED_UNITS.md](SUPPORTED_UNITS.md) for a categorized reference.

## Install / build

From the repository root:

```sh
go test ./...
go build ./cmd/convunits
```

Run directly during development:

```sh
go run ./cmd/convunits 10kg lb
```

If you want a local executable on your shell path:

```sh
go install ./cmd/convunits
```

## Basic unit conversion

Values may be attached to the input unit or provided as a separate argument:

```sh
convunits 10kg lb
convunits 10 kg lb
convunits 500ms s
convunits 60mph km/h
```

Output precision defaults to ten significant digits:

```sh
convunits --precision 4 10kg lb
convunits --scientific 1e9m km
```

List units with:

```sh
convunits units
convunits units length
convunits units force
```

Parsing is case-sensitive. For example, `b` is a bit, `B` is a byte, `Mb` is a
megabit, and `MB` is a megabyte.

## Compare mode

Compare mode expresses one quantity in one or more compatible units. It uses
the normal parser/converter and remains dimensionally strict:

```sh
convunits compare 38in Rj
convunits compare 38in banana smoot cubit hand Rj Rsun
convunits compare 38 in banana smoot Rj
convunits compare 1earthcircumference marathon
convunits compare 1olympicpool cup gal
convunits compare 1Rsun Rj Re LD
convunits compare 60mph m/s km/h ft/s
```

Presets skip units that are incompatible with the input:

```sh
convunits compare 38in --fun
convunits compare 38in --human
convunits compare 38in --astronomical
convunits compare 38in --ancient
convunits compare 38in --all
```

Use `--no-limit` with `--all` to print every compatible unit. Approximate
input or target units are marked as approximate in the output.

## Unit-aware calculator

`eval` is an experimental, intentionally small unit-aware calculator. It is
useful for quick arithmetic, physical expressions, and silly ratios:

```sh
convunits eval '38in / Rj'
convunits eval '1Rsun / 38in'
convunits eval '1olympicpool / 1cup'
convunits eval '2 * pi * 1Re -> km'
convunits eval '0.5 * 1500kg * (60mph)^2 -> kWh'
convunits eval '1kg * 9.80665m/s^2 -> N'
```

Supported features are numbers, unit-attached numbers, `+`, `-`, `*`, `/`,
`^`, parentheses, unary minus, and constants `pi`, `c`, `G`, and `g0`.
Addition and subtraction require matching dimensions. Powers require
dimensionless exponents. The optional `->` arrow converts the result to a
compatible output unit.

`eval` is not a programming language: there are no variables, user-defined
functions, loops, conditionals, assignment, or scripting. Recipe ingredient
conversions remain under `convunits recipe`.

## Explain conversions

`explain` shows how a normal conversion or eval expression with `->` is
derived. It is for user-facing debugging, not symbolic algebra:

```sh
convunits explain 60mph m/s
convunits explain 10kg lb
convunits explain 1N 'kg*m/s^2'
convunits explain 38in Rj
convunits explain '2*pi*1Re -> km'
convunits explain '0.5 * 1500kg * (60mph)^2 -> kWh'
convunits --json explain 60mph m/s
```

Explain currently supports normal conversions and eval expressions with an
output arrow. Recipe, scale, shoe, paper, wire, drill, sieve, formula, and
compare explanations are not implemented yet.

## Compound and derived units

Unit expressions support multiplication, division, signed integer powers, and
parentheses. Quote expressions containing `*` so the shell does not expand
them:

```sh
convunits 1N 'kg*m/s^2'
convunits 1Pa 'N/m^2'
convunits 1W J/s
convunits 1 'g/cm^3' 'kg/m^3'
convunits 1 MB/s Mb/s
```

Astronomical and historical units are normal units when they have a linear
dimension:

```sh
convunits 1Rsun km
convunits 1cubit in
convunits 1smoot ft
```

Fuel economy and fuel consumption are reciprocal conversions. `L/100km` is
handled explicitly:

```sh
convunits 30mpg L/100km
convunits 7.5L/100km mpg
```

## Approximate and weird units

Some supported values are based on conventions, tables, or historical
approximations. `convunits` keeps those out of the core dimensional registry
when they are not honest scalar units.

Examples:

```sh
convunits 1cubit in
convunits 1smoot ft
convunits 1banana cm
convunits 1olympicpool gal
```

Approximate catalog entries are labeled in listings:

```sh
convunits units ancient-length
convunits units human-scale
convunits units astronomical-length
```

## Nonlinear scales

Logarithmic, ordinal, and lookup scales use the `scale` command. They are not
normal units:

```sh
convunits scale 7 pH mol/L
convunits scale 60 dB power-ratio
convunits scale 5 beaufort mph
convunits scale 5 mag brightness-ratio
convunits scale 12 awg diameter-mm
```

List scale families:

```sh
convunits scales
convunits scales ratio
```

Notes:

- pH is a hydrogen-ion concentration scale, not generic concentration
  conversion.
- dB/bel conversions require a target ratio type such as `power-ratio`.
- Beaufort output is range/lookup based.
- AWG exists in `scale` for numeric gauge math and in `wire` for user-friendly
  gauge spellings.

## Recipe conversions

Recipe mode converts between mass and volume for known cooking ingredients
using approximate bulk density data. Ingredient densities are not part of the
normal unit registry.

```sh
convunits recipe 1cup flour g
convunits recipe 2tbsp butter g
convunits recipe 100g sugar cup
convunits recipe 1cup honey oz
convunits recipe 500ml water lb
convunits recipe 1cup rice g
convunits recipe ingredients
convunits recipe ingredients baking
```

Mass-to-mass and volume-to-volume recipe conversions use normal unit
conversion and do not need density:

```sh
convunits recipe 100g flour oz
convunits recipe 1cup flour tbsp
```

Cooking conversions are approximate. They can vary by packing, brand,
humidity, grind, ingredient form, and measurement method.

## Shoe sizes

Shoe commands estimate foot length. They are not fit recommendations:

```sh
convunits shoe us-men 10 yd
convunits shoe us-women 8.5 cm
convunits shoe uk-adult 9 cm
convunits shoe eu 43 cm
convunits shoe mondo 270 cm
convunits shoe jp 27 in
```

List accepted systems:

```sh
convunits shoe systems
```

Supported systems are `us-men`, `us-women`, `uk-adult`, `eu`, `mondo`, and
`jp`. Ambiguous `us` and children's systems are rejected.

## Paper sizes

Paper sizes are two-dimensional lookups. Width and height are converted through
the normal length engine:

```sh
convunits paper a4 mm
convunits paper letter in
convunits papers
convunits papers iso
convunits papers us
convunits papers photo
```

`size` and `scale-size` remain aliases for `paper`.

## Wire gauges

Wire gauge reports approximate AWG conductor diameter:

```sh
convunits wire 12awg mm
convunits wire 0000awg in
convunits wires
```

Accepted gauge forms include `12`, `12awg`, `awg12`, `0000awg`, `000awg`,
`00awg`, and `0awg`.

## Drill bits

Drill lookup supports fractional, number, letter, and direct metric sizes:

```sh
convunits drill '#7' mm
convunits drill '1/4' mm
convunits drill A in
convunits drill 6.8mm in
convunits drills
convunits drills number
```

Number drills run `#80` through `#1`; letter drills run A through Z.

## Sieve openings

Sieve lookup reports approximate nominal ASTM E11/US openings:

```sh
convunits sieve 'No. 200' um
convunits sieve '#40' mm
convunits sieve 4mesh mm
convunits sieves
```

## Formula mode

Formula mode parses every named input with the normal unit parser, validates
dimensions, computes the formula, and validates the output unit.

Astronomy and physics examples:

```sh
convunits formula escape-velocity --mass 1Mearth --radius 1Re km/s
convunits formula schwarzschild-radius --mass 1Msun km
convunits formula gravity-force --mass1 1Mearth --mass2 1kg --distance 1Re N
convunits formula surface-gravity --mass 1Mearth --radius 1Re g0
convunits formula kinetic-energy --mass 1500kg --speed 60mph kWh
```

Geometry and everyday examples:

```sh
convunits formula bmi --mass 180lb --height 6ft bmi
convunits formula density --mass 1kg --volume 1L kg/m^3
convunits formula sphere-volume --radius 10cm L
convunits formula circle-area --radius 1ft in^2
convunits formula cylinder-volume --radius 10cm --height 1m L
convunits formula pressure --force 1lbf --area 1in^2 psi
convunits formula flow-rate --volume 1gal --time 1min gpm
convunits formula pace --distance 1mi --time 8min min/mi
convunits formula speed --distance 5km --time 25min mph
```

List formulas:

```sh
convunits formulas
```

BMI is only calculated. It is not interpreted medically.

## JSON output

Normal conversions support global `--json`:

```sh
convunits --json 10kg lb
```

Formula and weird commands also support JSON:

```sh
convunits --json formula bmi --mass 180lb --height 6ft bmi
convunits --json compare 38in banana smoot Rj
convunits --json eval '38in / Rj'
convunits --json explain 60mph m/s
convunits --json recipe 1cup flour g
convunits paper --json a4 mm
convunits --json wire 12awg mm
convunits --json drill '#7' mm
convunits --json sieve 'No. 200' um
```

Text output is unchanged unless JSON is requested.

## Design notes and limitations

### Philosophy / math honesty

- Normal conversion is dimensionally strict.
- Weird conversions live in separate commands.
- Approximate conversions say `approximately` or carry an approximation note.
- Nonlinear, ordinal, and lookup conversions are not treated as scalar units.

### Unsupported by design

- Currency is not supported.
- Redshift is not treated as a distance.
- pH is not generic concentration conversion.
- Beaufort is range/lookup based.
- Mohs hardness is not implemented as a conversion.
- Shoe sizes are approximate foot-length mappings, not fit recommendations.
- Recipe conversions are approximate cooking estimates, not scientific density
  definitions.
- Ancient units are approximate conventions.
- BMI is only calculated, not interpreted medically.

Temperature note: Celsius and Fahrenheit are affine units. They convert
directly, but cannot appear in compound expressions. In compound expressions,
`C` and `F` mean coulomb and farad.
