# convunits

`convunits` is a fast, dependency-free Go CLI for scalar, compound, derived,
temperature, angle, and information-unit conversions. It uses dimensional
analysis rather than pairwise conversion rules.

## Build and test

```sh
go test ./...
go build ./cmd/convunits
```

## Usage

```sh
convunits 10kg lb
convunits 10 kg lb
convunits 500ms s
convunits 60mph km/h
convunits 1N 'kg*m/s^2'
convunits 1J 'N*m'
convunits 100F C
```

Unit expressions support multiplication, division, signed integer powers, and
parentheses. Quote expressions containing `*` to prevent shell expansion.
Parsing is case-sensitive: `b` is a bit, `B` is a byte, `Mb` is a megabit,
and `MB` is a megabyte.

```sh
convunits 1Pa 'N/m^2'
convunits 1W J/s
convunits 1 'g/cm^3' 'kg/m^3'
convunits 1 MB/s Mb/s
convunits 30mpg L/100km
convunits 7.5L/100km mpg
```

List the catalog, or restrict it to a category:

```sh
convunits units
convunits units length
convunits units force
```

Control output:

```sh
convunits --precision 4 10kg lb
convunits --scientific 1e9m km
convunits --json 10kg lb
```

The default output uses ten significant digits.

## Temperatures and ambiguous SI symbols

Celsius and Fahrenheit are affine units. `C`, `F`, `K`, and `R` convert
directly, but Celsius and Fahrenheit cannot appear in compound expressions.
Kelvin and Rankine are scalar absolute-temperature units.

The symbols `C` and `F` also mean coulomb and farad. `convunits` resolves a
plain symbol using the other expression's dimensions. Long names (`celsius`,
`coulomb`, `fahrenheit`, `farad`) are available when explicit notation is
preferred. Inside compound expressions, `C` and `F` mean the electrical units.

## Fuel economy and consumption

Distance-per-volume fuel economy and volume-per-distance fuel consumption are
reciprocals. `convunits` handles the normalized `L/100km` notation explicitly:

```sh
convunits 30mpg L/100km
# 7.840486111 L/100km

convunits 5L/100km km/L
# 20 km/L
```

Values must be greater than zero because reciprocal conversion is undefined at
zero. Other distance-per-volume expressions, such as `mi/gal` and `km/L`, can
be used on the fuel-economy side.

## Constraint solving

`solve` mode uses the physical relationship `F = m*d/t^2` to infer one missing
force, mass, distance, or time value. The primary value and each `--given`
must identify distinct variables:

```sh
convunits solve 10N s --given mass=2kg --given distance=5m
# 1 s

convunits solve 2kg N --given distance=5m --given time=1s
# 10 N
```

Ranges use `minimum..maximum`. Bounds are propagated conservatively through
the monotonic force equation:

```sh
convunits solve 10N s --given mass=1..3kg --given distance=4..6m
# 0.632455532-1.341640786 s
```

Solve mode also accepts `--precision`, `--scientific`, and `--json`. It is a
specific physical-relation solver, not a mechanism for pretending arbitrary
dimension mismatches are directly convertible. Values are positive scalar
magnitudes; direction and signed vector components are outside this model.

## Non-linear scales

Non-linear, ordinal, and lookup conversions use a separate command and never
enter the dimensional unit engine:

```sh
convunits scale 10 dB power-ratio
convunits scale 20 dB amplitude-ratio
convunits scale 7 pH mol/L
convunits scale 5 mag brightness-ratio
convunits scale 5 beaufort m/s
convunits scale 12 awg diameter-mm
convunits scales
convunits scales ratio
```

The scale layer supports decibels/bels and power or amplitude ratios; pH and
hydrogen-ion concentration; stellar magnitude differences and brightness
ratios; Beaufort wind-force lookup ranges; and American wire gauge. Positive
ratios and concentrations are required for logarithmic conversion. Stellar
magnitude values are differences, not absolute photometric calibration.

Beaufort conversion returns a speed range. AWG uses numeric `-1`, `-2`, and
`-3` for 00, 000, and 0000 gauge. Scale output supports `--precision`,
`--scientific`, and `--json`; options may appear before or after operands.

## Paper sizes

Paper sizes are two-dimensional lookups rather than scalar scales:

```sh
convunits paper a4 mm
# 210 x 297 mm

convunits paper letter in
# 8.5 x 11 in
```

`size` and `scale-size` remain aliases for `paper`. Supported sizes are ISO A,
B, and C series from 0 through 10; US Letter, Legal, Tabloid, Ledger, Executive,
and common index cards; and common photo sizes. Output dimensions are converted
through the normal length-unit engine. List them with `papers`, `papers iso`,
`papers us`, or `papers photo`.

## Approximate shoe sizes

Shoe conversion reports approximate foot length, not a fit recommendation:

```sh
convunits shoe us-men 10 in
convunits shoe us-women 8.5 cm
convunits shoe uk-adult 9 cm
convunits shoe eu 43 cm
convunits shoe mondo 270 cm
convunits shoe jp 27 in
```

Supported systems are `us-men`, `us-women`, `uk-adult`, `eu`, `mondo`, and
`jp`. US, UK, and EU sizing formulas are approximate and brand-dependent.
Ambiguous `us` and all children's systems are rejected. Mondopoint input is
millimeters; JP input is centimeters. The output must be a length unit.

List the accepted system names with `convunits shoe systems`.

## Weird conversions

Convention- and table-based conversions use dedicated commands. They do not
add fake units to the dimensional registry.

Wire gauge reports approximate conductor diameter and accepts numeric AWG,
prefix/suffix forms, and zero-gauge notation:

```sh
convunits wire 12awg mm
convunits wire 0000awg in
convunits wires
```

Drill lookup supports 1/64-inch fractional increments, number drills `#80`
through `#1`, letter drills A through Z, and direct metric sizes:

```sh
convunits drill '1/4' mm
convunits drill '#7' mm
convunits drill A in
convunits drills number
```

Sieve lookup reports approximate nominal ASTM E11/US openings:

```sh
convunits sieve '#40' mm
convunits sieve 'No. 200' um
convunits sieve 4mesh mm
convunits sieves
```

Formula mode parses and dimension-checks every named input with the normal unit
engine before computing a result:

```sh
convunits formula escape-velocity --mass 1Mearth --radius 1Re km/s
convunits formula orbital-period --mass 1Msun --radius 1au d
convunits formula freefall-time --height 100m s
convunits formula pendulum-period --length 1m s
convunits formula bmi --mass 180lb --height 6ft bmi
convunits formulas
```

Available formulas are escape velocity, orbital period, orbital speed,
free-fall time, pendulum period, and BMI. The `bmi` output alias means
`kg/m^2`; it reports only the calculated scalar and provides no medical
interpretation. Idealized physical-formula results are labeled approximately;
BMI is reported as a direct scalar calculation.

## Scope

The built-in catalog covers the requested length, mass, time, area, volume,
speed, acceleration, force, energy, power, pressure, frequency, angle,
information, data-rate, electrical, substance, luminous, density, and flow
units. Reciprocal fuel economy/consumption conversion and force-relation
constraint solving are also supported explicitly.

## Expanded and approximate units

The built-in catalog also includes astronomical lengths, masses, times, and
speeds; ancient and historical measures; nautical and geologic units; atomic
and particle constants; radiation and radioactivity units; concentration,
typography, screen, computing, media-time, and human-scale units.

Historical and human-scale definitions are explicit approximations. View their
assumptions with the category listings:

```sh
convunits units ancient-length
convunits units astronomical-length
convunits units radiation
convunits units human-scale
```

Approximate entries include a short note, for example the common cubit uses 18
inches and `olympicpool` assumes a 50 m x 25 m x 2 m pool. Existing symbols
remain stable: `pt` is pint, `rad` is radian, and `lm` is lumen. The typographic
point is `point`, absorbed-dose rad is `rad-dose`, and light-minute is
`lightmin`.

The existing `tsp` is a US teaspoon. Use `teaspoon-metric` or
`metricteaspoon` for exactly 5 mL. `Kbps` is accepted as an explicit alias for
decimal kilobits per second, equivalent to `kbps`; byte symbols remain
case-sensitive.

Ordinal, logarithmic, subjective, and model-dependent scales are not linear
units. The supported cases use the explicit `scale` command. Redshift distance,
Mohs, semitones, octaves, Richter, and currency remain unsupported.
