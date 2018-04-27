package golang

import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/project/iface"

type Program struct {
	options OptionsBase
}

func (p *Program) Options() iface.Options { // nil - no package
	return &p.options
}

func (p *Program) Name() *names.Name { // nil - no package
	return nil
}

func (p *Program) GetPackage(path *names.Name) iface.Package { // nil - no package
	return nil
}
