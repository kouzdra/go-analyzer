package server

import "strings"
import "github.com/kouzdra/go-analyzer/defs"
import "github.com/kouzdra/go-analyzer/results"
import "github.com/kouzdra/go-analyzer/options"
import golang "github.com/kouzdra/go-analyzer/golang/project"
import "github.com/kouzdra/go-analyzer/iface/iproject"

type Project struct {
	*Server
	Project iproject.IProject
}

func NewProject(server *Server) *Project {
	return &Project{server, golang.NewProject ()}
}

//-------------------------------------------------------------------

func (p *Project) Complete (no int, fname string, pos defs.Pos) {
	if src, err := p.Project.GetSrc(fname); src != nil {
		results := p.Project.Complete (src, pos)
		if results != nil {
		    results.Write(p.Writer)
		}
	} else {
		p.Msg(err.Error())
	}
}

//-------------------------------------------------------------------

func (p *Project) Analyze (no int, fName string) {
	if src, err := p.Project.GetSrc(fName); src != nil {
		errors, fonts := p.Project.Analyze (src, no);
		p.Server.Writer.Beg("ERRORS-CLEAR").Eol().End("ERRORS-CLEAR")
		errors.Write(p.Server.Writer)
		if (options.Codeblocks) {
			fonts.Write(p.Server.Writer)
		}
	} else {
		p.Msg(err.Error())
	}
}

//-------------------------------------------------------------------

func (p *Project) FindFiles (no int, pfx string, system bool, max int) {
	
	files := results.Files{system, make([]string, 0, 1000)}
	for _, pkg := range p.Project.GetPackages () {
		for _, f := range pkg.GetSrcs () {
			if strings.HasPrefix(f.GetName().Name, pfx) {
				files.Files = append (files.Files, iproject.FName(f))
			}
		}
	}
	if len(files.Files) > max {
		files.Files = files.Files [:max]
	}
	p.Project.FindFiles (no, pfx, system, max).Write(p.Writer)
}
