package env

import "fmt"

type Mode interface {
        Head () string;
        Repr () string;
        Hash ()   uint;
        Equals (b Mode) bool;
}

var ModeNil = Mode(nil)

func ModeRepr (m Mode) string {
        if m == nil { return "<nil>" } else { return m.Repr () }
}

type Builtin int

func MAny (m Mode) bool {
        switch m := m.(type) {
        case *Builtin: return *m == KAny || *m == KErr
        default: return false
        }
}

func DerefOpt (m Mode) Mode {
        switch m := m.(type) {
        case *Ref: return m.Mode
        default: return m
        }
}

func DeId (m Mode) Mode {
        switch m := m.(type) {
        case *IdType: return m.Base
        default: return m
        }
}

//--------------------------------------------------------------------------------------

const (
        KAny = Builtin (iota)
        KErr

        KInt
        KInt8
        KInt16
        KInt32
        KInt64

        KUInt
        KUInt8
        KUInt16
        KUInt32
        KUInt64

        KFloat32
        KFloat64
        KComplex64
        KComplex128

        KBool
        KString

        KAppend
        KMake
        KLen
        KNew
        KCap

        KPanic
        KRecover
)

type BModes struct {
        Err    *Builtin
        Any    *Builtin
        Int    *Builtin
        Int8   *Builtin
        Int16  *Builtin
        Int32  *Builtin
        Int64  *Builtin

        UInt   *Builtin
        UInt8  *Builtin
        UInt16 *Builtin
        UInt32 *Builtin
        UInt64 *Builtin

        Bool   *Builtin
        String *Builtin
        RUint  *Ref

        Float32 *Builtin
        Float64 *Builtin
        Complex64  *Builtin
        Complex128 *Builtin

        Append     *Builtin
        Make       *Builtin
        Len        *Builtin
        New        *Builtin
        Cap        *Builtin

        Panic      *Builtin
        Recover    *Builtin
}

func (t *ModeTab) NewBModes ()  {
        t.Err        =  t.Builtin (KErr)
        t.Any        =  t.Builtin (KAny)

        t.Int        =  t.Builtin (KInt   )
        t.Int8       =  t.Builtin (KInt8  )
        t.Int16      =  t.Builtin (KInt16 )
        t.Int32      =  t.Builtin (KInt32 )
        t.Int64      =  t.Builtin (KInt64 )

        t.UInt       =  t.Builtin (KUInt  )
        t.UInt8      =  t.Builtin (KUInt8 )
        t.UInt16     =  t.Builtin (KUInt16)
        t.UInt32     =  t.Builtin (KUInt32)
        t.UInt64     =  t.Builtin (KUInt64)

        t.Float32    =  t.Builtin (KFloat32)
        t.Float64    =  t.Builtin (KFloat64)
        t.Complex64  =  t.Builtin (KComplex64)
        t.Complex128 =  t.Builtin (KComplex128)

        t.Bool       =  t.Builtin (KBool)
        t.String     =  t.Builtin (KString)
        t.RUint      =  t.Ref     (t.UInt)

        t.Append     =  t.Builtin (KAppend)
        t.Make       =  t.Builtin (KMake)
        t.New        =  t.Builtin (KNew)
        t.Cap        =  t.Builtin (KCap)

        t.Panic      =  t.Builtin (KPanic)
        t.Recover    =  t.Builtin (KRecover)
}

//------------------------------------------------------------------------------

func (b Builtin) Head () string {
        switch b {
        case KAny: return "<any>"
        case KErr: return "<error>"

        case KInt  : return "int"
        case KInt8 : return "int8"
        case KInt16: return "int16"
        case KInt32: return "int32"
        case KInt64: return "int64"

        case KUInt  : return "uint"
        case KUInt8 : return "uint8"
        case KUInt16: return "uint16"
        case KUInt32: return "uint32"
        case KUInt64: return "uint64"

        case KFloat32   : return "float32"
        case KFloat64   : return "float64"
        case KComplex64 : return "complex64"
        case KComplex128: return "complex128"

        case KBool   : return "string"
        case KString : return "string"

        case KAppend : return "<append fn>"
        case KMake   : return "<make fn>"
        case KLen    : return "<len fn>"
        case KNew    : return "<new fn>"
        case KCap    : return "<cap fn>"

        case KPanic  : return "<panic fn>"
        case KRecover: return "<recover fn>"
        }
        return (fmt.Sprintf ("invalid builtin kind: %d",  b))
}

func (b *Builtin) Repr () string { return b.Head () }
func (b *Builtin) Hash () uint   { return uint (*b); }
func (b *Builtin) Equals (m Mode) bool {
        switch m := m.(type) {
        case *Builtin: return *b == *m;
        default: return false
        }
}

func NewBuiltin () *Builtin {
        return nil
}

//------------------------------------------------------------------------------

type Ref struct { Mode }

func (m *Ref) Head () string { return "*" + m.Mode.Head () }
func (m *Ref) Repr () string { return "*" + m.Mode.Repr () }
func (m *Ref) Hash () uint { return m.Mode.Hash () * 3 + 1 }
func (m *Ref) Equals (b Mode) bool {
        switch b := b.(type) {
        case *Ref: return b.Mode == m.Mode
        default: return false
        }
}

//----------------------------------- Temporary for top level scan -------------------------------

type IdType struct {
        Decl *Decl
        Base Mode
        Meths * EnvBldr
}

func (id *IdType) Head () string { return "type " + id.Decl.Name.Name }
func (id *IdType) Repr () string { return id.Head () }
func (id *IdType) Hash () uint { return 0 }
func (id *IdType) Equals (m Mode) bool {
        switch m := m.(type) {
        case *IdType: return id.Decl == m.Decl
        default: return false
        }
}

//------------------------------------------------------------------------------

type Struct struct {
        Flds *Env
}

func (s *Struct) Head () string { return "struct" }
func (s *Struct) Repr () string {
        res := "struct {\n"
        s.Flds.Scan (func (d *Decl) bool {
                res += "  " + d.Name.Name + " " + d.Mode.Repr ()
                return true
        });
        res += "}\n"
        return res
}

func (s *Struct) Hash () uint {
        return 0 // todo
}

func (s *Struct) Equals (m Mode) bool {
        return false //TODO
}

//------------------------------------------------------------------------------

type Func struct {
        Args    []Mode
        Results []Mode
}

func (f *Func) Head () string { return "function" }
func (f *Func) Repr () string {
        res := "Func ("
        sep :=""
        for _, arg := range f.Args {
                res += sep  + arg.Repr ()
                sep = ", "
        }
        res += ") "
        sep = ""
        for _, result:= range f.Results {
                res += sep  +  result.Repr ()
                sep = ", "
        }
        return res
}

func (f *Func) Hash () uint {
        return 0 // todo
}

func (f *Func) Equals (m Mode) bool {
        switch m := m.(type) {
        case *Func:
                if len (f.Args) != len (m.Args) { return false; }
                for i , arg := range (f.Args) {
                        if !arg.Equals (m.Args [i])  { return false; }
                }
                if len (f.Results) != len (m.Results) { return false; }
                for i , result := range (f.Results) {
                        if !result.Equals (m.Results [i])  { return false; }
                }
                return true;
        default: return false
        }
}

//------------------------------------------------------------------------------

type Array struct {
        Size   *int
        Elem    Mode
}

func (v *Array) Head () string {
        size := ""
        if v.Size != nil  {
           size = "#" //v.Size TODO
        }
        return "[" + size + "]" + v.Elem.Repr()
}

func (v *Array) Repr () string { return v.Head () }

func (v *Array) Hash () uint {
        return 0 // todo
}

func (v *Array) Equals (m Mode) bool {
        switch m := m.(type) {
        case *Array:
                if v.Size != m.Size && (v.Size == nil || m.Size == nil) { return false }
		//log.Printf ("v.size=%v m.size=%v\n", v.Size, m.Size)
                return (v.Size == nil && m.Size == nil || *v.Size == *m.Size) && v.Elem.Equals (m.Elem)
        default: return false
        }
}
