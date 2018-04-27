package golang

import "github.com/kouzdra/go-analyzer/project/iface"
import "github.com/kouzdra/go-analyzer/names"

type Package struct {
	Path *names.Name
	super *Program
}

func (pkg *Package) GetPath() *names.Name {
	return pkg.Path
}

func (pkg *Package) GetProgram() iface.Program {
	return pkg.super
}

func (pkg *Package) GetSource(name *names.Name) iface.Source {
	return nil
}
