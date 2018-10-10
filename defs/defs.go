package defs

type Pos int
type Rng struct { Beg, End Pos }
func NewRng (beg, end Pos) Rng { return Rng{beg, end} }
