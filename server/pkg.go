package server

import "go/build"
//import "fmt"
import "github.com/kouzdra/go-analyzer/env"

type  Pkg struct {
	Prj *Project
	Dir  string
	Name string
	Pkg *build.Package
	Srcs map [string] *Src
	EnvGbl *env.Env
	EnvLcl *env.Env
}

func NewPkg (p *Project, bpkg *build.Package) *Pkg {
	return &Pkg{p, bpkg.Dir, bpkg.Name, bpkg, make (map [string]*Src), nil, nil}
}

func (pkg *Pkg) Reload () {
	pkg.EnvLcl = nil
	pkg.EnvGbl = nil
}


func (pkg *Pkg) UpdateAsts () {
	if pkg.EnvGbl == nil || pkg.EnvLcl == nil {
		gbl := env.NewBldr ()
		lcl := env.NewBldr ()
		//lcl.Nested (gbl)
		pkg.Prj.Server.MsgF ("+Update Package %s\n", pkg.Name)
		for _, src := range pkg.Srcs {
                        pkg.Prj.Server.MsgF ("+--Update File %s\n", src.Name)
                        src.UpdateAst ()
		}
		pkg.EnvGbl = gbl.Close ()
		pkg.EnvLcl = lcl.Close ()
	}
}
