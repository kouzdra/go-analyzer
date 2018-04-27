package sym

//import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/project"

type T struct {
	Path *paths.Path
	Defs []project.Loc
	Refs []project.Loc
}

type Builder struct {
	Path *paths.Path
	Defs []project.Loc
	Refs []project.Loc
}

func NewBuilder(path *paths.Path) Builder {
	return Builder{path, make([]project.Loc, 0, 2), make([]project.Loc, 0, 128)}
}

func (bldr *Builder) Close() T {
	return T{bldr.Path, bldr.Defs[0:], bldr.Refs[0:]}
}
