package units

type Unit struct {
	Symbol, Name, Category string
	Aliases                []string
	Dimension              Dimension
	Multiplier             float64
	Offset                 float64 // base = value*Multiplier + Offset
	Affine, Approximate    bool
	Note                   string
	SourceNote             string
}

func (u *Unit) ToBase(v float64) float64   { return v*u.Multiplier + u.Offset }
func (u *Unit) FromBase(v float64) float64 { return (v - u.Offset) / u.Multiplier }

type Expression struct {
	Text        string
	Dimension   Dimension
	Multiplier  float64
	Affine      *Unit
	Approximate bool
}

func (e Expression) ToBase(v float64) float64 {
	if e.Affine != nil {
		return e.Affine.ToBase(v)
	}
	return v * e.Multiplier
}

func (e Expression) FromBase(v float64) float64 {
	if e.Affine != nil {
		return e.Affine.FromBase(v)
	}
	return v / e.Multiplier
}
