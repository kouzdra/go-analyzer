package golang

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

type src struct {
	pkg iproject.IPackage
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

func (s *src) GetPackage () iproject.IPackage { return s.pkg ; }
func (s *src) GetDir     () *names.Name { return s.dir ; }
func (s *src) GetName    () *names.Name { return s.name; }
func (s *src) GetAst     () *  ast.File { return s.ast ; }
func (s *src) GetFile    () *token.File { return s.file; }
func (s *src) GetSize    () int { return s.file.Size(); }

func (s *src) GetOuterErrors () scanner.ErrorList { return s.outerErrors; }
func (s *src) GetInnerErrors () []results.Error   { return s.innerErrors; }

//-------------------------------------------------------

func srcNew (pkg iproject.IPackage, dir *names.Name, name *names.Name) *src {
	return &src{pkg, dir, name, false, "", nil, nil, nil, nil}
}

func (src *src) FName () string {
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

func (src *src) GetText() string {
	if !src.actual {
		src.text = readFile(src.FName ())
		src.actual = true
	}
	return src.text
}

func (src *src) Reload () {
	src.actual = false
	src.text   = ""
	src.ast    = nil
	src.file   = nil
	src.outerErrors = nil
	src.pkg.Reload ()
}

func (src *src) SetText (text string) {
	src.Reload ()
	src.text = text
	src.actual = true
}

func (src *src) Changed (pos int, end int, newText string) {
	old := src.GetText()
	src.SetText (old[:pos] + newText + old[end:])
	//src.Reload()U
	//src.text = old[:pos] + newText + old[end:]
	//src.actual = true
}

func (src *src) ReParse () (*token.File, *ast.File, scanner.ErrorList) {
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

func (src *src) UpdateAst () {
	if src.ast == nil {
		src.file, src.ast, src.outerErrors = src.ReParse ()
	}
}
