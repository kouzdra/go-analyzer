package env

import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/golang/stdnames"

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
	bldr.Declare(KType, stdnames.N_int  , paths.Root, nil, tab.Int  , nil)
	bldr.Declare(KType, stdnames.N_int8 , paths.Root, nil, tab.Int8 , nil)
	bldr.Declare(KType, stdnames.N_int16, paths.Root, nil, tab.Int16, nil)
	bldr.Declare(KType, stdnames.N_int32, paths.Root, nil, tab.Int32, nil)
	bldr.Declare(KType, stdnames.N_int64, paths.Root, nil, tab.Int64, nil)

	bldr.Declare(KType, stdnames.N_uint  , paths.Root, nil, tab.UInt  , nil)
	bldr.Declare(KType, stdnames.N_uint8 , paths.Root, nil, tab.UInt8 , nil)
	bldr.Declare(KType, stdnames.N_uint16, paths.Root, nil, tab.UInt16, nil)
	bldr.Declare(KType, stdnames.N_uint32, paths.Root, nil, tab.UInt32, nil)
	bldr.Declare(KType, stdnames.N_uint64, paths.Root, nil, tab.UInt64, nil)

	bldr.Declare(KType, stdnames.N_float32   , paths.Root, nil, tab.Float32   , nil)
	bldr.Declare(KType, stdnames.N_float64   , paths.Root, nil, tab.Float64   , nil)
	bldr.Declare(KType, stdnames.N_complex64 , paths.Root, nil, tab.Complex64 , nil)
	bldr.Declare(KType, stdnames.N_complex128, paths.Root, nil, tab.Complex128, nil)

	bldr.Declare(KType, stdnames.N_uintptr, paths.Root, nil, tab.RUint, nil)
	bldr.Declare(KType, stdnames.N_bool   , paths.Root, nil, tab.Bool, nil)
	bldr.Declare(KType, stdnames.N_byte   , paths.Root, nil, tab.Any, nil)
	bldr.Declare(KType, stdnames.N_rune   , paths.Root, nil, tab.Any, nil)
	bldr.Declare(KType, stdnames.N_string , paths.Root, nil, tab.String, nil)

	bldr.Declare(KConst, stdnames.N_nil  , paths.Root, nil, tab. Any, nil)
	bldr.Declare(KConst, stdnames.N_iota , paths.Root, nil, tab. Int, nil)
	bldr.Declare(KConst, stdnames.N_true , paths.Root, nil, tab.Bool, nil)
	bldr.Declare(KConst, stdnames.N_false, paths.Root, nil, tab.Bool, nil)

	bldr.Declare(KFunc, stdnames.N_append, paths.Root, nil, tab.Append , nil)
	bldr.Declare(KFunc, stdnames.N_make,   paths.Root, nil, tab.Make   , nil)
	bldr.Declare(KFunc, stdnames.N_new,    paths.Root, nil, tab.New    , nil)
	bldr.Declare(KFunc, stdnames.N_len,    paths.Root, nil, tab.Len    , nil)
	bldr.Declare(KFunc, stdnames.N_cap,    paths.Root, nil, tab.Cap    , nil)

	bldr.Declare(KFunc, stdnames.N_panic,   paths.Root, nil, tab.Panic  , nil)
	bldr.Declare(KFunc, stdnames.N_recover, paths.Root, nil, tab.Recover, nil)
	return bldr.Close ()
}

