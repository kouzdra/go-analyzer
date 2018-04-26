package paths

import "github.com/kouzdra/go-analyzer/names"

type Path struct {
	No uint
	_hash uint
	Base *Path
	Name *names.Name
}

const pathSize = 1024

type pathElem struct {
	Path
	next *pathElem
}

var cnt = 0

var pathTab [pathSize]pathElem

func (base *Path) hashName (name *names.Name) uint {
	return base.hash () + name.Hash
}

func (base *Path) hash () uint {
	if base == nil {
		return 0
	} else {
		return base._hash
	}
		
}

func (base *Path) Make (name *names.Name) *Path {
	return nil // TODO
}

func Put (names ...*names.Name) *Path {
	var path *Path = nil
	for _, name := range names {
		path = path.Make(name)
	}
	return path
}
