package sym

//import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/project/iface"

type T struct {
	Path *paths.Path
	Defs []iface.Loc
	Refs []iface.Loc
}

type Builder struct {
	Path *paths.Path
	Defs []iface.Loc
	Refs []iface.Loc
}

func NewBuilder(path *paths.Path) Builder {
	return Builder{path, make([]iface.Loc, 0, 2), make([]iface.Loc, 0, 128)}
}

func (bldr *Builder) Close() T {
	return T{bldr.Path, bldr.Defs[0:], bldr.Refs[0:]}
}
