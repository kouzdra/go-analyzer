package iface

import "fmt"
import "github.com/kouzdra/go-analyzer/iface/iproject"

type Loc struct {
	Src iproject.ISource
	Pos uint
}

func NewLoc(src iproject.ISource, pos uint) Loc {
	return Loc{src, pos}
}

func (loc *Loc) Repr() string {
	return fmt.Sprintf("%s:%d", iproject.FName(loc.Src), loc.Pos)
}
