package project

import "go/build"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/golang/env"
import "github.com/kouzdra/go-analyzer/iface/iproject"

type packag struct {
	prj  *project
	dir  *names.Name
	name *names.Name
	pkg  *build.Package
	srcs map [*names.Name] iproject.ISource
	envGbl *env.Env
	envLcl *env.Env
}

//--------------------------------------------------

func (p *packag) GetProject () iproject.IProject { return p.prj; }
func (p *packag) GetDir     () *names.Name { return p.dir ; }
func (p *packag) GetName    () *names.Name { return p.name; }
func (p *packag) GetSrcs    () map [*names.Name] iproject.ISource { return p.srcs; }
func (p *packag) GetPackage () *build.Package { return p.pkg ; }
func (p *packag) GetEnvLcl  () *env.Env { return p.envLcl; }
func (p *packag) GetEnvGbl  () *env.Env { return p.envGbl; }


//--------------------------------------------------

func newPkg (p *project, bpkg *build.Package) *packag {
	return &packag{p, names.Put (bpkg.Dir), names.Put (bpkg.Name), bpkg, make (map [*names.Name] iproject.ISource), nil, nil}
}

func (pkg *packag) Reload () {
	pkg.envLcl = nil
	pkg.envGbl = nil
}


func (pkg *packag) updateAsts () {
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
