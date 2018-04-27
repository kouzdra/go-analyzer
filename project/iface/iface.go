package iface

import "github.com/kouzdra/go-analyzer/names"
//import "github.com/kouzdra/go-analyzer/paths"

type Program interface {
	GetName() *names.Name
	GetOptions() Options
	//Packages() []Package

	//Update()
	//UpdatePkg(pkg Package)
	//UpdateSrc(src Source)

	GetPackage(path *names.Name) Package // nil - no package
}

type Package interface {
	GetPath() *names.Name
	GetProgram() Program
	//GetSources() []Source
	GetSource(name *names.Name) Source // nil - no source
}

type Source interface {
	GetName() *names.Name
	GetPackage() Package
}
