package golang

import "go/token"
//import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/project/iface"

type Project struct {
	Options OptionsBase
	Program *Program
	FileSet *token.FileSet
	//ModeTab *env.ModeTab
}

func (p *Project) GetOptions() iface.Options {
	return &p.Options
}

func (p *Project) GetProgram() iface.Program {
	return p.Program
}

func (p *Project) GetFileSet() *token.FileSet {
	return p.FileSet
}
