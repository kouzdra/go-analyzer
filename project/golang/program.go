package golang

import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/project/iface"

type Program struct {
	Options OptionsBase
	Name *names.Name
	Package *Package
}

func (p *Program) GetOptions() iface.Options { // nil - no package
	return &p.Options
}

func (p *Program) GetName() *names.Name { // nil - no package
	return p.Name
}

func (p *Program) GetPackage(path *names.Name) iface.Package { // nil - no package
	return p.Package
}
