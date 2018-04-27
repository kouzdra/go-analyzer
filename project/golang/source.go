package golang

//import "github.com/kouzdra/go-analyzer/project/iface"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"

type Source struct {
	Name *names.Name
	Path *paths.Path
	super *Package
	//Package() Package
}

func (s *Source) Package() *Package {
	return s.super
}

