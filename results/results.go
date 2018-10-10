package results

import "github.com/kouzdra/go-analyzer/defs"
import "github.com/kouzdra/go-analyzer/writer"

type Result interface {
	Write(out *writer.Writer)
}

//--------------------------------------------------------------------

type PlainText struct {	Text string }
func (r PlainText) Write(out *writer.Writer) { out.Put ("GO: " + r.Text).Flush () }

//--------------------------------------------------------------------

type Message struct { Text string }
func (m Message) Write(out *writer.Writer) {
	out.Beg ("MESSAGE").Write (m.Text).Eol ().End ("MESSAGE")
}

//--------------------------------------------------------------------

const (
	ERR_NOTE  = "INFO"
	ERR_WARN  = "WARN"
	ERR_ERROR = "ERROR"
)

type Error struct {
	Lvl string
	Rng defs.Rng
	Msg string
}

type Errors struct {
	FName string
	No  int
	Errors []Error
}

type Errs struct {
	errs []Error
}

func NewErrs () Errs {
	return Errs{make([]Error, 0, 10)}
}

func (errs *Errs) Add (lvl string, beg, end defs.Pos, msg string) {
	errs.errs = append (errs.errs, Error{lvl, defs.NewRng (beg, end), msg})
}

func (errs *Errs) GetErrors (fname string, no int) *Errors {
	return &Errors {fname, no, errs.errs}
}

func (errs Errors) Write(out *writer.Writer) {
	out.Beg ("ERRORS").Write(errs.FName).Sep().WriteInt(errs.No).Eol()
	for _, err := range errs.Errors {
		out.Put("  ").Put(err.Lvl).Sep().
			WriteRng(err.Rng).Sep().
			Write(err.Msg).Eol()
	}
	out.End ("ERRORS")
}

//--------------------------------------------------------------------

type FontMarker struct {
	Color string
	Rng defs.Rng
}


type Hils struct {
	hils []FontMarker
}

func NewHils () Hils {
	return Hils{make([]FontMarker, 0, 10000)}
}

func (hils *Hils) Add (color string, beg, end defs.Pos) {
	hils.hils = append(hils.hils, FontMarker{color, defs.NewRng (beg, end)})
}

func (hils *Hils) GetFonts (fname string, bname string, no int, rng defs.Rng) *Fontify {
	return &Fontify{fname, bname, no, rng, hils.hils}
}

type Fontify struct {
	FName string
	BName string
	No  int
	Rng defs.Rng
	Markers []FontMarker
}

func (f Fontify) Write(out *writer.Writer) {
	out.Beg ("FONTIFY").Write(f.FName).Sep().Write(f.BName).Sep().
		WriteInt(f.No).Sep().WriteRng(f.Rng).Eol()
	for _, m := range f.Markers {
		out.Put("  ").Put(m.Color).Sep().WriteRng(m.Rng).Eol()
	}
	out.End ("FONTIFY")
}

//--------------------------------------------------------------------

type Choice struct {
	Kind string
	Name string
	Full string
	Pos  defs.Pos
	End  defs.Pos
}

type Completion struct {
	Pref string
	Name string
	Pos  defs.Pos
	End  defs.Pos
	Choices []Choice
}

func (c Completion) Write (w *writer.Writer) {
	w.Beg("COMPLETE").Write(c.Name).Sep().WritePos(c.Pos).Sep().WritePos(c.End).Sep().Write(c.Pref).Eol()
	for _, c := range c.Choices {
		w.Put("  ").Put(c.Kind).Sep().Write(c.Name).Sep().Write(c.Full).Sep().WritePos(c.Pos).Sep().WritePos(c.End).Eol()
	}
	w.End("COMPLETE")
}

//--------------------------------------------------------------------

type Files struct {
	System bool
	Files  [] string
}

func (f Files) Write (out *writer.Writer) {
	out.Beg ("FILES");
	if f.System  {
	   out.Write ("system")
	} else {
	   out.Write ("non-system")
	}
	out.Eol ()
	for _, file := range f.Files {
	    out.Put ("  ").Write (file).Eol ()
	}
	out.End ("FILES")
}
