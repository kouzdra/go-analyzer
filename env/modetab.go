package env

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
	bldr.Declare(KType, Put("int"   ), nil, tab.Int  , nil)
	bldr.Declare(KType, Put("int8"  ), nil, tab.Int8 , nil)
	bldr.Declare(KType, Put("int16" ), nil, tab.Int16, nil)
	bldr.Declare(KType, Put("int32" ), nil, tab.Int32, nil)
	bldr.Declare(KType, Put("int64" ), nil, tab.Int64, nil)

	bldr.Declare(KType, Put("uint"  ), nil, tab.UInt  , nil)
	bldr.Declare(KType, Put("uint8" ), nil, tab.UInt8 , nil)
	bldr.Declare(KType, Put("uint16"), nil, tab.UInt16, nil)
	bldr.Declare(KType, Put("uint32"), nil, tab.UInt32, nil)
	bldr.Declare(KType, Put("uint64"), nil, tab.UInt64, nil)

	bldr.Declare(KType, Put("float32"   ), nil, tab.Float32   , nil)
	bldr.Declare(KType, Put("float64"   ), nil, tab.Float64   , nil)
	bldr.Declare(KType, Put("complex64" ), nil, tab.Complex64 , nil)
	bldr.Declare(KType, Put("complex128"), nil, tab.Complex128, nil)

	bldr.Declare(KType, Put("uintptr"), nil, tab.RUint, nil)
	bldr.Declare(KType, Put("bool"   ), nil, tab.Bool, nil)
	bldr.Declare(KType, Put("byte"   ), nil, tab.Any, nil)
	bldr.Declare(KType, Put("rune"   ), nil, tab.Any, nil)
	bldr.Declare(KType, Put("string" ), nil, tab.String, nil)

	bldr.Declare(KConst, Put("nil"   ), nil, tab. Any, nil)
	bldr.Declare(KConst, Put("iota"  ), nil, tab. Int, nil)
	bldr.Declare(KConst, Put("true"  ), nil, tab.Bool, nil)
	bldr.Declare(KConst, Put("false" ), nil, tab.Bool, nil)

	bldr.Declare(KFunc, Put("append" ), nil, tab.Append , nil)
	bldr.Declare(KFunc, Put("make"   ), nil, tab.Make   , nil)
	bldr.Declare(KFunc, Put("new"    ), nil, tab.New    , nil)
	bldr.Declare(KFunc, Put("len"    ), nil, tab.Len    , nil)
	bldr.Declare(KFunc, Put("cap"    ), nil, tab.Cap    , nil)

	bldr.Declare(KFunc, Put("panic"  ), nil, tab.Panic  , nil)
	bldr.Declare(KFunc, Put("recover"), nil, tab.Recover, nil)
	return bldr.Close ()
}

