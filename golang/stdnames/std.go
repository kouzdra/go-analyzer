package stdnames

import "github.com/kouzdra/go-analyzer/names"

var Dummy = names.Put("_")
var Null  = names.Put("<NONE>")

var N_main = names.Put("main")

var N_int    = names.Put("int"   )
var N_int8   = names.Put("int8"  )
var N_int16  = names.Put("int16" )
var N_int32  = names.Put("int32" )
var N_int64  = names.Put("int64" )

var N_uint   = names.Put("uint"  )
var N_uint8  = names.Put("uint8" )
var N_uint16 = names.Put("uint16")
var N_uint32 = names.Put("uint32")
var N_uint64 = names.Put("uint64")

var N_float32    = names.Put("float32"   )
var N_float64    = names.Put("float64"   )
var N_complex64  = names.Put("complex64" )
var N_complex128 = names.Put("complex128")

var N_uintptr = names.Put("uintptr")
var N_bool    = names.Put("bool"   )
var N_byte    = names.Put("byte"   )
var N_rune    = names.Put("rune"   )
var N_string  = names.Put("string" )

var N_nil   = names.Put("nil"   )
var N_iota  = names.Put("iota"  )
var N_true  = names.Put("true"  )
var N_false = names.Put("false" )

var N_append = names.Put("append" )
var N_make   = names.Put("make"   )
var N_new    = names.Put("new"    )
var N_len    = names.Put("len"    )
var N_cap    = names.Put("cap"    )

var N_panic   = names.Put("panic"  )
var N_recover = names.Put("recover")
