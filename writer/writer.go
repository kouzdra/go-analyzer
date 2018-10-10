package writer

import "os"
import "fmt"
import "bufio"
import "strconv"
import "github.com/kouzdra/go-analyzer/defs"
import "github.com/kouzdra/go-analyzer/options"

type Writer struct {
	*bufio.Writer
	LogWriter *bufio.Writer
}

func New(f *os.File) *Writer {
	log, err := os.Create(options.OutLogName)
	if err != nil {
		fmt.Println(err)
	}
	return &Writer{bufio.NewWriter(f), bufio.NewWriter(log)}
}

func (w Writer) Put (s string) Writer {
	w.WriteString(s)
	w.LogWriter.WriteString(s)
	//w.LogWriter.Flush()
	return w
}

func (w Writer) Log (s string) Writer {
	return w //.Put ("LOG: ").Put (s).Eol ().Flush ()
}

func (out Writer) Flush () Writer {
	out.   Writer.Flush ()
	out.LogWriter.Flush()
	return out
}

func (out Writer) End (name string) Writer {
	return out.Put (name).Put ("-END").Eol ().Eol ().Flush ()
}

func (out Writer) Beg (name string) Writer {
	return out.Put (name).Sep ()
}

func (out Writer) Sep () Writer {
	return out.Put (" ")
}

func (out Writer) Eol () Writer {
	return out.Put ("\n")
}

func (out Writer) Write (s string) Writer {
	return out.Put (strconv.QuoteToASCII (s))
}

func (out Writer) WriteInt (n int) Writer {
	return out.Put(strconv.Itoa(n))
}

func (out Writer) WritePos (p defs.Pos) Writer {
	return out.WriteInt(int (p))
}

func (out Writer) WriteRng (p defs.Rng) Writer {
	return out.WritePos(p.Beg).Sep ().WritePos (p.End)
}

