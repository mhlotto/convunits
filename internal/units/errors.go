package units

import "fmt"

type UnknownUnitError struct{ Name string }

func (e UnknownUnitError) Error() string { return fmt.Sprintf("unknown unit %q", e.Name) }

type ParseError struct{ Message string }

func (e ParseError) Error() string { return "invalid unit expression: " + e.Message }

type IncompatibleError struct {
	Input, Output                   string
	InputDimension, OutputDimension Dimension
}

func (e IncompatibleError) Error() string {
	return fmt.Sprintf("cannot convert %s to %s: incompatible dimensions\n%s has dimensions: %s\n%s has dimensions: %s\n\nA direct conversion is impossible without additional constraints.",
		e.Input, e.Output, e.Input, e.InputDimension, e.Output, e.OutputDimension)
}
