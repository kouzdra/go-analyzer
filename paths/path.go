package paths

import "github.com/kouzdra/go-analyzer/names"

type Path struct {
	No uint
	Name *names.Name
}

const pathSize = 1024

type pathElem struct {
	Path
	next *pathElem
}

var cnt = 0

var pathTab [pathSize]pathElem

func Make (base *Path, name *names.Name) *Path {
	return nil // TODO
}

func Put (names ...*names.Name) *Path {
	var path *Path = nil
	for _, name := range names {
		path = Make(path, name)
	}
	return path
}
