package units

import (
	"fmt"
	"strings"
)

// Dimension contains integer exponents for the supported base dimensions.
type Dimension struct {
	Length, Mass, Time, ElectricCurrent, Temperature int
	Amount, LuminousIntensity, Angle, Information    int
}

func (d Dimension) Add(o Dimension) Dimension {
	return Dimension{d.Length + o.Length, d.Mass + o.Mass, d.Time + o.Time,
		d.ElectricCurrent + o.ElectricCurrent, d.Temperature + o.Temperature,
		d.Amount + o.Amount, d.LuminousIntensity + o.LuminousIntensity,
		d.Angle + o.Angle, d.Information + o.Information}
}

func (d Dimension) Sub(o Dimension) Dimension { return d.Add(o.Mul(-1)) }

func (d Dimension) Mul(n int) Dimension {
	return Dimension{d.Length * n, d.Mass * n, d.Time * n, d.ElectricCurrent * n,
		d.Temperature * n, d.Amount * n, d.LuminousIntensity * n, d.Angle * n, d.Information * n}
}

func (d Dimension) String() string {
	parts := []struct {
		s string
		n int
	}{{"kg", d.Mass}, {"m", d.Length}, {"A", d.ElectricCurrent}, {"s", d.Time},
		{"K", d.Temperature}, {"mol", d.Amount}, {"cd", d.LuminousIntensity},
		{"rad", d.Angle}, {"B", d.Information}}
	num, den := "", ""
	for _, p := range parts {
		if p.n > 0 {
			num = appendDim(num, p.s, p.n)
		}
		if p.n < 0 {
			den = appendDim(den, p.s, -p.n)
		}
	}
	if num == "" {
		num = "1"
	}
	if den != "" {
		if strings.Contains(den, "*") {
			return num + "/(" + den + ")"
		}
		return num + "/" + den
	}
	return num
}

func appendDim(dst, symbol string, power int) string {
	if dst != "" {
		dst += "*"
	}
	if power == 1 {
		return dst + symbol
	}
	return dst + fmt.Sprintf("%s^%d", symbol, power)
}

func (d Dimension) Category() string {
	known := map[Dimension]string{
		{Length: 1}: "length", {Mass: 1}: "mass", {Time: 1}: "time",
		{Temperature: 1}: "temperature", {Length: 2}: "area", {Length: 3}: "volume",
		{Length: 1, Time: -1}: "speed", {Length: 1, Time: -2}: "acceleration",
		{Mass: 1, Length: 1, Time: -2}: "force", {Mass: 1, Length: 2, Time: -2}: "energy",
		{Mass: 1, Length: 2, Time: -3}: "power", {Mass: 1, Length: -1, Time: -2}: "pressure",
		{Time: -1}: "frequency", {Angle: 1}: "angle", {Information: 1}: "information",
		{ElectricCurrent: 1}: "electric current", {Amount: 1}: "amount", {LuminousIntensity: 1}: "luminous intensity",
	}
	if s, ok := known[d]; ok {
		return s
	}
	return d.String()
}
