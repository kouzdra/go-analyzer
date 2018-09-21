package iface

import "go/token"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/gproject"
//import "github.com/kouzdra/go-analyzer/paths"


type Project interface {
	GetOptions() Options
	GetFileSet() *token.FileSet
}

type Program interface {
	GetName() *names.Name
	GetProject() Project
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
	GetDir () *names.Name
	GetName() *names.Name
	GetPackage() gproject.Pkg
}
