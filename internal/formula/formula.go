package formula

import (
	"fmt"
	"math"
	"sort"

	"convunits/internal/units"
)

const (
	gravitationalConstant = 6.67430e-11
	standardGravity       = 9.80665
	speedOfLight          = 299792458
)

type Input struct {
	Value float64
	Unit  string
}

type Result struct {
	Value       float64
	Unit        string
	Approximate bool
}

type Definition struct {
	Name, Arguments, OutputDimension string
}

var definitions = []Definition{
	{"bmi", "--mass MASS --height HEIGHT", "mass/length^2"},
	{"circle-area", "--radius RADIUS", "area"},
	{"cylinder-volume", "--radius RADIUS --height HEIGHT", "volume"},
	{"density", "--mass MASS --volume VOLUME", "density"},
	{"energy", "--power POWER --time TIME", "energy"},
	{"escape-velocity", "--mass MASS --radius RADIUS", "speed"},
	{"flow-rate", "--volume VOLUME --time TIME", "volume/time"},
	{"freefall-time", "--height HEIGHT [--gravity ACCELERATION]", "time"},
	{"gravity-force", "--mass1 MASS --mass2 MASS --distance DISTANCE", "force"},
	{"kinetic-energy", "--mass MASS --speed SPEED", "energy"},
	{"momentum", "--mass MASS --speed SPEED", "momentum"},
	{"orbital-period", "--mass MASS --radius RADIUS", "time"},
	{"orbital-speed", "--mass MASS --radius RADIUS", "speed"},
	{"pace", "--distance DISTANCE --time TIME", "time/length"},
	{"pendulum-period", "--length LENGTH [--gravity ACCELERATION]", "time"},
	{"power", "--energy ENERGY --time TIME", "power"},
	{"pressure", "--force FORCE --area AREA", "pressure"},
	{"schwarzschild-radius", "--mass MASS", "length"},
	{"sphere-area", "--radius RADIUS", "area"},
	{"sphere-volume", "--radius RADIUS", "volume"},
	{"speed", "--distance DISTANCE --time TIME", "speed"},
	{"surface-gravity", "--mass MASS --radius RADIUS", "acceleration"},
}

func Definitions() []Definition {
	out := append([]Definition(nil), definitions...)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

type Calculator struct{ Registry *units.Registry }

func New(registry *units.Registry) *Calculator { return &Calculator{Registry: registry} }

func (c *Calculator) Compute(name string, inputs map[string]Input, outputUnit string) (Result, error) {
	allowedByFormula := map[string]map[string]bool{
		"escape-velocity":      {"mass": true, "radius": true},
		"orbital-period":       {"mass": true, "radius": true},
		"orbital-speed":        {"mass": true, "radius": true},
		"freefall-time":        {"height": true, "gravity": true},
		"pendulum-period":      {"length": true, "gravity": true},
		"bmi":                  {"mass": true, "height": true},
		"schwarzschild-radius": {"mass": true},
		"gravity-force":        {"mass1": true, "mass2": true, "distance": true},
		"surface-gravity":      {"mass": true, "radius": true},
		"density":              {"mass": true, "volume": true},
		"sphere-volume":        {"radius": true},
		"sphere-area":          {"radius": true},
		"circle-area":          {"radius": true},
		"cylinder-volume":      {"radius": true, "height": true},
		"kinetic-energy":       {"mass": true, "speed": true},
		"momentum":             {"mass": true, "speed": true},
		"power":                {"energy": true, "time": true},
		"energy":               {"power": true, "time": true},
		"pressure":             {"force": true, "area": true},
		"flow-rate":            {"volume": true, "time": true},
		"pace":                 {"distance": true, "time": true},
		"speed":                {"distance": true, "time": true},
	}
	allowed, known := allowedByFormula[name]
	if !known {
		return Result{}, fmt.Errorf("unknown formula %q", name)
	}
	for argument := range inputs {
		if !allowed[argument] {
			return Result{}, fmt.Errorf("formula %s does not accept --%s", name, argument)
		}
	}
	var value float64
	var canonicalOutput string
	approximate := true
	switch name {
	case "escape-velocity":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = math.Sqrt(2 * gravitationalConstant * mass / radius)
		canonicalOutput = "m/s"
	case "orbital-period":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = 2 * math.Pi * math.Sqrt(radius*radius*radius/(gravitationalConstant*mass))
		canonicalOutput = "s"
	case "orbital-speed":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = math.Sqrt(gravitationalConstant * mass / radius)
		canonicalOutput = "m/s"
	case "freefall-time":
		height, err := c.required(inputs, "height", "m")
		if err != nil {
			return Result{}, err
		}
		gravity, err := c.optional(inputs, "gravity", "m/s^2", standardGravity)
		if err != nil {
			return Result{}, err
		}
		value = math.Sqrt(2 * height / gravity)
		canonicalOutput = "s"
	case "pendulum-period":
		length, err := c.required(inputs, "length", "m")
		if err != nil {
			return Result{}, err
		}
		gravity, err := c.optional(inputs, "gravity", "m/s^2", standardGravity)
		if err != nil {
			return Result{}, err
		}
		value = 2 * math.Pi * math.Sqrt(length/gravity)
		canonicalOutput = "s"
	case "bmi":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		height, err := c.required(inputs, "height", "m")
		if err != nil {
			return Result{}, err
		}
		value = mass / (height * height)
		canonicalOutput = "kg/m^2"
		approximate = false
	case "schwarzschild-radius":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		value = 2 * gravitationalConstant * mass / (speedOfLight * speedOfLight)
		canonicalOutput = "m"
	case "gravity-force":
		mass1, err := c.required(inputs, "mass1", "kg")
		if err != nil {
			return Result{}, err
		}
		mass2, err := c.required(inputs, "mass2", "kg")
		if err != nil {
			return Result{}, err
		}
		distance, err := c.required(inputs, "distance", "m")
		if err != nil {
			return Result{}, err
		}
		value = gravitationalConstant * mass1 * mass2 / (distance * distance)
		canonicalOutput = "N"
	case "surface-gravity":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = gravitationalConstant * mass / (radius * radius)
		canonicalOutput = "m/s^2"
	case "density":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		volume, err := c.required(inputs, "volume", "m^3")
		if err != nil {
			return Result{}, err
		}
		value = mass / volume
		canonicalOutput = "kg/m^3"
	case "sphere-volume":
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = 4 * math.Pi * radius * radius * radius / 3
		canonicalOutput = "m^3"
	case "sphere-area":
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = 4 * math.Pi * radius * radius
		canonicalOutput = "m^2"
	case "circle-area":
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		value = math.Pi * radius * radius
		canonicalOutput = "m^2"
	case "cylinder-volume":
		radius, err := c.required(inputs, "radius", "m")
		if err != nil {
			return Result{}, err
		}
		height, err := c.required(inputs, "height", "m")
		if err != nil {
			return Result{}, err
		}
		value = math.Pi * radius * radius * height
		canonicalOutput = "m^3"
	case "kinetic-energy":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		speed, err := c.required(inputs, "speed", "m/s")
		if err != nil {
			return Result{}, err
		}
		value = 0.5 * mass * speed * speed
		canonicalOutput = "J"
	case "momentum":
		mass, err := c.required(inputs, "mass", "kg")
		if err != nil {
			return Result{}, err
		}
		speed, err := c.required(inputs, "speed", "m/s")
		if err != nil {
			return Result{}, err
		}
		value = mass * speed
		canonicalOutput = "kg*m/s"
	case "power":
		energy, err := c.required(inputs, "energy", "J")
		if err != nil {
			return Result{}, err
		}
		time, err := c.required(inputs, "time", "s")
		if err != nil {
			return Result{}, err
		}
		value = energy / time
		canonicalOutput = "W"
	case "energy":
		power, err := c.required(inputs, "power", "W")
		if err != nil {
			return Result{}, err
		}
		time, err := c.required(inputs, "time", "s")
		if err != nil {
			return Result{}, err
		}
		value = power * time
		canonicalOutput = "J"
	case "pressure":
		force, err := c.required(inputs, "force", "N")
		if err != nil {
			return Result{}, err
		}
		area, err := c.required(inputs, "area", "m^2")
		if err != nil {
			return Result{}, err
		}
		value = force / area
		canonicalOutput = "Pa"
	case "flow-rate":
		volume, err := c.required(inputs, "volume", "m^3")
		if err != nil {
			return Result{}, err
		}
		time, err := c.required(inputs, "time", "s")
		if err != nil {
			return Result{}, err
		}
		value = volume / time
		canonicalOutput = "m^3/s"
	case "pace":
		distance, err := c.required(inputs, "distance", "m")
		if err != nil {
			return Result{}, err
		}
		time, err := c.required(inputs, "time", "s")
		if err != nil {
			return Result{}, err
		}
		value = time / distance
		canonicalOutput = "s/m"
	case "speed":
		distance, err := c.required(inputs, "distance", "m")
		if err != nil {
			return Result{}, err
		}
		time, err := c.required(inputs, "time", "s")
		if err != nil {
			return Result{}, err
		}
		value = distance / time
		canonicalOutput = "m/s"
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Result{}, fmt.Errorf("formula %s produced a non-finite result", name)
	}
	requested := outputUnit
	convertUnit := outputUnit
	if outputUnit == "bmi" {
		if name != "bmi" {
			return Result{}, fmt.Errorf("output alias bmi is only valid for the bmi formula")
		}
		convertUnit = "kg/m^2"
	}
	converted, err := c.Registry.Convert(value, canonicalOutput, convertUnit)
	if err != nil {
		return Result{}, fmt.Errorf("output unit: %w", err)
	}
	return Result{Value: converted.Value, Unit: requested, Approximate: approximate}, nil
}

func (c *Calculator) required(inputs map[string]Input, name, canonical string) (float64, error) {
	input, ok := inputs[name]
	if !ok {
		return 0, fmt.Errorf("missing required argument --%s", name)
	}
	return c.convertPositive(name, input, canonical)
}

func (c *Calculator) optional(inputs map[string]Input, name, canonical string, fallback float64) (float64, error) {
	input, ok := inputs[name]
	if !ok {
		return fallback, nil
	}
	return c.convertPositive(name, input, canonical)
}

func (c *Calculator) convertPositive(name string, input Input, canonical string) (float64, error) {
	converted, err := c.Registry.Convert(input.Value, input.Unit, canonical)
	if err != nil {
		return 0, fmt.Errorf("--%s: %w", name, err)
	}
	if converted.Value <= 0 {
		return 0, fmt.Errorf("--%s must be greater than zero", name)
	}
	return converted.Value, nil
}
