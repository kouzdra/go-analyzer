package results

import "sort"
import "github.com/kouzdra/go-analyzer/writer"

const (
	Def = "def"
	Dcl = "decl"
	Ref = "ref"

	Rec   = "recursive"
	NoRec = "no" + Rec
)

type Usage struct {
	Kind string
	FName string
	n1, n2 int
	s1 string
	n3, n4, n5 int
	Name string
	n6, n7 int
	Rec string
}

func (u *Usage) Write (w *writer.Writer) {
	w.Put(u.Kind).Sep().Write(u.FName).Sep().
		WriteInt(u.n1).Sep().
		WriteInt(u.n2).Sep().
		WriteInt(u.n3).Sep().
		Write(u.s1).Sep().
		WriteInt(u.n4).Sep().
		WriteInt(u.n5).Sep().
		Write(u.Name).Sep().
		WriteInt(u.n6).Sep().
		WriteInt(u.n7).Sep().
		Put(NoRec).Eol()
}

func UsageFile(name, fname string) *Usage {
	return &Usage{Def, fname, 0, 0, "", 0, 0, 0, name, 0, 0, NoRec}
}

//====================================================================

type Usages struct {
	Kind string
	System bool
	Pfx  string
	Values []*Usage
}

func (u *Usages) Write (w *writer.Writer) {
	w.Beg("USAGES").Put(u.Kind).Sep().Write(u.Pfx).Sep();
	if u.System { w.Put("system") } else { w.Put("non-system") }
	w.Eol()
	for _, u := range u.Values { u.Write(w) }
	w.End("USAGES")
}

type UsageSorter struct { Vals []*Usage }
func (us *UsageSorter) Len () int { return len(us.Vals) }
func (us *UsageSorter) Swap (i, j int) { us.Vals[i], us.Vals[j] = us.Vals[j], us.Vals[i] }
func (us *UsageSorter) Less (i, j int) bool { return us.Vals[i].Name < us.Vals[j].Name }

func (us Usages) Sort () {
	sort.Sort(&UsageSorter{us.Values})
}

