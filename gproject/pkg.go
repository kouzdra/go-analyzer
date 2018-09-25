package gproject

import "go/build"
//import "fmt"
import "github.com/kouzdra/go-analyzer/names"
//import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/env"

type  pkg struct {
	prj  *prj
	dir  *names.Name
	name *names.Name
	pkg  *build.Package
	srcs map [*names.Name] Source
	envGbl *env.Env
	envLcl *env.Env
}

//--------------------------------------------------

func (p *pkg) GetProject () *prj { return p.prj; }
func (p *pkg) GetDir     () *names.Name { return p.dir ; }
func (p *pkg) GetName    () *names.Name { return p.name; }
func (p *pkg) GetSrcs    () map [*names.Name] Source { return p.srcs; }
func (p *pkg) GetPackage () *build.Package { return p.pkg ; }
func (p *pkg) GetEnvLcl  () *env.Env { return p.envLcl; }
func (p *pkg) GetEnvGbl  () *env.Env { return p.envGbl; }


//--------------------------------------------------

func newPkg (p *prj, bpkg *build.Package) *pkg {
	return &pkg{p, names.Put (bpkg.Dir), names.Put (bpkg.Name), bpkg, make (map [*names.Name]Source), nil, nil}
}

func (pkg *pkg) Reload () {
	pkg.envLcl = nil
	pkg.envGbl = nil
}


func (pkg *pkg) UpdateAsts () {
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
