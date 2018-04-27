package iface

type Options interface {
	Set (name, val string)
	Get (name string) *string
	Clear (name string)
	GetDefault (name, def string) string
}
