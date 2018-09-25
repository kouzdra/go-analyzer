package gproject

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
//import "github.com/kouzdra/go-analyzer/paths"

type Src struct {
	pkg Package
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

func (s *Src) GetPackage () Package     { return s.pkg ; }
func (s *Src) GetDir     () *names.Name { return s.dir ; }
func (s *Src) GetName    () *names.Name { return s.name; }
func (s *Src) GetAst     () *  ast.File { return s.ast ; }
func (s *Src) GetFile    () *token.File { return s.file; }

func (s *Src) GetOuterErrors () scanner.ErrorList { return s.outerErrors; }
func (s *Src) GetInnerErrors () []results.Error   { return s.innerErrors; }

//-------------------------------------------------------

func SrcNew (pkg Package, dir *names.Name, name *names.Name) *Src {
	return &Src{pkg, dir, name, false, "", nil, nil, nil, nil}
}

func (src *Src) FName () string {
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

func (src *Src) GetText() string {
	if !src.actual {
		src.text = readFile(src.FName ())
		src.actual = true
	}
	return src.text
}

func (src *Src) Reload () {
	src.actual = false
	src.text   = ""
	src.ast    = nil
	src.file   = nil
	src.outerErrors = nil
	src.pkg.Reload ()
}

func (src *Src) SetText (text string) {
	src.Reload ()
	src.text = text
	src.actual = true
}

func (src *Src) Changed (pos int, end int, newText string) {
	old := src.GetText()
	src.SetText (old[:pos] + newText + old[end:])
	//src.Reload()U
	//src.text = old[:pos] + newText + old[end:]
	//src.actual = true
}

func (src *Src) ReParse () (*token.File, *ast.File, scanner.ErrorList) {
	//src.Changed(3, 4, "a")
	fset := src.GetPackage().GetProject ().GetFileSet()
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

func (src *Src) UpdateAst () {
	if src.ast == nil {
		src.file, src.ast, src.outerErrors = src.ReParse ()
	}
}
