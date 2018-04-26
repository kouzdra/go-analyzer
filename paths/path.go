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

func Put (
