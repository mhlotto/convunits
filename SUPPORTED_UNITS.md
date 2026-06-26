# Supported units and commands

This is a categorized reference for the built-in catalog and non-dimensional
commands. For the exact generated unit table, run:

```sh
convunits units
```

For a filtered generated table:

```sh
convunits units length
convunits units astronomical-length
convunits units ancient-length
convunits units human-scale
convunits scales
convunits formulas
```

Parsing is case-sensitive. Unit expressions support `*`, `/`, integer powers,
and parentheses.

## SI and common units

The core catalog includes the usual SI and common engineering units:

- length: `m`, `km`, `cm`, `mm`, `um`, `nm`, `in`, `ft`, `yd`, `mi`, `nmi`,
  `au`
- mass: `kg`, `g`, `mg`, `ug`, `lb`, `oz`, `ton`, `tonne`, `slug`
- time: `s`, `ms`, `min`, `h`, `d`, `wk`, `y`
- area and volume: `m^2`, `ha`, `acre`, `m^3`, `L`, `mL`, `gal`, `qt`, `pt`,
  `cup`, `floz`, `tbsp`, `tsp`
- speed and acceleration: `m/s`, `km/h`, `mph`, `fps`, `knot`, `g0`
- force, energy, power, pressure: `N`, `lbf`, `J`, `kWh`, `BTU`, `W`, `hp`,
  `Pa`, `bar`, `atm`, `psi`, `torr`
- frequency and angle: `Hz`, `rpm`, `rad`, `deg`, `arcmin`, `arcsec`
- electrical units: `A`, `C`, `V`, `Ohm`, `F`, `H`, `Wb`, `T`
- substance and luminous units: `mol`, `Molar`, `cd`, `lm`, `lx`
- flow and density expressions: `gpm`, `gal/min`, `kg/m^3`, `g/cm^3`

Examples:

```sh
convunits 10kg lb
convunits 60mph km/h
convunits 1N 'kg*m/s^2'
convunits 1 'g/cm^3' 'kg/m^3'
convunits 1gal/min gpm
```

## Astronomical units

Astronomical units are linear units in the normal conversion engine when they
have physical dimensions:

- lengths: `ls`, `lightmin`, `lh`, `ld`, `au`, `LD`, `Re`, `Rj`, `Rsun`
- masses: `Mearth`, `Mjup`, `Msun`
- times: `siderealday`, `siderealyear`, `galacticyear`
- speeds: `c0`, `earthorbitalvelocity`, `escapeearth`

Examples:

```sh
convunits 1Rsun km
convunits 1Mearth kg
convunits 1c0 mph
convunits formula schwarzschild-radius --mass 1Msun km
```

## Ancient and historical units

Ancient, nautical, and survey units are approximate conventions where history
does not define a single universal value:

- ancient length: `cubit`, `royalcubit`, `span`, `handbreadth`,
  `fingerbreadth`, `pace`, `romanfoot`, `romanmile`, `stadion`, `parasang`
- nautical and historical length: `fathom`, `cable`, `shot`, `nleague`,
  `league`, `chain`, `rod`, `furlong`, `hand`
- historical area: `rood`, `section`, `township`

Examples:

```sh
convunits 1cubit in
convunits 1fathom ft
convunits units ancient-length
```

## Human-scale units

Human-scale units are explicit approximations for memorable comparisons:

- `banana`: 18 cm banana length convention
- `smoot`: nominal Oliver Smoot height
- `footballfield`: 100 yards, excluding end zones
- `marathon`: official marathon distance
- `earthcircumference`: Earth equatorial circumference
- `olympicpool`: assumed 50 m x 25 m x 2 m pool volume

Examples:

```sh
convunits 1smoot ft
convunits 1banana cm
convunits 1olympicpool gal
```

## Atomic and physics units

The catalog includes particle/atomic constants and small physics-area units:

- mass: `u`, `Da`, `me`, `mp`, `mn`
- length: `angstrom`, `bohr`, `fermi`
- energy: `eV`, `hartree`, `rydberg`, `erg`
- area: `barn`, `shed`, `outhouse`

Examples:

```sh
convunits 1eV J
convunits 1angstrom nm
convunits 1barn m^2
```

## Radiation units

Radiation and radioactivity units are included with explicit category notes:

- absorbed dose / dose equivalent dimension: `Gy`, `rad-dose`, `Sv`, `rem`
- radioactivity: `Bq`, `Ci`

`rad` remains the angle unit. Use `rad-dose` for radiation absorbed dose.

Examples:

```sh
convunits 1Gy rad-dose
convunits 1Ci Bq
```

## Information units

Information units are byte-based internally. Bit units are represented as
one-eighth of a byte:

- decimal bytes: `B`, `kB`, `MB`, `GB`, `TB`, `PB`, `EB`
- decimal bits: `b`, `kb`, `Mb`, `Gb`, `Tb`, `Pb`, `Eb`
- binary bytes: `KiB`, `MiB`, `GiB`, `TiB`, `PiB`, `EiB`
- data rates: `bps`, `kbps`, `Mbps`, `Gbps`, `Kbps`
- computing conveniences: `nibble`, `word16`, `word32`, `word64`, `block512`,
  `page4k`

Examples:

```sh
convunits 1MB Mb
convunits 1MiB MB
convunits 1MB/s Mb/s
```

## Approximate units

Approximate units are marked in the generated catalog. They include many
astronomical nominal values, ancient and historical conventions, human-scale
comparisons, and some typography/screen units.

Useful listings:

```sh
convunits units astronomical-length
convunits units astronomical-mass
convunits units ancient-length
convunits units historical-length
convunits units human-scale
convunits units typography-length
```

Examples:

```sh
convunits 1cubit in
convunits 1Rsun km
convunits 1smoot ft
```

## Weird and lookup commands

These commands are supported, but deliberately separate from the dimensional
unit registry:

- `convunits paper SIZE UNIT`
- `convunits shoe SYSTEM SIZE UNIT`
- `convunits wire GAUGE UNIT`
- `convunits drill SIZE UNIT`
- `convunits sieve SIZE UNIT`
- `convunits scale VALUE INPUT-SCALE OUTPUT-SCALE`
- `convunits formula NAME --ARG VALUEUNIT OUTPUT-UNIT`
- `convunits compare VALUEUNIT TARGET-UNIT...`
- `convunits eval 'EXPRESSION [-> OUTPUT-UNIT]'`
- `convunits explain VALUEUNIT OUTPUT-UNIT`
- `convunits recipe AMOUNT INGREDIENT OUTPUT-UNIT`

Examples:

```sh
convunits paper a4 mm
convunits shoe us-men 10 yd
convunits wire 12awg mm
convunits drill '#7' mm
convunits sieve 'No. 200' um
convunits scale 5 beaufort mph
convunits formula bmi --mass 180lb --height 6ft bmi
convunits compare 38in banana smoot Rj
convunits eval '2*pi*1Re -> km'
convunits explain 60mph m/s
convunits recipe 1cup flour g
```

## Recipe ingredients

Recipe ingredients are approximate cooking density entries, not normal units.
List them with:

```sh
convunits recipe ingredients
convunits recipe ingredients baking
```

Supported categories include liquids, baking, fats, grains, and produce.
Conversions can vary by packing, brand, humidity, grind, ingredient form, and
measurement method.

## Unsupported things by design

- Currency is not supported.
- pH, dB, Beaufort, and magnitude are scale conversions, not normal units.
- pH is not generic concentration conversion.
- Beaufort is range/lookup based.
- Redshift is not treated as a distance.
- Mohs hardness is not implemented as a conversion.
- Shoe sizes are approximate foot-length mappings, not fit recommendations.
- Recipe conversions are approximate cooking estimates and ingredient-specific.
- Ancient units are approximate conventions.
- BMI is only calculated, not interpreted medically.

## Notes on symbol stability

Some symbols have common conflicts. `convunits` keeps existing meanings stable:

- `pt` is pint; use `point` for typographic point.
- `rad` is radian; use `rad-dose` for radiation absorbed dose.
- `lm` is lumen; use `lightmin` for light-minute.
- `C` and `F` can mean Celsius/Fahrenheit or coulomb/farad depending on
  context. Use long names such as `celsius`, `fahrenheit`, `coulomb`, and
  `farad` when explicit notation is clearer.
- `tsp` is a US teaspoon. Use `teaspoon-metric` or `metricteaspoon` for
  exactly 5 mL.
