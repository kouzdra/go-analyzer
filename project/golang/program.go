package golang

import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/project/iface"

type Program struct {
	Project *Project
	Name *names.Name
	Package *Package
}

func (p *Program) GetProject() iface.Project { 
	return p.Project
}

func (p *Program) GetName() *names.Name { 
	return p.Name
}

func (p *Program) GetPackage(path *names.Name) iface.Package { // nil - no package
	return p.Package
}
