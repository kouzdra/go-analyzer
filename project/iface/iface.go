package iface

import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"

type Project interface {
	Name() names.Name
	Options() Options
	Programs() []Program
	Packages() []Package
}

type Program interface {
}

type Package interface {
	Srcs() []Src
}

type Src interface {
	Name() names.Name
	Path() paths.Path
}
