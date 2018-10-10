package iproject

import "path/filepath"
import "github.com/kouzdra/go-analyzer/defs"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/results"

type IProject interface {
	GetTree    () []IDir

	GetDirs     () []*names.Name
	GetPackages () map [*names.Name]IPackage
	GetSrc (fname string) (ISource, error)

	SetRoot (path string) 
	SetPath (path string) 
	GetRoot () *names.Name
	GetPath () *names.Name

	Load () 
	Complete  (src ISource, pos defs.Pos) *results.Completion
	Analyze   (src ISource, no int) (*results.Errors, *results.Fontify)
	FindFiles (no int, pfx string, system bool, max int) *results.Files

	Msg  (msg string)
	MsgF (msg string, args... interface{})
}

type IDir interface  {
	GetPath () *names.Name
	GetSub  () [] IDir     
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

type IPackage interface {
	GetProject () IProject
	GetDir     () *names.Name
	GetName    () *names.Name
	GetSrcs    () map [*names.Name] ISource

	Reload ()
}

type ISource interface {
	GetPackage () IPackage
	GetDir     () *names.Name
	GetName    () *names.Name
	GetSize    () int
	
	GetText () string
	SetText (string)
	Reload  ()
	Changed (int, int, string)
}

func FName (src ISource) string {
	return filepath.Join(src.GetDir().Name, src.GetName().Name)
}

 
