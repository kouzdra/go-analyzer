package sym

//import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/gproject"

type T struct {
	Path *paths.Path
	Defs []gproject.Loc
	Refs []gproject.Loc
}

type Builder struct {
	Path *paths.Path
	Defs []gproject.Loc
	Refs []gproject.Loc
}

func NewBuilder(path *paths.Path) Builder {
	return Builder{path, make([]gproject.Loc, 0, 2), make([]gproject.Loc, 0, 128)}
}

func (bldr *Builder) Close() T {
	return T{bldr.Path, bldr.Defs[0:], bldr.Refs[0:]}
}
