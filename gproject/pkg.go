package gproject

import "go/build"
//import "fmt"
import "github.com/kouzdra/go-analyzer/names"
//import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/env"

type  Pkg struct {
	prj  *Prj
	dir  *names.Name
	name *names.Name
	pkg  *build.Package
	srcs map [*names.Name] Source
	envGbl *env.Env
	envLcl *env.Env
}

//--------------------------------------------------

func (p *Pkg) GetProject () *Prj { return p.prj; }
func (p *Pkg) GetDir     () *names.Name { return p.dir ; }
func (p *Pkg) GetName    () *names.Name { return p.name; }
func (p *Pkg) GetSrcs    () map [*names.Name] Source { return p.srcs; }
func (p *Pkg) GetPackage () *build.Package { return p.pkg ; }
func (p *Pkg) GetEnvLcl  () *env.Env { return p.envLcl; }
func (p *Pkg) GetEnvGbl  () *env.Env { return p.envGbl; }


//--------------------------------------------------

func NewPkg (p *Prj, bpkg *build.Package) *Pkg {
	return &Pkg{p, names.Put (bpkg.Dir), names.Put (bpkg.Name), bpkg, make (map [*names.Name]Source), nil, nil}
}

func (pkg *Pkg) Reload () {
	pkg.envLcl = nil
	pkg.envGbl = nil
}


func (pkg *Pkg) UpdateAsts () {
	if pkg.GetEnvGbl() == nil || pkg.GetEnvLcl() == nil {
		gbl := env.NewBldr ()
		lcl := env.NewBldr ()
		//lcl.Nested (gbl)
		//pkg.Prj.Server.MsgF ("+Update Package %s\n", pkg.Name)
		for _, src := range pkg.GetSrcs() {
                        //pkg.Prj.Server.MsgF ("+--Update File %s\n", src.Name)
                        src.UpdateAst ()
		}
		pkg.envGbl = gbl.Close ()
		pkg.envLcl = lcl.Close ()
	}
}
