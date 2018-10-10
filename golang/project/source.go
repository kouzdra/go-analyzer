package project

//import "os"
//import "io"
import "io/ioutil"
import "github.com/kouzdra/go-analyzer/results"
import "path/filepath"
import "go/scanner"
import "go/parser"
import "go/token"
import "go/ast"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/iface/iproject"

type source struct {
	pkg *packag
	dir  *names.Name
	name *names.Name
	actual bool
	text string
	file *token.File
	ast  *ast.File
	outerErrors scanner.ErrorList
	innerErrors []results.Error
}

//-------------------------------------------------------

func (s *source) GetPackage () iproject.IPackage { return s.pkg ; }
func (s *source) GetDir     () *names.Name { return s.dir ; }
func (s *source) GetName    () *names.Name { return s.name; }
func (s *source) GetAst     () *  ast.File { return s.ast ; }
func (s *source) GetFile    () *token.File { return s.file; }
func (s *source) GetSize    () int { return s.file.Size(); }

func (s *source) GetOuterErrors () scanner.ErrorList { return s.outerErrors; }
func (s *source) GetInnerErrors () []results.Error   { return s.innerErrors; }

//-------------------------------------------------------

func NewSource (pkg iproject.IPackage, dir *names.Name, name *names.Name) *source {
	return &source{pkg.(*packag), dir, name, false, "", nil, nil, nil, nil}
}

func (src *source) FName () string {
	return filepath.Join(src.GetDir().Name, src.GetName().Name)
}

func readFile (fname string) string {
	text, err := ioutil.ReadFile (fname)
	if err == nil {
		return string (text)
	}
	return ""

	/*if in, err := os.Open(fname); err != nil {
		return ""
	} else {
		defer in.Close ()
		bytes := make ([]byte, 0, 16000)
		buf := []byte{0}
		for {
			_, err := in.Read(buf)
			if err == io.EOF { return string(bytes) }
			bytes = append(bytes, buf[0])
		}
	}*/
}

func (src *source) GetText() string {
	if !src.actual {
		src.text = readFile(src.FName ())
		src.actual = true
	}
	return src.text
}

func (src *source) Reload () {
	src.actual = false
	src.text   = ""
	src.ast    = nil
	src.file   = nil
	src.outerErrors = nil
	src.pkg.Reload ()
}

func (src *source) SetText (text string) {
	src.Reload ()
	src.text = text
	src.actual = true
}

func (src *source) Changed (pos int, end int, newText string) {
	old := src.GetText()
	src.SetText (old[:pos] + newText + old[end:])
	//src.Reload()U
	//src.text = old[:pos] + newText + old[end:]
	//src.actual = true
}

func (src *source) reParse () (*token.File, *ast.File, scanner.ErrorList) {
	//src.Changed(3, 4, "a")
	fset := src.GetPackage().GetProject ().(*project).GetFileSet()
	base := token.Pos(fset.Base())
	ast, err := parser.ParseFile (fset, src.FName(), src.GetText(), parser.ParseComments)
	file := fset.File(base)
	elist := scanner.ErrorList(nil)
	switch err := err.(type) {
	case scanner.ErrorList: elist = err
	}
	if err != nil && elist == nil || ast == nil {
		src.GetPackage().GetProject().MsgF("PARSE %s Failed: %v", src.GetName().Name, err)
		return nil, nil, nil
	}
	src.GetPackage().GetProject().MsgF("PARSE %s OK", src.GetName().Name)
	return file, ast, elist
}

func (src *source) updateAst () {
	if src.ast == nil {
		src.file, src.ast, src.outerErrors = src.reParse ()
	}
}
