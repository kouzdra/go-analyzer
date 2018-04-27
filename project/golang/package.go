package golang

import "github.com/kouzdra/go-analyzer/project/iface"
//import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"

type Package struct {
	Path *paths.Path
	super *Program
}

func (p *Package) Program() iface.Program {
	return p.super
}
