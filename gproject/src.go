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

type Src struct {
	Pkg *Pkg
	Dir  string
	Name string
	actual bool
	text string
	File *token.File
	Ast  *ast.File
	OuterErrors scanner.ErrorList
	InnerErrors []results.Error
}

func SrcNew (pkg *Pkg, dir string, name string) *Src {
	return &Src{pkg, dir, name, false, "", nil, nil, nil, nil}
}

func (src *Src) FName () string {
	return filepath.Join(src.Dir, src.Name)
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

func (src *Src) Text() string {
	if !src.actual {
		src.text = readFile(src.FName ())
		src.actual = true
	}
	return src.text
}

func (src *Src) Reload () {
	src.actual = false
	src.text   = ""
	src.Ast    = nil
	src.File   = nil
	src.OuterErrors = nil
	src.Pkg.Reload ()
}

func (src *Src) SetText (text string) {
	src.Reload ()
	src.text = text
	src.actual = true
}

func (src *Src) Changed (pos int, end int, newText string) {
	old := src.Text()
	src.SetText (old[:pos] + newText + old[end:])
	//src.Reload()
	//src.text = old[:pos] + newText + old[end:]
	//src.actual = true
}

func (src *Src) ReParse () (*token.File, *ast.File, scanner.ErrorList) {
	//src.Changed(3, 4, "a")
	base := token.Pos(src.Pkg.Prj.FSet.Base())
	ast, err := parser.ParseFile (src.Pkg.Prj.FSet, src.FName(), src.Text(), parser.ParseComments)
	file := src.Pkg.Prj.FSet.File(base)
	elist := scanner.ErrorList(nil)
	switch err := err.(type) {
	case scanner.ErrorList: elist = err
	}
	if err != nil && elist == nil || ast == nil {
		src.Pkg.Prj.MsgF("PARSE %s Failed: %v", src.Name, err)
		return nil, nil, nil
	}
	src.Pkg.Prj.MsgF("PARSE %s OK", src.Name)
	return file, ast, elist
}

func (src *Src) UpdateAst () {
	if src.Ast == nil {
		src.File, src.Ast, src.OuterErrors = src.ReParse ()
	}
}
