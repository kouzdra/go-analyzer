package project

import "fmt"

type Loc struct {
	Src *Src
	Pos uint
}

func (src *Src) NewLoc(pos uint) Loc {
	return Loc{src, pos}
}

func (loc *Loc) Repr() string {
	return fmt.Sprintf("%s:%d", loc.Src.FName(), loc.Pos)
}
