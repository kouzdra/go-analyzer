package env

type Path struct {
	Name *Name
}

const pathSize = 1024

type pathElem struct {
	Path
	next *pathElem
}

var pathTab [pathSize]pathElem
