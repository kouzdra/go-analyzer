package golang

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
import "github.com/kouzdra/go-analyzer/golang/analyzer"
import "github.com/kouzdra/go-analyzer/results"
//import "github.com/kouzdra/go-analyzer/options"
import "github.com/kouzdra/go-analyzer/iface/iproject"

type prj struct {
	context build.Context
	dirs [] *names.Name
	tree [] iproject.IDir
	packages map [*names.Name] iproject.IPackage
	fileSet *token.FileSet
	modeTab *env.ModeTab
}

//--------------------------------------------------

func (p *prj) GetContext () build.Context { return p.context }
func (p *prj) GetTree    () []iproject.IDir { return p.tree }
func (p *prj) GetFileSet () *token.FileSet { return p.fileSet }
func (p *prj) GetModeTab () *env.ModeTab { return p.modeTab }

func (pr *prj) GetDirs     () []*names.Name { return pr.dirs; }
func (pr *prj) GetPackages () map [*names.Name] iproject.IPackage{ return pr.packages; }

//--------------------------------------------------

type dir struct {
	path *names.Name
	sub [] iproject.IDir
}

func (d *dir) GetPath () *names.Name { return d.path }
func (d *dir) GetSub  () [] iproject.IDir      { return d.sub  }

//--------------------------------------------------

func NewProject() iproject.IProject {
	return &prj{context:build.Default, fileSet:token.NewFileSet(), modeTab:env.NewModeTab ()}
}

//---------------------------------------------------

func (p *prj) SetRoot (path string) {
	p.MsgF ("ROOT=%s", path)
	p.context.GOROOT = path
}
///
func (p *prj) SetPath (path string) {
	p.MsgF ("PATH=%s", path)
	p.context.GOPATH = path
}

func (p *prj) GetRoot () *names.Name {
	return names.Put(p.context.GOROOT)
}

func (p *prj) GetPath () *names.Name {
	return names.Put(p.context.GOPATH)
}

func (s *prj) Msg (msg string) {
	log.Printf ("PRJ: %s", msg)
	//results.Message{msg}.Write(s.Writer)
}

func (s *prj) MsgF (f string, args... interface{}) {
	log.Printf ("PRJ: %s", fmt.Sprintf (f, args...))
	//results.Message{fmt.Sprintf (f, args...)}.Write(s.Writer)
}

//-------------------------------------------------------------------

func (p *prj) makeSubDirs (dest []*names.Name, rootPath *names.Name,
	rootName *names.Name, tree [] iproject.IDir) ([]*names.Name, [] iproject.IDir) {
	dest = append (dest, rootPath)
	//fmt.Printf ("File: [%s]\n", root)
	files, _ := ioutil.ReadDir (rootPath.Name)
	subTree := make ([] iproject.IDir, 0)
	for _, file := range files {
		if file.IsDir () {
			dest, subTree = p.makeSubDirs (dest, names.Put (filepath.Join (rootPath.Name, file.Name())), names.Put (file.Name()), subTree)
		}
	}
	return dest, append (tree, &dir{rootPath, subTree})
}

func (p *prj) makePackage (bpkg *build.Package) iproject.IPackage {
	name := names.Put (bpkg.Dir)
	pkg := p.GetPackages() [name]
	if pkg == nil {
		pkg = newPkg (p, bpkg)
		for _, f := range bpkg.GoFiles {
			ff := names.Put (f)
			pkg.GetSrcs()[ff] = srcNew(pkg, name, ff)
		}
		p.GetPackages () [name] = pkg
	}
	return pkg
}

func  (p *prj) makeDirs () {
	dirs := make ([]*names.Name, 0, 100)
	tree := make ([]iproject.IDir, 0)
	for _, dir := range p.context.SrcDirs () {
		//p.Server.Writer.Log(dir)
		dirN := names.Put (dir)
		dirs, tree = p.makeSubDirs(dirs, dirN, dirN, tree)
	}
	p.dirs = dirs
	p.tree = tree
}

func  (p *prj) makePackages () {
	p.packages = make (map[*names.Name]iproject.IPackage)
	for _, dir := range p.GetDirs() {
		pkg, err := p.context.ImportDir (dir.Name, 0)
		if err == nil {
			p.makePackage (pkg)
		}
	}
}

//-------------------------------------------------------------------

func (p *prj) Load () {
	p.makeDirs ()
	p.makePackages()
	/*for _, pkg := range p.Pkgs {
		for n, f := range pkg.Srcs {
			/.MsgF ("Src [%s]: Dir=%s", n, f.Dir)
			//f.UpdateAst()
			//printer.Fprint (p.Server.Writer.Writer, f.FSet, f.Ast)
		}
	}*/
}

//-------------------------------------------------------------------

func (p *prj) GetSrc (fname string) (iproject.ISource, error) {
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

func (p *prj) Complete (src iproject.ISource, pos int) *results.Completion {
	src.UpdateAst()
	a := analyzer.New(p, false)
	completeProcessor := CompleteProc{nil, analyzer.Processor{a}, token.Pos(src.GetFile().Base () + pos)}
	a.NodeProc = &completeProcessor
	a.Analyze(src.GetFile(), src.GetAst())
	return completeProcessor.Results
}

//-------------------------------------------------------------------

func (p *prj) Analyze (src iproject.ISource, no int) (*results.Errors, *results.Fontify) {
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
	
	a := analyzer.New(p, true)
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

func (p *prj) FindFiles (no int, pfx string, system bool, max int) *results.Files {
	files := results.Files{system, make([]string, 0, 1000)}
	for _, pkg := range p.GetPackages () {
		for _, f := range pkg.GetSrcs () {
			if strings.HasPrefix(f.GetName ().Name, pfx) {
				files.Files = append (files.Files, iproject.FName (f))
			}
		}
	}
	if len(files.Files) > max {
		files.Files = files.Files [:max]
	}
	return &files
}
