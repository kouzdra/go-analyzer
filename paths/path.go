package paths

import "path/filepath"
import "github.com/kouzdra/go-analyzer/names"

type Path struct {
	No    uint
	_hash uint
	Base  *Path
	Name  *names.Name
}

const hashSize = 1024

type pathElem struct {
	Path
	next *pathElem
}

var cnt uint = 0

var pathTab [hashSize]*pathElem

var Root = Put(names.Root)

func (path *Path) Repr() string {
	if path == nil {
		panic("paths.Repr: nil path")
	}
	return path.Base.Repr() + "." + path.Name.Repr()
}

func (base *Path) hashName(name *names.Name) uint {
	return base.hash() + name.Hash
}

func (base *Path) hash() uint {
	if base == nil {
		return 0
	} else {
		return base._hash
	}

}

func (base *Path) Find(name *names.Name) *Path {
	hash := base.hashName(name)
	for elem := pathTab[hash%hashSize]; elem != nil; elem = elem.next {
		if elem.Base == base && elem.Name == name {
			return &elem.Path
		}
	}
	return nil
}

func (base *Path) Make(name *names.Name) *Path {
	if path := base.Find(name); path != nil {
		return path
	}
	hash := base.hashName(name)
	elem := pathElem{Path{cnt, hash, base, name}, pathTab[hash%hashSize]}
	cnt++
	return &elem.Path
}

func Put(names ...*names.Name) *Path {
	var path *Path = nil
	for _, name := range names {
		path = path.Make(name)
	}
	return path
}

func PutStr(path string) *Path {
	var res *Path = nil
	for _, name := range filepath.SplitList (path) {
		res = res.Make (names.Put (name))
	}
	return res
}

