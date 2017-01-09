package server

import "fmt"
import "strings"
import "io/ioutil"
import "path"
import "path/filepath"
import "go/build"
import "go/token"
import "go/ast"
import "env"
import "analyzer"
import "results"
import "options"

type Project struct {
	*Server
	Context build.Context
	Dirs [] string
	Pkgs map [string]*Pkg
	FSet *token.FileSet
	ModeTab *env.ModeTab
}

func NewProject(server *Server) *Project {
	return &Project{server, build.Default, nil, nil, token.NewFileSet(), env.NewModeTab ()}
}

func (p *Project) SetRoot (path string) {
	p.Context.GOROOT = path
}

func (p *Project) SetPath (path string) {
	p.Context.GOPATH = path
}

//-------------------------------------------------------------------

func (p *Project) GetSubDirs (dest [] string, root string) [] string {
	dest = append (dest, root)
	//fmt.Printf ("File: [%s]\n", root)
	files, _ := ioutil.ReadDir (root)
	for _, file := range files {
		if file.IsDir () {
			dest = p.GetSubDirs (dest, filepath.Join (root, file.Name()))
		}
	}
	return dest
}

func (p *Project) GetPkg (bpkg *build.Package) *Pkg {
	name := bpkg.Dir
	pkg := p.Pkgs [name]
	if pkg == nil {
		pkg = NewPkg (p, bpkg)
		for _, f := range bpkg.GoFiles {
			pkg.Srcs [f] = SrcNew(pkg, bpkg.Dir, f)
		}
		p.Pkgs [name] = pkg
	}
	return pkg
}

func  (p *Project) GetDirs () {
	dirs := make ([]string, 0, 100)
	for _, dir := range p.Context.SrcDirs () {
		p.Server.Writer.Log(dir)
		dirs = p.GetSubDirs(dirs, dir)
	}
	p.Dirs = dirs
}

func  (p *Project) GetPackages () {
	p.Pkgs = make (map[string]*Pkg)
	for _, dir := range p.Dirs {
		pkg, err := p.Context.ImportDir (dir, 0)
		if err == nil {
			p.GetPkg (pkg)
		}
	}
}

//-------------------------------------------------------------------

func (p *Project) Load () {
	p.GetDirs ()
	p.GetPackages()
	/*for _, pkg := range p.Pkgs {
		for n, f := range pkg.Srcs {
			/.MsgF ("Src [%s]: Dir=%s", n, f.Dir)
			//f.UpdateAst()
			//printer.Fprint (p.Server.Writer.Writer, f.FSet, f.Ast)
		}
	}*/
}

//-------------------------------------------------------------------

func (p *Project) GetSrc (fname string) (*Src, error) {
	dir, name := filepath.Split (fname)
	dir = path.Clean(dir)
	if pkg := p.Pkgs [dir]; pkg != nil {
		if src := pkg.Srcs[name]; src != nil {
			return src, nil
		} else {
			return nil, fmt.Errorf("Src [%s] not found in %s\n", name, dir)
		}
	} else {
		return nil, fmt.Errorf("Package [%s] not found\n", dir)
	}
}

//-------------------------------------------------------------------

type CompleteProc struct {
	Results *results.Completion
	analyzer.Processor
	Pos token.Pos
}
func (p *CompleteProc) Before (n ast.Node) bool {
	if n.Pos () <= p.Pos && p.Pos <= n.End () {
		switch n := n.(type) {
		case *ast.Ident:
			pref := n.Name[:p.Pos-n.Pos()]
                        //fmt.Printf ("Compl: [%s] at %d:%d, Req: %d, Pref=[%s]\n", n.Name, n.Pos (), n.End (), p.Pos, pref)
			res := make([]results.Choice, 0, 100)
			p.CurrLcl.Scan(func (d *env.Decl) bool {
				nm := d.Name.Name
				if strings.HasPrefix (nm, pref) {
					res = append (res, results.Choice{"VAR", nm, nm, n.Pos(), token.Pos (int(n.Pos ()) + len(nm))})
				}
				return true
			})
			p.Results = &results.Completion{pref, n.Name, n.Pos(), n.End(), res}
		}
	}
	return true
}

func (p *Project) Complete (no int, fname string, pos int) {
	if src, err := p.GetSrc(fname); src != nil {
		//fmt.Fprintf (p.Server.Writer.Writer, "Src [%s]: Dir=%s analyzed\n", name, dir)
		src.UpdateAst()
		//printer.Fprint (p.Server.Writer.Writer, p.FSet, src.Ast)
		a := analyzer.New(analyzer.NewKer (p.ModeTab), p.FSet, p.Server.Writer, false)
		completeProcessor := CompleteProc{nil, analyzer.Processor{a}, token.Pos(pos)}
		a.NodeProc = &completeProcessor
		a.Analyze(src.File, src.Ast)
		if completeProcessor.Results != nil {
		    completeProcessor.Results.Write(p.Writer)
		}
	} else {
		p.Msg(err.Error())
	}
}

//-------------------------------------------------------------------

func (p *Project) Analyze (no int, fName string) {
	if src, err := p.GetSrc(fName); src != nil {
		//fmt.Fprintf (p.Server.Writer.Writer, "Src [%s]: Dir=%s analyzed\n", name, dir)
		src.Pkg.UpdateAsts ()
		//src.UpdateAst()

		if false {
			fmt.Printf ("imports of package [%s]\n", src.Pkg.Name);
			for _, name := range  src.Pkg.Pkg.Imports {
			    fmt.Printf ("  Import [%s]\n", name);
			}
			fmt.Printf ("Files of package [%s]\n", src.Pkg.Name);
			for _, name := range  src.Pkg.Pkg.GoFiles {
			    fmt.Printf ("  File [%s]\n", name);
			}
		}

		//printer.Fprint (p.Server.Writer.Writer, p.FSet, src.Ast)
		ker := analyzer.NewKer (p.ModeTab)
		a := analyzer.New(ker, p.FSet, p.Server.Writer, true)
		fName := src.File.Name()
		//a.Analyze(src.Ast)
		a.SetTokenFile (src.File)
		a.SetOuterErrors (src.OuterErrors)
		a.AnalyzeFileIntr(src.Ast)
		a.AnalyzeFileBody(src.Ast)
		//a.Curr.Print()
		p.Server.Writer.Beg("ERRORS-CLEAR").Eol().End("ERRORS-CLEAR")
		a.GetErrors (fName, no).Write(p.Server.Writer)
		if (options.Codeblocks) {
			a.GetFonts (fName, fName, no).Write(p.Server.Writer)
		}
	} else {
		p.Msg(err.Error())
	}
}

//-------------------------------------------------------------------

func (p *Project) FindFiles (no int, pfx string, system bool, max int) {
	files := results.Files{system, make([]string, 0, 1000)}
	for _, pkg := range p.Pkgs {
		for _, f := range pkg.Srcs {
			if strings.HasPrefix(f.Name, pfx) {
				files.Files = append (files.Files, f.FName())
			}
		}
	}
	if len(files.Files) > max {
		files.Files = files.Files [:max]
	}
	files.Write(p.Writer)
}
