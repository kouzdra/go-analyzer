package iface

import "github.com/kouzdra/go-analyzer/names"
//import "github.com/kouzdra/go-analyzer/paths"

type Program interface {
	Name() *names.Name
	Options() Options
	//Packages() []Package

	//Update()
	//UpdatePkg(pkg Package)
	//UpdateSrc(src Source)

	GetPackage(path *names.Name) Package // nil - no package
}

type Package interface {
	Name() *names.Name
	Path() *names.Name
	Program() Program
	Sources() []Source
	GetSource(name *names.Name) Source // nil - no source
}

type Source interface {
	Name() *names.Name
	Package() Package
}
