package iface

import "go/token"
import "go/ast"
import "go/scanner"
import "go/build"
import "path/filepath"
import "github.com/kouzdra/go-analyzer/env"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/results"

type Project interface {
	//	GetOptions() Options
	GetContext () build.Context
	GetTree    () []Dir
	GetFileSet () *token.FileSet
	GetModeTab () *env.ModeTab

	GetDirs     () []*names.Name
	GetPackages () map [*names.Name]Package
	GetSrc (fname string) (Source, error)

	SetRoot (path string) 
	SetPath (path string) 
	GetRoot () *names.Name
	GetPath () *names.Name

	Load () 
	Complete  (src Source, pos int) *results.Completion
	Analyze   (src Source, no int) (*results.Errors, *results.Fontify)
	FindFiles (no int, pfx string, system bool, max int) *results.Files

	Msg  (msg string)
	MsgF (msg string, args... interface{})
}

type Dir interface  {
	GetPath () *names.Name
	GetSub  () [] Dir     
}



/*
type Program interface {
	GetName() *names.Name
	GetProject() Project
	//Packages() []Package

	//Update()
	//UpdatePkg(pkg Package)
	//UpdateSrc(src Source)

	GetPackage(path *names.Name) Package // nil - no package
}*/

type Package interface {
	GetProject () Project
	GetDir     () *names.Name
	GetName    () *names.Name
	GetSrcs    () map [*names.Name] Source
	GetPackage () *build.Package
	GetEnvLcl  () *env.Env
	GetEnvGbl  () *env.Env

	Reload ()
	UpdateAsts ()
}

type Source interface {
	GetPackage () Package
	GetDir     () *names.Name
	GetName    () *names.Name
	GetAst     () *  ast.File
	GetFile    () *token.File
	
	GetOuterErrors () scanner.ErrorList
	GetInnerErrors () []results.Error

	GetText() string
	SetText(string)
	Reload ()
	Changed (int, int, string)
	ReParse ()  (*token.File, *ast.File, scanner.ErrorList)
	UpdateAst()
}

func FName (src Source) string {
	return filepath.Join(src.GetDir().Name, src.GetName().Name)
}

 
