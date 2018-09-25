package iface

import "fmt"

type Loc struct {
	Src Source
	Pos uint
}

func NewLoc(src Source, pos uint) Loc {
	return Loc{src, pos}
}

func (loc *Loc) Repr() string {
	return fmt.Sprintf("%s:%d", FName(loc.Src), loc.Pos)
}
