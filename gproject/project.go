package gproject

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
import "github.com/kouzdra/go-analyzer/names"
//import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/analyzer"
import "github.com/kouzdra/go-analyzer/results"
//import "github.com/kouzdra/go-analyzer/options"

type Project struct {
	Context build.Context
	dirs [] *names.Name
	Tree [] Dir
	packages map [*names.Name]*Pkg
	FSet *token.FileSet
	ModeTab *env.ModeTab
}

type Dir struct {
	path *names.Name
	Sub [] Dir
}

func NewProject() *Project {
	return &Project{Context:build.Default, FSet:token.NewFileSet(), ModeTab:env.NewModeTab ()}
}

//---------------------------------------------------

func (pr *Project) GetDirs     () []*names.Name { return pr.dirs; }
func (pr *Project) GetPackages () map [*names.Name]*Pkg { return pr.packages; }

func (dir *Dir) GetPath () *names.Name { return dir.path }

//---------------------------------------------------

func (p *Project) SetRoot (path string) {
	p.MsgF ("ROOT=%s", path)
	p.Context.GOROOT = path
}

func (p *Project) SetPath (path string) {
	p.MsgF ("PATH=%s", path)
	p.Context.GOPATH = path
}

func (p *Project) N_GOROOT () *names.Name {
	return names.Put(p.Context.GOROOT)
}

func (p *Project) N_GOPATH () *names.Name {
	return names.Put(p.Context.GOPATH)
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

func (p *Project) MakeSubDirs (dest []*names.Name, rootPath *names.Name, rootName *names.Name, tree [] Dir) ([]*names.Name, [] Dir) {
	dest = append (dest, rootPath)
	//fmt.Printf ("File: [%s]\n", root)
	files, _ := ioutil.ReadDir (rootPath.Name)
	subTree := make ([]Dir, 0)
	for _, file := range files {
		if file.IsDir () {
			dest, subTree = p.MakeSubDirs (dest, names.Put (filepath.Join (rootPath.Name, file.Name())), names.Put (file.Name()), subTree)
		}
	}
	return dest, append (tree, Dir{rootPath, subTree})
}

func (p *Project) MakePackage (bpkg *build.Package) *Pkg {
	name := names.Put (bpkg.Dir)
	pkg := p.GetPackages() [name]
	if pkg == nil {
		pkg = NewPkg (p, bpkg)
		for _, f := range bpkg.GoFiles {
			ff := names.Put (f)
			pkg.GetSrcs () [ff] = SrcNew(pkg, name, ff)
		}
		p.GetPackages () [name] = pkg
	}
	return pkg
}

func  (p *Project) MakeDirs () {
	dirs := make ([]*names.Name, 0, 100)
	tree := make ([]Dir, 0)
	for _, dir := range p.Context.SrcDirs () {
		//p.Server.Writer.Log(dir)
		dirN := names.Put (dir)
		dirs, tree = p.MakeSubDirs(dirs, dirN, dirN, tree)
	}
	p.dirs = dirs
	p.Tree = tree
}

func  (p *Project) MakePackages () {
	p.packages = make (map[*names.Name]*Pkg)
	for _, dir := range p.GetDirs() {
		pkg, err := p.Context.ImportDir (dir.Name, 0)
		if err == nil {
			p.MakePackage (pkg)
		}
	}
}

//-------------------------------------------------------------------

func (p *Project) Load () {
	p.MakeDirs ()
	p.MakePackages()
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
	if pkg := p.GetPackages() [names.Put (dir)]; pkg != nil {
		if src := pkg.GetSrcs () [names.Put (name)]; src != nil {
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
	//log.Printf ("Node %d:%d %v\n", n.Pos(), n.End (), n)
	if n.Pos () <= p.Pos && p.Pos <= n.End () {
		switch n := n.(type) {
		case *ast.Ident:
			pref := n.Name[:p.Pos-n.Pos()]
                        //fmt.Printf ("Compl: [%s] at %d:%d, Req: %d, Pref=[%s]\n", n.Name, n.Pos (), n.End (), p.Pos, pref)
			res := make([]results.Choice, 0, 100)
			//loc := p.file.Position (n.Pos)
			p.CurrLcl.Scan(func (d *env.Decl) bool {
				nm := d.Name.Name
				if strings.HasPrefix (nm, pref) {
					k := "???"
					switch d.Kind {
					case env.KType:    k = "TYPE"
					case env.KFunc:    k = "FUNC"
					case env.KVar:     k = "VAR"
					case env.KPackage: k = "PACKAGE"
					case env.KConst:   k = "CONST"
					}
					res = append (res, results.Choice{k, nm, nm, n.Pos(), token.Pos (int(n.Pos ()) + len(nm))})
				}
				return true
			})
			p.Results = &results.Completion{pref, n.Name, n.Pos(), n.End(), res}
		}
	}
	return true
}

func (p *Project) Complete (src *Src, pos int) *results.Completion {
	src.UpdateAst()
	a := analyzer.New(analyzer.NewKer (p.ModeTab), p.FSet, false)
	completeProcessor := CompleteProc{nil, analyzer.Processor{a}, token.Pos(src.File.Base () + pos)}
	a.NodeProc = &completeProcessor
	a.Analyze(src.File, src.Ast)
	return completeProcessor.Results
}

//-------------------------------------------------------------------

func (p *Project) Analyze (src *Src, no int) (*results.Errors, *results.Fontify) {
	src.Pkg.UpdateAsts ()
	//src.UpdateAst()
	
	if false {
		fmt.Printf ("imports of package [%s]\n", src.Pkg.GetName ().Name);
		for _, name := range  src.Pkg.Pkg.Imports {
			fmt.Printf ("  Import [%s]\n", name);
		}
		fmt.Printf ("Files of package [%s]\n", src.Pkg.GetName ().Name);
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
	for _, pkg := range p.GetPackages () {
		for _, f := range pkg.GetSrcs () {
			if strings.HasPrefix(f.GetName ().Name, pfx) {
				files.Files = append (files.Files, f.FName())
			}
		}
	}
	if len(files.Files) > max {
		files.Files = files.Files [:max]
	}
	return &files
}
