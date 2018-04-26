package paths

import "github.com/kouzdra/go-analyzer/names"

type Path struct {
	Name *names.Name
}

const pathSize = 1024

type pathElem struct {
	Path
	No int
	next *pathElem
}

var cnt = 0

var pathTab [pathSize]pathElem

//func put (
