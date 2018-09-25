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

type Prj struct {
	Context build.Context
	dirs [] *names.Name
	tree [] Dir
	packages map [*names.Name]Package
	fileSet *token.FileSet
	modeTab *env.ModeTab
}

type Dir struct {
	path *names.Name
	Sub [] Dir
}

//--------------------------------------------------

func (p *Prj) GetTree () []Dir { return p.tree }
func (p *Prj) GetFileSet () *token.FileSet { return p.fileSet }
func (p *Prj) GetModeTab () *env.ModeTab { return p.modeTab }

func (pr *Prj) GetDirs     () []*names.Name { return pr.dirs; }
func (pr *Prj) GetPackages () map [*names.Name]Package{ return pr.packages; }

func (dir *Dir) GetPath () *names.Name { return dir.path }

//--------------------------------------------------

func NewProject() *Prj {
	return &Prj{Context:build.Default, fileSet:token.NewFileSet(), modeTab:env.NewModeTab ()}
}

//---------------------------------------------------

func (p *Prj) SetRoot (path string) {
	p.MsgF ("ROOT=%s", path)
	p.Context.GOROOT = path
}

func (p *Prj) SetPath (path string) {
	p.MsgF ("PATH=%s", path)
	p.Context.GOPATH = path
}

func (p *Prj) N_GOROOT () *names.Name {
	return names.Put(p.Context.GOROOT)
}

func (p *Prj) N_GOPATH () *names.Name {
	return names.Put(p.Context.GOPATH)
}

func (s *Prj) Msg (msg string) {
	log.Printf ("PRJ: %s", msg)
	//results.Message{msg}.Write(s.Writer)
}

func (s *Prj) MsgF (f string, args... interface{}) {
	log.Printf ("PRJ: %s", fmt.Sprintf (f, args...))
	//results.Message{fmt.Sprintf (f, args...)}.Write(s.Writer)
}

//-------------------------------------------------------------------

func (p *Prj) MakeSubDirs (dest []*names.Name, rootPath *names.Name, rootName *names.Name, tree [] Dir) ([]*names.Name, [] Dir) {
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

func (p *Prj) MakePackage (bpkg *build.Package) Package {
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

func  (p *Prj) MakeDirs () {
	dirs := make ([]*names.Name, 0, 100)
	tree := make ([]Dir, 0)
	for _, dir := range p.Context.SrcDirs () {
		//p.Server.Writer.Log(dir)
		dirN := names.Put (dir)
		dirs, tree = p.MakeSubDirs(dirs, dirN, dirN, tree)
	}
	p.dirs = dirs
	p.tree = tree
}

func  (p *Prj) MakePackages () {
	p.packages = make (map[*names.Name]Package)
	for _, dir := range p.GetDirs() {
		pkg, err := p.Context.ImportDir (dir.Name, 0)
		if err == nil {
			p.MakePackage (pkg)
		}
	}
}

//-------------------------------------------------------------------

func (p *Prj) Load () {
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

func (p *Prj) GetSrc (fname string) (Source, error) {
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

func (p *Prj) Complete (src Source, pos int) *results.Completion {
	src.UpdateAst()
	a := analyzer.New(analyzer.NewKer (p.GetModeTab ()), p.GetFileSet(), false)
	completeProcessor := CompleteProc{nil, analyzer.Processor{a}, token.Pos(src.GetFile().Base () + pos)}
	a.NodeProc = &completeProcessor
	a.Analyze(src.GetFile(), src.GetAst())
	return completeProcessor.Results
}

//-------------------------------------------------------------------

func (p *Prj) Analyze (src Source, no int) (*results.Errors, *results.Fontify) {
	pkg := src.GetPackage()
	pkg.UpdateAsts ()
	//src.UpdateAst()
	
	if false {
		fmt.Printf ("imports of package [%s]\n", pkg.GetName ().Name);
		for _, name := range  pkg.GetPackage ().Imports {
			fmt.Printf ("  Import [%s]\n", name);
		}
		fmt.Printf ("Files of package [%s]\n", pkg.GetName ().Name);
		for _, name := range  pkg.GetPackage ().GoFiles {
			fmt.Printf ("  File [%s]\n", name);
		}
	}
	
	ker := analyzer.NewKer (p.GetModeTab())
	a := analyzer.New(ker, p.GetFileSet(), true)
	fName := src.GetFile().Name()
	//a.Analyze(src.Ast)
	a.SetTokenFile (src.GetFile())
	a.SetOuterErrors (src.GetOuterErrors())
	a.AnalyzeFileIntr(src.GetAst())
	a.AnalyzeFileBody(src.GetAst())
	//a.Curr.Print()
	return a.GetErrors (fName, no), a.GetFonts (fName, fName, no)
}

//-------------------------------------------------------------------

func (p *Prj) FindFiles (no int, pfx string, system bool, max int) *results.Files {
	files := results.Files{system, make([]string, 0, 1000)}
	for _, pkg := range p.GetPackages () {
		for _, f := range pkg.GetSrcs () {
			if strings.HasPrefix(f.GetName ().Name, pfx) {
				files.Files = append (files.Files, FName (f))
			}
		}
	}
	if len(files.Files) > max {
		files.Files = files.Files [:max]
	}
	return &files
}
