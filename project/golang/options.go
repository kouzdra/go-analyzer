package golang

type OptionsBase struct {
	vals map[string]string
}

func NewOptionsBase () OptionsBase {
	return OptionsBase{make(map[string]string, 128)}
}

func (opt *OptionsBase) Set(name, val string) {
	opt.vals[name] = val
}

func (opt *OptionsBase) Clear(name string) {
	delete(opt.vals, name)
}

func (opt *OptionsBase) Get(name string) (string, bool) {
	res, ok := opt.vals[name]
	return res, ok
}

func (opt *OptionsBase) GetDefault(name, def string) string {
	if res, ok := opt.Get(name); ok {
		return res
	}
	return def
}
