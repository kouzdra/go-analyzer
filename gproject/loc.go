package gproject

import "fmt"

type Loc struct {
	Src Source
	Pos uint
}

func (src *src) NewLoc(pos uint) Loc {
	return Loc{src, pos}
}

func (loc *Loc) Repr() string {
	return fmt.Sprintf("%s:%d", FName(loc.Src), loc.Pos)
}
