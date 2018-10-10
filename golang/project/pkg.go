package project

import "go/build"
//import "fmt"
import "github.com/kouzdra/go-analyzer/names"
//import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/golang/env"
import "github.com/kouzdra/go-analyzer/iface/iproject"

type  pkg struct {
	prj  *prj
	dir  *names.Name
	name *names.Name
	pkg  *build.Package
	srcs map [*names.Name] iproject.ISource
	envGbl *env.Env
	envLcl *env.Env
}

//--------------------------------------------------

func (p *pkg) GetProject () iproject.IProject { return p.prj; }
func (p *pkg) GetDir     () *names.Name { return p.dir ; }
func (p *pkg) GetName    () *names.Name { return p.name; }
func (p *pkg) GetSrcs    () map [*names.Name] iproject.ISource { return p.srcs; }
func (p *pkg) GetPackage () *build.Package { return p.pkg ; }
func (p *pkg) GetEnvLcl  () *env.Env { return p.envLcl; }
func (p *pkg) GetEnvGbl  () *env.Env { return p.envGbl; }


//--------------------------------------------------

func newPkg (p *prj, bpkg *build.Package) *pkg {
	return &pkg{p, names.Put (bpkg.Dir), names.Put (bpkg.Name), bpkg, make (map [*names.Name] iproject.ISource), nil, nil}
}

func (pkg *pkg) Reload () {
	pkg.envLcl = nil
	pkg.envGbl = nil
}


func (pkg *pkg) updateAsts () {
	if pkg.GetEnvGbl() == nil || pkg.GetEnvLcl() == nil {
		gbl := env.NewBldr ()
		lcl := env.NewBldr ()
		//lcl.Nested (gbl)
		//pkg.Prj.Server.MsgF ("+Update Package %s\n", pkg.Name)
		for _, src := range pkg.GetSrcs() {
                        //pkg.Prj.Server.MsgF ("+--Update File %s\n", src.Name)
                        src.(*source).updateAst ()
		}
		pkg.envGbl = gbl.Close ()
		pkg.envLcl = lcl.Close ()
	}
}
