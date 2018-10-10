package project

//import "os"
//import "fmt"
import "regexp"
import "path"
import "strconv"
import "go/ast"
import "go/token"
import "go/scanner"
//import "github.com/kouzdra/go-analyzer/writer"
import "github.com/kouzdra/go-analyzer/results"
import "github.com/kouzdra/go-analyzer/defs"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"
import "github.com/kouzdra/go-analyzer/golang/env"
import "github.com/kouzdra/go-analyzer/iface/iproject"

const (
	Operator  = "Operator"
	Separator = "Separator"
	Keyword = "Keyword"
	Comment = "Comment"
	Token   = "Token"
	Error   = "Error"
	String  = "String"
	Char    = "Char"
	Number  = "Number"

	Binary  = Operator
	Unary   = Operator

	VarRef  = "Var"
	VarDef  = "VarDef"
	ConRef  = "Con"
	ConDef  = "ConDef"
	ValRef  = VarRef
	ValDef  = VarDef
	TypRef  = "Type"
	TypDef  = "TypeDef"
	FunRef  = "Meth"
	FunDef  = "MethDef"
	PkgRef  = TypRef
	PkgDef  = TypDef
	MthRef  = FunRef
	MthDef  = FunDef
	LblRef  = TypRef
	LblDef  = TypDef
)

const (
	inTypeSwitch       = (1 << iota)
	inSimpleSwitch
	breakAllowed
	continueAllowed
	fallthroughAllowed
)

type NClass int
const (
        InTop = NClass (iota)
        InType
        InExpr
        InStmt
        InDecl
        InPackage
)

type ker struct {
	Hils results.Hils
	Errs results.Errs
	src  *source
	path *paths.Path // TODO
	gbl *env.Env
	lcl *env.Env
	modeTab *env.ModeTab
}

type Analyzer struct {
	*ker
	fileSet *token.FileSet
	file *token.File
	flags uint
	CurrGbl, CurrLcl, CurrSearch *env.EnvBldr
	CollectHils bool
	CollectErrs bool
	NodeClass NClass
	NodeProc IProcessor
}

func newKer(modeTab *env.ModeTab, src *source) *ker {
	return &ker{
		Errs   :results.NewErrs(),
		Hils   :results.NewHils(),
		src    :src,
		gbl    :env.Empty,
		lcl    :env.Empty,
		modeTab:modeTab}
}

//	ker := analyzer.NewKer (p.GetModeTab())
func newAnalyzer(ker *ker, fileSet *token.FileSet, collect  bool) *Analyzer {
	var res Analyzer
	res.ker = ker
	res.fileSet = fileSet
	res.flags = 0
	res.CollectHils = collect
	res.CollectErrs = collect
	res.NodeClass = InTop
	return &res
}

func NewAnalyzer(p iproject.IProject, src iproject.ISource, collect  bool) *Analyzer {
	return newAnalyzer (newKer (p.(*project).GetModeTab (), src.(*source)), p.(*project).GetFileSet(), collect)
}

func (a *Analyzer) withEnv (e *env.Env, fn func ()) {
	oldL, oldS := a.CurrLcl, a.CurrSearch
	a.CurrLcl = env.NewBldr ()
	a.CurrSearch = a.CurrLcl
	a.CurrLcl.Nested (e)
	fn ()
	a.CurrLcl, a.CurrSearch = oldL, oldS
}

func (a *Analyzer) withEnvDecl (b *env.EnvBldr, fn func ()) {
	oldL := a.CurrLcl
	a.CurrLcl = b
	fn ()
	a.CurrLcl = oldL
}

func (a *Analyzer) has  (flags uint) bool { return a.flags & flags != 0 }
func (a *Analyzer) set  (flags uint) { a.flags |= flags }
func (a *Analyzer) clear(flags uint) { a.flags &=^flags }

func (a *Analyzer) withFlags (body func()) {
	flags := a.flags
	body ()
	a.flags = flags
}

func (a *Analyzer) posValid(p token.Pos) bool {
	beg := token.Pos(                a.file.Base())
	end := token.Pos(a.file.Base() + a.file.Size())
	return beg <= p && p <= end
}

func (a *Analyzer) offset (p token.Pos) defs.Pos {
	beg := token.Pos(                a.file.Base())
	end := token.Pos(a.file.Base() + a.file.Size())
	if p < beg { p = beg }
	if p > end { p = end }
	return defs.Pos (a.file.Offset(p))
}

func (a *Analyzer) hil (color string, n ast.Node) *Analyzer {
	a.hilR(color, n.Pos(), n.End())
	return a
}

func (a *Analyzer) hilAt (color string, p token.Pos, l int) *Analyzer {
	return a.hilR(color, p, token.Pos(int (p) + l))
}

func (a *Analyzer) hilR (color string, beg token.Pos,  end token.Pos)  *Analyzer {
	if !a.CollectHils { return a }
	if a.posValid(beg) && a.posValid(end) {
		a.Hils.Add (color, a.offset(beg), a.offset(end))
	}
	return a
}

func (a *Analyzer) errorG (lvl string, beg token.Pos, end token.Pos, msg string) {
	if !a.CollectErrs { return }
	obeg := a.offset(beg)
	oend := a.offset(end)
	if obeg == oend && obeg != 0 { obeg -- }
	a.Errs.Add (lvl, obeg, oend, msg)
}

func (a *Analyzer) note       (beg token.Pos, end token.Pos, msg string) { a.errorG (results.ERR_NOTE , beg, end, msg) }
func (a *Analyzer) warning    (beg token.Pos, end token.Pos, msg string) { a.errorG (results.ERR_WARN , beg, end, msg) }
func (a *Analyzer) error      (beg token.Pos, end token.Pos, msg string) { a.errorG (results.ERR_ERROR, beg, end, msg) }
func (a *Analyzer) newError   (beg token.Pos, end token.Pos, msg string) {  }

//===================================================================

func (a *Analyzer) withNode (kind NClass, node ast.Node, fn func ()) {
	old := a.NodeClass
	a.NodeClass = kind
	defer func () { a.NodeClass = old } ()
	if a.NodeProc != nil &&  node != nil { a.NodeProc.Before (node) }
	fn ()
	if a.NodeProc != nil &&  node != nil { a.NodeProc.After  (node) }
}

//===================================================================

func (a *Analyzer) stmt (s ast.Stmt) {
	a.withNode (InStmt, s, func () {
		switch s := s.(type) {
		case *ast.      ExprStmt: a.expr(s.X)
		case *ast.       BadStmt: a.hil(Error, s)
		case *ast.     BlockStmt: a.     blockStmt(s)
		case *ast.    AssignStmt: a.    assignStmt(s)
		case *ast.    ReturnStmt: a.    returnStmt(s)
		case *ast.    BranchStmt: a.    branchStmt(s)
		case *ast.    IncDecStmt: a.    incDecStmt(s)
		case *ast.TypeSwitchStmt: a.typeSwitchStmt(s)
		case *ast.    SwitchStmt: a.    switchStmt(s)
		case *ast.    SelectStmt: a.    selectStmt(s)
		case *ast.      SendStmt: a.      sendStmt(s)
		case *ast.     RangeStmt: a.     rangeStmt(s)
		case *ast.     EmptyStmt: a.hilAt(Separator, s.Semicolon, 1)
		case *ast.     DeferStmt: a.     deferStmt(s)
		case *ast.   LabeledStmt: a.   labeledStmt(s)
		case *ast.       ForStmt: a.       forStmt(s)
		case *ast.        GoStmt: a.        goStmt(s)
		case *ast.        IfStmt: a.        ifStmt(s)
		case *ast.CaseClause: a.caseClause(s)
		case *ast.CommClause: a.commClause(s)
		default: if s != nil { a.error(s.Pos(), s.End(), "unknown stmt node") }
		}
	})
}

func (a *Analyzer) hilAssn (tok token.Token, pos token.Pos) {
	switch tok {
	case token.DEFINE: a.hilAt(Separator, pos, 2)
	case token.ASSIGN: a.hilAt(Separator, pos, 1)
	}
}

func (a *Analyzer) incDecStmt (s *ast.IncDecStmt) {
	a.expr(s.X)
	a.hilAt(Unary, s.TokPos, 2)
}

func (a *Analyzer) labeledStmt (s *ast.LabeledStmt) {
	a.identExpr(s.Label)
	a.hilAt(Separator, s.Colon, 1)
	a.stmt(s.Stmt)
}

func (a *Analyzer) declIdentExpr (e ast.Expr) {
	switch e := e.(type) {
	case *ast.Ident:
		a.hil(VarDef, e)
		a.declare (env.KVar, names.Put(e.Name), nil, a.modeTab.Any, e)
	default:
		a.error(e.Pos(), e.End(), "ident node required")
	}
}

func (a *Analyzer) assignStmt (s *ast.AssignStmt) {
	//ast.Print(nil, s)
	switch s.Tok {

	case token.DEFINE:
		for _, l := range s.Lhs {
			a.declIdentExpr(l)
		}
		a.hilAt(Separator, s.TokPos, 2)

	default:
		for _, l := range s.Lhs { a.expr(l) }
		a.hilAt(Operator, s.TokPos, len (s.Tok.String()))
	}
	for _, r := range s.Rhs { a.expr(r) }
}

func (a *Analyzer) sendStmt (s *ast.SendStmt) {
	a.expr(s.Chan)
	a.hilAt(Separator, s.Arrow, 2)
	a.expr(s.Value)
}

func (a *Analyzer) rangeStmt (s *ast.RangeStmt) {
	a.CurrLcl.With (func () {
		a.hilAt(Keyword, s.For, 3)
		a.declIdentExpr(s.Key)
		if s.Value != nil { a.declIdentExpr(s.Value) }
		a.hilAssn (s.Tok, s.TokPos)
		a.expr(s.X)
		a.blockStmt(s.Body)
	})
}

func (a *Analyzer) blockStmt (e *ast.BlockStmt) {
	if e != nil {
		if e.Lbrace != 0 { a.hilAt(Token, e.Lbrace, 1) }
		a.CurrLcl.With (func () {
			for _, s := range e.List { a.stmt(s) }
		})
		if e.Rbrace != 0 { a.hilAt(Token, e.Rbrace, 1) }
	}
}

func (a *Analyzer) returnStmt (s *ast.ReturnStmt) {
	a.hilAt(Keyword, s.Return, 6)
	for _, v := range s.Results { a.expr(v) }
}

func (a *Analyzer) branchStmt (s *ast.BranchStmt) {
	switch s.Tok {
	case token.      BREAK: a.hilAt(Keyword, s.TokPos, 5)
	case token.   CONTINUE: a.hilAt(Keyword, s.TokPos, 8)
	case token.       GOTO: a.hilAt(Keyword, s.TokPos, 4)
	case token.FALLTHROUGH: a.hilAt(Keyword, s.TokPos, 11)
	}
	if s.Label != nil { a.hil(LblRef, s.Label) }
}

func (a *Analyzer) selectStmt (s *ast.SelectStmt) {
	a.CurrLcl.With (func () {
		a.hilAt(Keyword, s.Select, 6)
		a.blockStmt(s.Body)
	})
}

func (a *Analyzer) switchStmt (s *ast.SwitchStmt) {
	a.CurrLcl.With (func () {
		a.hilAt(Keyword, s.Switch, 6)
		if s.Init != nil { a.stmt(s.Init) }
		if s.Tag  != nil { a.expr(s.Tag ) }
		a.withFlags (func () { a.set (inSimpleSwitch); a.blockStmt(s.Body)})
	})
}

func (a *Analyzer) typeSwitchStmt (s *ast.TypeSwitchStmt) {
	a.CurrLcl.With (func () {
		a.hilAt(Keyword, s.Switch, 6)
		if s.Init != nil { a.stmt(s.Init) }
		a.stmt(s.Assign)
		a.withFlags (func () { a.set (inTypeSwitch); a.blockStmt(s.Body)})
	})
}

func (a *Analyzer) commClause (s *ast.CommClause) {
	if s.Comm == nil {
		a.hilAt(Keyword, s.Case, 7)
	} else {
		a.hilAt(Keyword, s.Case, 4)
		a.stmt(s.Comm)
	}
	a.hilAt(Separator, s.Colon, 1)
	for _, s := range s.Body {
		a.stmt(s)
	}
}

func (a *Analyzer) caseClause (s *ast.CaseClause) {
	if s.List == nil {
		a.hilAt(Keyword, s.Case, 7)
	} else {
		a.hilAt(Keyword, s.Case, 4)
		for _, v := range s.List {
			if (a.has (inTypeSwitch)) {
				a.typ (v)
			} else {
				a.expr(v)
			}
		}
	}
	a.hilAt(Separator, s.Colon, 1)
	for _, s := range s.Body {
		a.withFlags (func () { a.clear(inTypeSwitch + inSimpleSwitch); a.stmt(s) })
	}
}

func (a *Analyzer) forStmt (s *ast.ForStmt) {
	a.CurrLcl.With (func () {
		a.hilAt(Keyword, s.For, 3)
		if s.Init != nil { a.stmt(s.Init) }
		if s.Cond != nil { a.expr(s.Cond) }
		if s.Post != nil { a.stmt(s.Post) }
		a.blockStmt(s.Body)
	})
}

func (a *Analyzer) goStmt (s *ast.GoStmt) {
	a.hilAt(Keyword, s.Go, 2)
	a.callExpr(s.Call)
}

func (a *Analyzer) deferStmt (s *ast.DeferStmt) {
	a.hilAt(Keyword, s.Defer, 5)
	a.callExpr(s.Call)
}

func (a *Analyzer) ifStmt (s *ast.IfStmt) {
	a.hilAt(Keyword, s.If, 2)
	if s.Init != nil { a.stmt(s.Init) }
	a.expr(s.Cond)
	a.blockStmt(s.Body)
	if s.Else != nil { a.stmt(s.Else) }
}

//===================================================================

func (a *Analyzer) expr (e ast.Expr) env.Mode {
	if mode, ok := a.tryExpr(e); !ok {
		a.error(e.Pos(), e.End(), "invalid expression node")
		return a.typ(e)
	} else {
		return mode
	}
}

func (a *Analyzer) tryExpr (e ast.Expr) (env.Mode, bool) {
	res := env.ModeNil
	ok := true
	a.withNode (InExpr, e, func () {
		switch e := e.(type) {
		case *ast.       BadExpr: a.hil(Error, e); res = a.modeTab.Err
		case *ast.         Ident: res = a.     identExpr(e)
		case *ast.      CallExpr: res = a.      callExpr(e)
		case *ast.      StarExpr: res = a.      starExpr(e)
		case *ast.     IndexExpr: res = a.     indexExpr(e)
		case *ast.     SliceExpr: res = a.     sliceExpr(e)
		case *ast.     ParenExpr: res = a.     parenExpr(e)
		case *ast.     UnaryExpr: res = a.     unaryExpr(e)
		case *ast.    BinaryExpr: res = a.    binaryExpr(e)
		case *ast.  SelectorExpr: res = a.  selectorExpr(e)
		case *ast.  KeyValueExpr: res = a.  keyValueExpr(e)
		case *ast.TypeAssertExpr: res = a.typeAssertExpr(e)
		case *ast.  CompositeLit: res = a.  compositeLit(e)
		case *ast.      BasicLit: res = a.      basicLit(e)
		case *ast.       FuncLit: res = a.       funcLit(e)
		default: if e != nil { ok = false; }; res = a.modeTab.Err
		}
	})
	return res, ok
}

func (a *Analyzer) identExpr (id *ast.Ident) env.Mode {
    if d, any := a.CurrSearch.Find (names.Put(id.Name)); !any && d == nil {
	a.newError(id.Pos(), id.End(), "`" + id.Name + "' undefined")
	return a.modeTab.Err
    } else if d == nil {
	a.hil(VarRef, id)
	return a.modeTab.Any
    } else {
	switch d.Kind {
 	case env.KVar    : a.hil(VarRef, id)
	case env.KType   : a.hil(TypRef, id)
	case env.KFunc   : a.hil(FunRef, id)
	case env.KConst  : a.hil(ConRef, id)
	case env.KPackage: a.hil(PkgRef, id)
	default: a.hil(Error, id)
	}
	//fmt.Printf ("id=%s dMode=%s\n", id.Name, d.Mode.Repr())
	return d.Mode
    }
}

func (a *Analyzer) callExpr (e * ast.CallExpr) env.Mode {
	fMode := a.expr(e.Fun)
	a.hilAt(Token, e.Lparen, 1)
	if fMode == a.modeTab.Make {
	   if len (e.Args) == 0 {
     	       a.error (e.Pos (), e.End (), "type argument required")
	   } else {
		for i, arg := range e.Args {
		    if i == 0 {
			a.typ (arg)
		    } else {
			a.expr(arg)
		    }
		}
	   }
	} else {
	   for _, arg := range e.Args { a.expr(arg) }
	}
	a.hilAt(Token, e.Rparen, 1)
	return a.modeTab.Err
}

func (a *Analyzer) starExpr (t *ast.StarExpr) env.Mode {
	a.hilAt(Operator, t.Star, 1)
	switch m := a.expr(t.X).(type) {
	case *env.Ref: return m.Mode
	default:
		if !env.MAny (m) {
			a.error (t.X.Pos (), t.X.End (), "pointer type required")
		}
		return a.modeTab.Err
	}
}

func (a *Analyzer) indexExpr (e *ast.IndexExpr) env.Mode {
	a.expr(e.X)
	a.hilAt(Token, e.Lbrack, 1)
	a.expr(e.Index)
	a.hilAt(Token, e.Rbrack, 1)
	return a.modeTab.Err
}

func (a *Analyzer) sliceExpr (e *ast.SliceExpr) env.Mode {
	a.expr(e.X)
	a.hilAt(Token, e.Lbrack, 1)
	if e.Low  != nil { a.expr(e.Low ) }
	if e.High != nil { a.expr(e.High) }
	if e.Max  != nil { a.expr(e.Max ) }
	a.hilAt(Token, e.Rbrack, 1)
	return a.modeTab.Err
}

func (a *Analyzer) parenExpr (e *ast.ParenExpr) env.Mode {
	a.hilAt(Token, e.Lparen, 1)
	a.expr(e.X)
	a.hilAt(Token, e.Rparen, 1)
	return a.modeTab.Err
}

func (a *Analyzer) selectorExpr (e *ast.SelectorExpr) env.Mode {
	m := a.expr(e.X)
	res := env.ModeNil

	switch mm := m.(type) {
	case *env.IdType: m = env.DerefOpt (mm.Base)
	case *env.Ref   : m = env.DeId     (mm.Mode)
	}
	switch m := m.(type) {
	case *env.Struct:
		a.withEnv(m.Flds, func () { res = a.expr(e.Sel) })
	case *env.Builtin:
		if *m != env.KErr && *m != env.KAny {
			a.error(e.X.Pos(), e.X.End(), "struct type required instead of '" + m.Head() + "'")
		}
		a.withEnv(env.Any, func () { a.expr(e.Sel) })
		res = a.modeTab.Err
	default:
		a.error(e.X.Pos(), e.X.End(), "struct type required instead of '" + m.Head () + "'")
		a.withEnv(env.Any, func () { a.expr(e.Sel) })
		res = a.modeTab.Err
	}
	return res
}

func (a *Analyzer) keyValueExpr (e *ast.KeyValueExpr) env.Mode {
	a.expr(e.Key)
	a.hilAt(Separator, e.Colon, 1)
	a.expr(e.Value)
	return a.modeTab.Err
}

func (a *Analyzer) typeAssertExpr (e *ast.TypeAssertExpr) env.Mode {
	a.expr(e.X)
	a.hilAt(Token, e.Lparen, 1)
	a.typ (e.Type)
	a.hilAt(Token, e.Rparen, 1)
	return a.modeTab.Err
}

func (a *Analyzer) unaryExpr (e *ast.UnaryExpr) env.Mode {
	a.hilAt(Unary, e.OpPos, len(e.Op.String()))
	a.expr(e.X)
	return a.modeTab.Err
}

func (a *Analyzer) binaryExpr (e *ast.BinaryExpr) env.Mode {
	a.expr(e.X)
	a.hilAt(Binary, e.OpPos, len(e.Op.String()))
	a.expr(e.Y)
	return a.modeTab.Err
}

func (a *Analyzer) funcLit (f *ast.FuncLit) env.Mode {
	a.funcType (f.Type)
	a.blockStmt(f.Body)
	return a.modeTab.Err
}

func (a *Analyzer) basicLit (c *ast.BasicLit) env.Mode {
	switch c.Kind {
	case token.STRING: a.hil(String   , c); return a.modeTab.String
	case token.  CHAR: a.hil(Char     , c); return a.modeTab.Int
	case token.   INT: a.hil(Number   , c); return a.modeTab.Int
	case token. FLOAT: a.hil(Number   , c); return a.modeTab.Float64
	case token.  IMAG: a.hil(Number   , c); return a.modeTab.Complex128
	}
	return a.modeTab.Err
}

func (a *Analyzer) compositeLit (e *ast.CompositeLit) env.Mode {
	a.typ (e.Type)
	a.hilAt(Token, e.Lbrace, 1)
	for _, elt := range e.Elts { a.expr(elt) }
	a.hilAt(Token, e.Rbrace, 1)
	return a.modeTab.Err
}

//===================================================================

func (a *Analyzer) typ (t ast.Expr) env.Mode {
	if m, ok := a.tryType(t); !ok {
		a.error(t.Pos(), t.End(), "invalid type node")
		return a.expr(t)
	} else {
		return m
	}
}

func (a *Analyzer) tryType (t ast.Expr) (env.Mode, bool) {
	res := env.ModeNil
	ok := true;
	a.withNode (InType, t, func () {
		switch t := t.(type) {
		case *ast.        Ident: res = a.    identType(t)
		case *ast.      BadExpr: a.hil(Error,  t); res = a.modeTab.Err
		case *ast.      MapType: res = a.      mapType(t)
		case *ast.     StarExpr: res = a.     starType(t)
		case *ast.     FuncType: res = a.     funcType(t)
		case *ast.     ChanType: res = a.     chanType(t)
		case *ast.    ArrayType: res = a.    arrayType(t)
		case *ast.   StructType: res = a.   structType(t)
		case *ast.     Ellipsis: res = a. ellipsisType(t)
		case *ast.InterfaceType: res = a.interfaceType(t)
		case *ast. SelectorExpr: res = a. selectorType(t)
		default: ok = (t == nil); res = a.modeTab.Err
		}
	})
	return res, ok
}

func (a *Analyzer) identType (id *ast.Ident) env.Mode {
    if d, any := a.CurrSearch.Find (names.Put(id.Name)); !any && d == nil {
	a.error(id.Pos(), id.End(), "type `" + id.Name + "' undefined")
	return a.modeTab.Err
    } else if d == nil {
	a.hil(TypRef, id)
	return a.modeTab.Err
    } else {
	switch d.Kind {
 	case env.KVar    : a.hil(VarRef, id)
	case env.KType   : a.hil(TypRef, id)
	case env.KFunc   : a.hil(FunRef, id)
	case env.KConst  : a.hil(ConRef, id)
	case env.KPackage: a.hil(PkgRef, id)
	default: a.hil(Error, id)
	}
	if d.Kind != env.KType && d.Kind != env.KPackage {
		a.error(id.Pos(), id.End(), "type ident required")
	}
	return d.Mode
    }
}

func (a *Analyzer) ellipsisType (t *ast.Ellipsis) env.Mode {
	a.hilAt(Token, t.Ellipsis, 3)
	a.typ(t.Elt)
	return a.modeTab.Err
}

func (a *Analyzer) starType (t *ast.StarExpr) env.Mode {
	a.hilAt(Operator, t.Star, 1)
	m := a.typ(t.X)
	return a.modeTab.Ref (m)
}

func (a *Analyzer) chanType (t *ast.ChanType) env.Mode {
	if t.Arrow == token.NoPos || t.Begin < t.Arrow {
	  a.hilAt(Keyword, t.Begin, 4)
	  if  t.Arrow != token.NoPos {
	    a.hilAt(Separator, t.Arrow, 2)
	  }
	} else {
	  a.hilAt(Separator, t.Begin, 2)
	}
	a.typ(t.Value)
	return a.modeTab.Err
}

func (a *Analyzer) selectorType (t *ast.SelectorExpr) env.Mode {
	a.typ(t.X)
	a.withEnv(env.Any, func () { a.typ(t.Sel) })
	return a.modeTab.Err

}

func (a *Analyzer) funcType (t *ast.FuncType) env.Mode {
	a.hilAt(Keyword, t.Func, 4)
	a.fieldList(t.Params)
	a.fieldList(t.Results)
	return a.modeTab.Err
}

func (a *Analyzer) structType (t *ast.StructType) env.Mode {
	a.hilAt(Keyword, t.Struct, 6)
	bldr := env.NewBldr ()
	a.withEnvDecl (bldr, func () {
		a.fieldList(t.Fields)
	})
	return a.modeTab.Struct(bldr.Close ())
}

func (a *Analyzer) interfaceType (t *ast.InterfaceType) env.Mode {
	a.hilAt(Keyword, t.Interface, 9)
	a.fieldList(t.Methods)
	return a.modeTab.Err
}

func (a *Analyzer) arrayType (t *ast.ArrayType) env.Mode {
	a.hilAt(Token, t.Lbrack, 1)
	var size *int = nil
	if t.Len != nil { /*sizeV := a.expr(t.Len); size = &sizeV*/ } // TODO: size
	elem := a.typ(t.Elt)
	return a.modeTab.Array (size, elem)
}

func (a *Analyzer) mapType (t *ast.MapType) env.Mode {
	a.hilAt(Keyword, t.Map, 3)
	a.typ(t.Key)
	a.typ(t.Value)
	return a.modeTab.Err
}

//===================================================================

func (a *Analyzer) declare (k env.Kind, n *names.Name, t ast.Expr, m env.Mode, v ast.Node) {
	m = a.modeTab.Find (m);
	if n.Gbl () {
		a.CurrGbl.Declare(k, n, a.path, t, m, v)
	} else {
		a.CurrLcl.Declare(k, n, a.path, t, m, v)
	}
}

func (a *Analyzer) decl (d ast.Decl) {
	a.withNode (InDecl, d, func () {
		switch d := d.(type) {
		case *ast. BadDecl: a.hil(Error, d)
		case *ast. GenDecl: a. genDecl(d)
		case *ast.FuncDecl: a.funcDecl(d)
		default: if d != nil { a.error(d.Pos(), d.End(), "unknown decl node") }
		}
	})
}

func (a *Analyzer) funcDecl (d *ast.FuncDecl) {
	// make types
	//a.FuncType(d.Type)
	//if d.Recv != nil { a.FieldList(d.Recv) }
	a.hil(FunDef, d.Name)
	a.declare(env.KFunc, names.Put(d.Name.Name), d.Type, a.modeTab.Any, d)
}

func (a *Analyzer) genDecl (d *ast.GenDecl) {
	switch d.Tok {
	case token.IMPORT: a.hilAt(Keyword, d.Pos(), 6)
	case token.CONST : a.hilAt(Keyword, d.Pos(), 5)
	case token.TYPE  : a.hilAt(Keyword, d.Pos(), 4)
	case token.VAR   : a.hilAt(Keyword, d.Pos(), 3)
	}

	if d.Lparen != 0 { a.hilAt(Token, d.Lparen, 1) }

	for _, s := range d.Specs {
		switch s := s.(type) {
		case *ast.ImportSpec: a.importSpec(s)
		case *ast. ValueSpec: a. valueSpec(s)
		case *ast.  TypeSpec: a.  typeSpec(s)
		default: if s != nil { a.error(s.Pos(), s.End(), "unknown spec node") }
		}
	}

	if d.Rparen != 0 { a.hilAt(Token, d.Rparen, 1) }
}

func (a *Analyzer) field (f *ast.Field) {
	m := a.typ(f.Type)
	for _, id := range f.Names {
		a.CurrLcl.Declare(env.KVar, names.Put(id.Name), a.path, f.Type, m, nil)
		//fmt.Println (id)
		a.hil (VarDef, id)
	}
	//fmt.Println(f.Type)
	//ast.Fprint(os.Stdout, nil, f, nil)
	if f.Tag != nil { a.basicLit (f.Tag) }
	if f.Comment != nil { a.hil (Comment, f.Comment) }
}

func (a *Analyzer) fieldList (fl *ast.FieldList) {
	if fl != nil {
		if fl.Opening != 0 { a.hilAt(Token, fl.Opening, 1) }
		for _, f := range fl.List {
			a.field(f)
		}
		if fl.Closing != 0 { a.hilAt(Token, fl.Closing, 1) }
	}
}

//===================================================================

func (a *Analyzer) importSpec (imp *ast.ImportSpec) {
	if imp.Name != nil {
		a.hil (PkgDef, imp.Name)
		a.declare(env.KPackage, names.Put(imp.Name.Name), nil, a.modeTab.Any, imp)
	} else {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			a.error(imp.Path.Pos(), imp.Path.End(), "invalid path syntax")
		} else {
			a.declare(env.KPackage, names.Put (path.Base(p)), nil, a.modeTab.Any, imp) // TODO: get file name
		}
	}
	a.basicLit(imp.Path)
	if imp.Comment != nil { a.hil (Comment, imp.Comment) }
}

func (a *Analyzer) valueSpec (v *ast.ValueSpec) {
	for i, id := range v.Names {
		a.hil (ValDef, id)
		if v.Values == nil {
			a.declare(env.KConst, names.Put(id.Name), v.Type, a.modeTab.Any, nil)
		} else {
			a.declare(env.KConst, names.Put(id.Name), v.Type, a.modeTab.Any, v.Values[i])
		}
	}
	if v.Comment != nil { a.hil (Comment, v.Comment) }
}

func (a *Analyzer) typeSpec (t *ast.TypeSpec) {
	a.hil (TypDef, t.Name)
	a.declare(env.KType, names.Put(t.Name.Name), t.Type, a.modeTab.Any, nil)

	if t.Comment != nil { a.hil (Comment, t.Comment) }
}

//===================================================================

func (a *Analyzer) analyzeFileIntr (f *ast.File) {
	a.CurrGbl = env.NewBldr()
	a.CurrGbl.Nested (a.modeTab.BEnv)
	a.CurrLcl = env.NewBldr()
	a.CurrSearch = a.CurrLcl
	a.hilAt (Keyword, f.Package, 7)
	a.hil   (PkgDef , f.Name)
	a.declare (env.KPackage, names.Put(f.Name.Name), nil, a.modeTab.Any, nil)
	if f.Doc != nil { a.hil (Comment, f.Doc) }
 	for _, c    := range f.Comments   { a.hil   (Comment, c) }
	for _, decl := range f.Decls      { a.decl (decl) }
	a.gbl = a.CurrGbl.Close()
	a.CurrLcl.Nested (a.gbl)
	a.lcl = a.CurrLcl.Close()
}

//-----------------------------------------------------------------

func (a *Analyzer) analyzeFileBody (f *ast.File) {
	a.CurrLcl = env.NewBldr ()
	a.CurrGbl = a.CurrLcl
	a.CurrGbl.Nested(a.lcl)
	a.CurrSearch = a.CurrLcl

	a.CurrLcl.Scan(func (d *env.Decl) bool {
		if d.Kind == env.KType {
		        //fmt.Printf ("d.Name=%v, mode=%s\n", d.Name, d.Mode.Repr ())
			d.Mode = a.modeTab.IdType (d, a.typ(d.Type))
		}
		return true
	})

	scan := func (ftr func (d *env.Decl) bool) {

		a.CurrLcl.Scan (func (d *env.Decl) bool {
			if (!ftr (d)) { return  true }
			m := d.Mode
			if d.Type != nil { m = a.typ(d.Type) }
			if d.Kind == env.KType {
				m = a.modeTab.IdType (d, m)
			}
			switch n := d.Value.(type) {
			case *ast.FuncDecl:
				a.funcType(n.Type)
				if n.Recv != nil { a.fieldList(n.Recv) }
				a.blockStmt(n.Body)
			case  ast.Expr:
				em := a.expr (n)
				if d.Type == nil { m = em }
			default: break
			}
			//fmt.Printf("UPD: %s mode: %s => %s\n", d.Name.Name, d.Mode.Repr (), m.Repr ())
			d.Mode = m
			return true
		})
	}
	//scan (func (d *env.Decl) bool { return  d.Kind == env.KType  })
	scan (func (d *env.Decl) bool { return  d.Kind == env.KConst })
	scan (func (d *env.Decl) bool { return  d.Kind == env.KVar   })
	scan (func (d *env.Decl) bool { return  d.Kind == env.KFunc  })

	//for _, decl := range f.Decls      { a.decl (decl) }
 	//for _, id   := range f.Unresolved { a.hil   (Error, id)  }
}

//===================================================================

func (a *Analyzer) setOuterErrors (e scanner.ErrorList) {
	for _, err := range e {
		length := 1
		prefix := "expected '.', found 'IDENT' "
		if res, _ := regexp.MatchString (prefix + "[_A-Za-z]+", err.Msg); res {
			length = len (err.Msg) - len (prefix)
		}
		pos := a.file.Pos(err.Pos.Offset)
		a.error (pos, token.Pos (int (pos)+length), err.Msg)
	}
}

func (a *Analyzer) Analyze () {
	a.setOuterErrors (a.src.outerErrors)
	a.file =          a.src.GetFile()
	a.analyzeFileIntr(a.src.GetAst())
	a.analyzeFileBody(a.src.GetAst())
}
