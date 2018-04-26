package env

import "github.com/kouzdra/go-analyzer/names"

type list struct {
        mode Mode
        next *list
}

type ModeTab struct {
        BModes
        BEnv *Env
        modes *list
}

func NewModeTab () *ModeTab {
        var res ModeTab
        res.NewBModes ()
        res.BEnv = res.DeclareBuiltins ()
        return &res
}

func (tab *ModeTab) Find (m Mode) Mode {
        for e := tab.modes; e != nil; e = e.next {
                if m.Equals (e.mode) {
                        return e.mode
                }
        }
        tab.modes = &list{m, tab.modes}
        return m
}

func (tab *ModeTab) Builtin (m Builtin) *Builtin { return tab.Find (&m).(*Builtin) }
func (tab *ModeTab) Ref (m Mode) *Ref { return tab.Find (&Ref{m}).(*Ref) }
func (tab *ModeTab) IdType (d *Decl, m Mode) *IdType { return tab.Find (&IdType{d, m, NewBldr()}).(*IdType) }

func (tab *ModeTab) Struct (e *Env) *Struct { return tab.Find (&Struct{e}).(*Struct) }
func (tab *ModeTab) Func (args []Mode, result []Mode) *Func { return tab.Find (&Func{args, result}).(*Func) }
func (tab *ModeTab) Array (size *int, elem Mode) *Array { return tab.Find (&Array{size, elem}).(*Array) }

//===============================================================

func (tab *ModeTab) DeclareBuiltins () *Env {
        bldr := NewBldr ()
	bldr.Declare(KType, names.Put("int"   ), nil, tab.Int  , nil)
	bldr.Declare(KType, names.Put("int8"  ), nil, tab.Int8 , nil)
	bldr.Declare(KType, names.Put("int16" ), nil, tab.Int16, nil)
	bldr.Declare(KType, names.Put("int32" ), nil, tab.Int32, nil)
	bldr.Declare(KType, names.Put("int64" ), nil, tab.Int64, nil)

	bldr.Declare(KType, names.Put("uint"  ), nil, tab.UInt  , nil)
	bldr.Declare(KType, names.Put("uint8" ), nil, tab.UInt8 , nil)
	bldr.Declare(KType, names.Put("uint16"), nil, tab.UInt16, nil)
	bldr.Declare(KType, names.Put("uint32"), nil, tab.UInt32, nil)
	bldr.Declare(KType, names.Put("uint64"), nil, tab.UInt64, nil)

	bldr.Declare(KType, names.Put("float32"   ), nil, tab.Float32   , nil)
	bldr.Declare(KType, names.Put("float64"   ), nil, tab.Float64   , nil)
	bldr.Declare(KType, names.Put("complex64" ), nil, tab.Complex64 , nil)
	bldr.Declare(KType, names.Put("complex128"), nil, tab.Complex128, nil)

	bldr.Declare(KType, names.Put("uintptr"), nil, tab.RUint, nil)
	bldr.Declare(KType, names.Put("bool"   ), nil, tab.Bool, nil)
	bldr.Declare(KType, names.Put("byte"   ), nil, tab.Any, nil)
	bldr.Declare(KType, names.Put("rune"   ), nil, tab.Any, nil)
	bldr.Declare(KType, names.Put("string" ), nil, tab.String, nil)

	bldr.Declare(KConst, names.Put("nil"   ), nil, tab. Any, nil)
	bldr.Declare(KConst, names.Put("iota"  ), nil, tab. Int, nil)
	bldr.Declare(KConst, names.Put("true"  ), nil, tab.Bool, nil)
	bldr.Declare(KConst, names.Put("false" ), nil, tab.Bool, nil)

	bldr.Declare(KFunc, names.Put("append" ), nil, tab.Append , nil)
	bldr.Declare(KFunc, names.Put("make"   ), nil, tab.Make   , nil)
	bldr.Declare(KFunc, names.Put("new"    ), nil, tab.New    , nil)
	bldr.Declare(KFunc, names.Put("len"    ), nil, tab.Len    , nil)
	bldr.Declare(KFunc, names.Put("cap"    ), nil, tab.Cap    , nil)

	bldr.Declare(KFunc, names.Put("panic"  ), nil, tab.Panic  , nil)
	bldr.Declare(KFunc, names.Put("recover"), nil, tab.Recover, nil)
	return bldr.Close ()
}

