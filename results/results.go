package results

import "writer"
import "go/token"

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
	Beg int
	End int
	Msg string
}

type Errors struct {
	FName string
	No  int
	Errors []Error
}

func (errs Errors) Write(out *writer.Writer) {
	out.Beg ("ERRORS").Write(errs.FName).Sep().WriteInt(errs.No).Eol()
	for _, err := range errs.Errors {
		out.Put("  ").Put(err.Lvl).Sep().
			WriteInt(err.Beg).Sep().
			WriteInt(err.End).Sep().
			Write(err.Msg).Eol()
	}
	out.End ("ERRORS")
}

//--------------------------------------------------------------------

type FontMarker struct {
	Color string
	Beg int
	End int
}

type Fontify struct {
	FName string
	BName string
	No  int
	Beg int
	End int
	Markers []FontMarker
}

func (f Fontify) Write(out *writer.Writer) {
	out.Beg ("FONTIFY").Write(f.FName).Sep().Write(f.BName).Sep().
		WriteInt(f.No).Sep().WriteInt(f.Beg).Sep().WriteInt(f.End).Eol()
	for _, m := range f.Markers {
		out.Put("  ").Put(m.Color).Sep().
			WriteInt(m.Beg).Sep().
			WriteInt(m.End).Eol()
	}
	out.End ("FONTIFY")
}

//--------------------------------------------------------------------

type Choice struct {
	Kind string
	Name string
	Full string
	Pos  token.Pos
	End  token.Pos
}

type Completion struct {
	Pref string
	Name string
	Pos  token.Pos
	End  token.Pos
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
