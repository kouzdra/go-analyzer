package project

import "fmt"
import "log"
import "strings"
import "io/ioutil"
import "path"
import "path/filepath"
import "go/build"
import "go/token"
import "go/ast"
import "github.com/kouzdra/go-analyzer/env"
import "github.com/kouzdra/go-analyzer/analyzer"
import "github.com/kouzdra/go-analyzer/results"
//import "github.com/kouzdra/go-analyzer/options"

type Project struct {
	Context build.Context
	Dirs [] string
	Tree [] Dir
	Pkgs map [string]*Pkg
	FSet *token.FileSet
	ModeTab *env.ModeTab
}

type Dir struct {
	Path string
	Sub [] Dir
}

func NewProject() *Project {
	return &Project{Context:build.Default, FSet:token.NewFileSet(), ModeTab:env.NewModeTab ()}
}

func (p *Project) SetRoot (path string) {
	p.MsgF ("ROOT=%s", path)
	p.Context.GOROOT = path
}

func (p *Project) SetPath (path string) {
	p.MsgF ("PATH=%s", path)
	p.Context.GOPATH = path
}

func (s *Project) Msg (msg string) {
	log.Printf ("PRJ: %s", msg)
	//results.Message{msg}.Write(s.Writer)
}

func (s *Project) MsgF (f string, args... interface{}) {
	log.Printf ("PRJ: %s", fmt.Sprintf (f, args...))
	//results.Message{fmt.Sprintf (f, args...)}.Write(s.Writer)
}

//-------------------------------------------------------------------

func (p *Project) GetSubDirs (dest [] string, rootPath string, rootName string, tree [] Dir) ([] string, [] Dir) {
	dest = append (dest, rootPath)
	//fmt.Printf ("File: [%s]\n", root)
	files, _ := ioutil.ReadDir (rootPath)
	subTree := make ([]Dir, 0)
	for _, file := range files {
		if file.IsDir () {
			dest, subTree = p.GetSubDirs (dest, filepath.Join (rootPath, file.Name()), file.Name(), subTree)
		}
	}
	return dest, append (tree, Dir{rootPath, subTree})
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
	tree := make ([]Dir, 0)
	for _, dir := range p.Context.SrcDirs () {
		//p.Server.Writer.Log(dir)
		dirs, tree = p.GetSubDirs(dirs, dir, dir, tree)
	}
	p.Dirs = dirs
	p.Tree = tree
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

func (p *Project) Complete (src *Src, no int, pos int) *results.Completion {
	src.UpdateAst()
	a := analyzer.New(analyzer.NewKer (p.ModeTab), p.FSet, false)
	completeProcessor := CompleteProc{nil, analyzer.Processor{a}, token.Pos(pos)}
	a.NodeProc = &completeProcessor
	a.Analyze(src.File, src.Ast)
	return completeProcessor.Results
}

//-------------------------------------------------------------------

func (p *Project) Analyze (src *Src, no int) (*results.Errors, *results.Fontify) {
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
	
	ker := analyzer.NewKer (p.ModeTab)
	a := analyzer.New(ker, p.FSet, true)
	fName := src.File.Name()
	//a.Analyze(src.Ast)
	a.SetTokenFile (src.File)
	a.SetOuterErrors (src.OuterErrors)
	a.AnalyzeFileIntr(src.Ast)
	a.AnalyzeFileBody(src.Ast)
	//a.Curr.Print()
	return a.GetErrors (fName, no), a.GetFonts (fName, fName, no)
}

//-------------------------------------------------------------------

func (p *Project) FindFiles (no int, pfx string, system bool, max int) *results.Files {
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
	return &files
}
