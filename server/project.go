package server

import "strings"
import "github.com/kouzdra/go-analyzer/results"
import "github.com/kouzdra/go-analyzer/options"
import "github.com/kouzdra/go-analyzer/gproject"

type Project struct {
	*Server
	Project *gproject.Prj
}

func NewProject(server *Server) *Project {
	return &Project{server, gproject.NewProject ()}
}

//-------------------------------------------------------------------

func (p *Project) Complete (no int, fname string, pos int) {
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
				files.Files = append (files.Files, gproject.FName(f))
			}
		}
	}
	if len(files.Files) > max {
		files.Files = files.Files [:max]
	}
	p.Project.FindFiles (no, pfx, system, max).Write(p.Writer)
}
