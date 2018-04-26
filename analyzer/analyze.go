package analyzer

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
import "github.com/kouzdra/go-analyzer/env"
import "github.com/kouzdra/go-analyzer/names"

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

type Ker struct {
	Gbl *env.Env
	Lcl *env.Env
	errs []results.Error
	hils []results.FontMarker
	ModeTab *env.ModeTab
}

type Analyzer struct {
	*Ker
	FSet *token.FileSet
	file *token.File
	flags uint
	CurrGbl, CurrLcl, CurrSearch * env.EnvBldr
	CollectHils bool
	CollectErrs bool
	NodeClass NClass
	NodeProc IProcessor
}

func NewKer(modeTab *env.ModeTab) *Ker {
	var res Ker
	res.Gbl = env.Empty
	res.Lcl = env.Empty
	res.ModeTab = modeTab
	res.errs = make([]results.Error     , 0, 10)
	res.hils = make([]results.FontMarker, 0, 10000)
	return &res
}

func New(ker *Ker, fSet *token.FileSet, collect  bool) *Analyzer {
	var res Analyzer
	res.Ker = ker
	res.FSet = fSet
	res.flags = 0
	res.CollectHils = collect
	res.CollectErrs = collect
	res.NodeClass = InTop
	return &res
}

func (a *Analyzer) WithEnv (e *env.Env, fn func ()) {
	oldL, oldS := a.CurrLcl, a.CurrSearch
	a.CurrLcl = env.NewBldr ()
	a.CurrSearch = a.CurrLcl
	a.CurrLcl.Nested (e)
	fn ()
	a.CurrLcl, a.CurrSearch = oldL, oldS
}

func (a *Analyzer) WithEnvDecl (b *env.EnvBldr, fn func ()) {
	oldL := a.CurrLcl
	a.CurrLcl = b
	fn ()
	a.CurrLcl = oldL
}

func (a *Analyzer) Has  (flags uint) bool { return a.flags & flags != 0 }
func (a *Analyzer) Set  (flags uint) { a.flags |= flags }
func (a *Analyzer) Clear(flags uint) { a.flags &=^flags }

func (a *Analyzer) WithFlags (body func()) {
	flags := a.flags
	body ()
	a.flags = flags
}

func (a *Analyzer) PosValid(p token.Pos) bool {
	beg := token.Pos(                a.file.Base())
	end := token.Pos(a.file.Base() + a.file.Size())
	return beg <= p && p <= end
}

func (a *Analyzer) Offset (p token.Pos) int {
	beg := token.Pos(                a.file.Base())
	end := token.Pos(a.file.Base() + a.file.Size())
	if p < beg { p = beg }
	if p > end { p = end }
	return a.file.Offset(p)
}

func (a *Analyzer) Hil (color string, n ast.Node) *Analyzer {
	a.HilR(color, n.Pos(), n.End())
	return a
}

func (a *Analyzer) HilAt (color string, p token.Pos, l int) *Analyzer {
	return a.HilR(color, p, token.Pos(int (p) + l))
}

func (a *Analyzer) HilR (color string, beg token.Pos,  end token.Pos)  *Analyzer {
	if !a.CollectHils { return a }
	if a.PosValid(beg) && a.PosValid(end) {
		a.hils = append(a.hils, results.FontMarker{color, a.Offset(beg), a.Offset(end)})
	}
	return a
}

func (a *Analyzer) error (lvl string, beg token.Pos, end token.Pos, msg string) {
	if !a.CollectErrs { return }
	obeg := a.Offset(beg)
	oend := a.Offset(end)
	if obeg == oend && obeg != 0 { obeg -- }
	a.errs = append (a.errs, results.Error{lvl, obeg, oend, msg})
}

func (a *Analyzer) Note       (beg token.Pos, end token.Pos, msg string) { a.error (results.ERR_NOTE , beg, end, msg) }
func (a *Analyzer) Warning    (beg token.Pos, end token.Pos, msg string) { a.error (results.ERR_WARN , beg, end, msg) }
func (a *Analyzer) Error      (beg token.Pos, end token.Pos, msg string) { a.error (results.ERR_ERROR, beg, end, msg) }
func (a *Analyzer) NewError   (beg token.Pos, end token.Pos, msg string) {  }

//===================================================================

func (a *Analyzer) WithNode (kind NClass, node ast.Node, fn func ()) {
	old := a.NodeClass
	a.NodeClass = kind
	defer func () { a.NodeClass = old } ()
	if a.NodeProc != nil &&  node != nil { a.NodeProc.Before (node) }
	fn ()
	if a.NodeProc != nil &&  node != nil { a.NodeProc.After  (node) }
}

//===================================================================

func (a *Analyzer) Stmt (s ast.Stmt) {
	a.WithNode (InStmt, s, func () {
		switch s := s.(type) {
		case *ast.      ExprStmt: a.Expr(s.X)
		case *ast.       BadStmt: a.Hil(Error, s)
		case *ast.     BlockStmt: a.     BlockStmt(s)
		case *ast.    AssignStmt: a.    AssignStmt(s)
		case *ast.    ReturnStmt: a.    ReturnStmt(s)
		case *ast.    BranchStmt: a.    BranchStmt(s)
		case *ast.    IncDecStmt: a.    IncDecStmt(s)
		case *ast.TypeSwitchStmt: a.TypeSwitchStmt(s)
		case *ast.    SwitchStmt: a.    SwitchStmt(s)
		case *ast.    SelectStmt: a.    SelectStmt(s)
		case *ast.      SendStmt: a.      SendStmt(s)
		case *ast.     RangeStmt: a.     RangeStmt(s)
		case *ast.     EmptyStmt: a.HilAt(Separator, s.Semicolon, 1)
		case *ast.     DeferStmt: a.     DeferStmt(s)
		case *ast.   LabeledStmt: a.   LabeledStmt(s)
		case *ast.       ForStmt: a.       ForStmt(s)
		case *ast.        GoStmt: a.        GoStmt(s)
		case *ast.        IfStmt: a.        IfStmt(s)
		case *ast.CaseClause: a.CaseClause(s)
		case *ast.CommClause: a.CommClause(s)
		default: if s != nil { a.Error(s.Pos(), s.End(), "unknown stmt node") }
		}
	})
}

func (a *Analyzer) hilAssn (tok token.Token, pos token.Pos) {
	switch tok {
	case token.DEFINE: a.HilAt(Separator, pos, 2)
	case token.ASSIGN: a.HilAt(Separator, pos, 1)
	}
}

func (a *Analyzer) IncDecStmt (s *ast.IncDecStmt) {
	a.Expr(s.X)
	a.HilAt(Unary, s.TokPos, 2)
}

func (a *Analyzer) LabeledStmt (s *ast.LabeledStmt) {
	a.IdentExpr(s.Label)
	a.HilAt(Separator, s.Colon, 1)
	a.Stmt(s.Stmt)
}

func (a *Analyzer) DeclIdentExpr (e ast.Expr) {
	switch e := e.(type) {
	case *ast.Ident:
		a.Hil(VarDef, e)
		a.Declare (env.KVar, names.Put(e.Name), nil, a.ModeTab.Any, e)
	default:
		a.Error(e.Pos(), e.End(), "ident node required")
	}
}

func (a *Analyzer) AssignStmt (s *ast.AssignStmt) {
	//ast.Print(nil, s)
	switch s.Tok {

	case token.DEFINE:
		for _, l := range s.Lhs {
			a.DeclIdentExpr(l)
		}
		a.HilAt(Separator, s.TokPos, 2)

	default:
		for _, l := range s.Lhs { a.Expr(l) }
		a.HilAt(Operator, s.TokPos, len (s.Tok.String()))
	}
	for _, r := range s.Rhs { a.Expr(r) }
}

func (a *Analyzer) SendStmt (s *ast.SendStmt) {
	a.Expr(s.Chan)
	a.HilAt(Separator, s.Arrow, 2)
	a.Expr(s.Value)
}

func (a *Analyzer) RangeStmt (s *ast.RangeStmt) {
	a.CurrLcl.With (func () {
		a.HilAt(Keyword, s.For, 3)
		a.DeclIdentExpr(s.Key)
		if s.Value != nil { a.DeclIdentExpr(s.Value) }
		a.hilAssn (s.Tok, s.TokPos)
		a.Expr(s.X)
		a.BlockStmt(s.Body)
	})
}

func (a *Analyzer) BlockStmt (e *ast.BlockStmt) {
	if e != nil {
		if e.Lbrace != 0 { a.HilAt(Token, e.Lbrace, 1) }
		a.CurrLcl.With (func () {
			for _, s := range e.List { a.Stmt(s) }
		})
		if e.Rbrace != 0 { a.HilAt(Token, e.Rbrace, 1) }
	}
}

func (a *Analyzer) ReturnStmt (s *ast.ReturnStmt) {
	a.HilAt(Keyword, s.Return, 6)
	for _, v := range s.Results { a.Expr(v) }
}

func (a *Analyzer) BranchStmt (s *ast.BranchStmt) {
	switch s.Tok {
	case token.      BREAK: a.HilAt(Keyword, s.TokPos, 5)
	case token.   CONTINUE: a.HilAt(Keyword, s.TokPos, 8)
	case token.       GOTO: a.HilAt(Keyword, s.TokPos, 4)
	case token.FALLTHROUGH: a.HilAt(Keyword, s.TokPos, 11)
	}
	if s.Label != nil { a.Hil(LblRef, s.Label) }
}

func (a *Analyzer) SelectStmt (s *ast.SelectStmt) {
	a.CurrLcl.With (func () {
		a.HilAt(Keyword, s.Select, 6)
		a.BlockStmt(s.Body)
	})
}

func (a *Analyzer) SwitchStmt (s *ast.SwitchStmt) {
	a.CurrLcl.With (func () {
		a.HilAt(Keyword, s.Switch, 6)
		if s.Init != nil { a.Stmt(s.Init) }
		if s.Tag  != nil { a.Expr(s.Tag ) }
		a.WithFlags (func () { a.Set (inSimpleSwitch); a.BlockStmt(s.Body)})
	})
}

func (a *Analyzer) TypeSwitchStmt (s *ast.TypeSwitchStmt) {
	a.CurrLcl.With (func () {
		a.HilAt(Keyword, s.Switch, 6)
		if s.Init != nil { a.Stmt(s.Init) }
		a.Stmt(s.Assign)
		a.WithFlags (func () { a.Set (inTypeSwitch); a.BlockStmt(s.Body)})
	})
}

func (a *Analyzer) CommClause (s *ast.CommClause) {
	if s.Comm == nil {
		a.HilAt(Keyword, s.Case, 7)
	} else {
		a.HilAt(Keyword, s.Case, 4)
		a.Stmt(s.Comm)
	}
	a.HilAt(Separator, s.Colon, 1)
	for _, s := range s.Body {
		a.Stmt(s)
	}
}

func (a *Analyzer) CaseClause (s *ast.CaseClause) {
	if s.List == nil {
		a.HilAt(Keyword, s.Case, 7)
	} else {
		a.HilAt(Keyword, s.Case, 4)
		for _, v := range s.List {
			if (a.Has (inTypeSwitch)) {
				a.Type(v)
			} else {
				a.Expr(v)
			}
		}
	}
	a.HilAt(Separator, s.Colon, 1)
	for _, s := range s.Body {
		a.WithFlags (func () { a.Clear(inTypeSwitch + inSimpleSwitch); a.Stmt(s) })
	}
}

func (a *Analyzer) ForStmt (s *ast.ForStmt) {
	a.CurrLcl.With (func () {
		a.HilAt(Keyword, s.For, 3)
		if s.Init != nil { a.Stmt(s.Init) }
		if s.Cond != nil { a.Expr(s.Cond) }
		if s.Post != nil { a.Stmt(s.Post) }
		a.BlockStmt(s.Body)
	})
}

func (a *Analyzer) GoStmt (s *ast.GoStmt) {
	a.HilAt(Keyword, s.Go, 2)
	a.CallExpr(s.Call)
}

func (a *Analyzer) DeferStmt (s *ast.DeferStmt) {
	a.HilAt(Keyword, s.Defer, 5)
	a.CallExpr(s.Call)
}

func (a *Analyzer) IfStmt (s *ast.IfStmt) {
	a.HilAt(Keyword, s.If, 2)
	if s.Init != nil { a.Stmt(s.Init) }
	a.Expr(s.Cond)
	a.BlockStmt(s.Body)
	if s.Else != nil { a.Stmt(s.Else) }
}

//===================================================================

func (a *Analyzer) Expr (e ast.Expr) env.Mode {
	if mode, ok := a.TryExpr(e); !ok {
		a.Error(e.Pos(), e.End(), "invalid expression node")
		return a.Type(e)
	} else {
		return mode
	}
}

func (a *Analyzer) TryExpr (e ast.Expr) (env.Mode, bool) {
	res := env.ModeNil
	ok := true
	a.WithNode (InExpr, e, func () {
		switch e := e.(type) {
		case *ast.       BadExpr: a.Hil(Error, e); res = a.ModeTab.Err
		case *ast.         Ident: res = a.     IdentExpr(e)
		case *ast.      CallExpr: res = a.      CallExpr(e)
		case *ast.      StarExpr: res = a.      StarExpr(e)
		case *ast.     IndexExpr: res = a.     IndexExpr(e)
		case *ast.     SliceExpr: res = a.     SliceExpr(e)
		case *ast.     ParenExpr: res = a.     ParenExpr(e)
		case *ast.     UnaryExpr: res = a.     UnaryExpr(e)
		case *ast.    BinaryExpr: res = a.    BinaryExpr(e)
		case *ast.  SelectorExpr: res = a.  SelectorExpr(e)
		case *ast.  KeyValueExpr: res = a.  KeyValueExpr(e)
		case *ast.TypeAssertExpr: res = a.TypeAssertExpr(e)
		case *ast.  CompositeLit: res = a.  CompositeLit(e)
		case *ast.      BasicLit: res = a.      BasicLit(e)
		case *ast.       FuncLit: res = a.       FuncLit(e)
		default: if e != nil { ok = false; }; res = a.ModeTab.Err
		}
	})
	return res, ok
}

func (a *Analyzer) IdentExpr (id *ast.Ident) env.Mode {
    if d, any := a.CurrSearch.Find (names.Put(id.Name)); !any && d == nil {
	a.NewError(id.Pos(), id.End(), "`" + id.Name + "' undefined")
	return a.ModeTab.Err
    } else if d == nil {
	a.Hil(VarRef, id)
	return a.ModeTab.Any
    } else {
	switch d.Kind {
 	case env.KVar    : a.Hil(VarRef, id)
	case env.KType   : a.Hil(TypRef, id)
	case env.KFunc   : a.Hil(FunRef, id)
	case env.KConst  : a.Hil(ConRef, id)
	case env.KPackage: a.Hil(PkgRef, id)
	default: a.Hil(Error, id)
	}
	//fmt.Printf ("id=%s dMode=%s\n", id.Name, d.Mode.Repr())
	return d.Mode
    }
}

func (a *Analyzer) CallExpr (e * ast.CallExpr) env.Mode {
	fMode := a.Expr(e.Fun)
	a.HilAt(Token, e.Lparen, 1)
	if fMode == a.ModeTab.Make {
	   if len (e.Args) == 0 {
     	       a.Error (e.Pos (), e.End (), "type argument required")
	   } else {
		for i, arg := range e.Args {
		    if i == 0 {
			a.Type(arg)
		    } else {
			a.Expr(arg)
		    }
		}
	   }
	} else {
	   for _, arg := range e.Args { a.Expr(arg) }
	}
	a.HilAt(Token, e.Rparen, 1)
	return a.ModeTab.Err
}

func (a *Analyzer) StarExpr (t *ast.StarExpr) env.Mode {
	a.HilAt(Operator, t.Star, 1)
	switch m := a.Expr(t.X).(type) {
	case *env.Ref: return m.Mode
	default:
		if !env.MAny (m) {
			a.Error (t.X.Pos (), t.X.End (), "pointer type required")
		}
		return a.ModeTab.Err
	}
}

func (a *Analyzer) IndexExpr (e *ast.IndexExpr) env.Mode {
	a.Expr(e.X)
	a.HilAt(Token, e.Lbrack, 1)
	a.Expr(e.Index)
	a.HilAt(Token, e.Rbrack, 1)
	return a.ModeTab.Err
}

func (a *Analyzer) SliceExpr (e *ast.SliceExpr) env.Mode {
	a.Expr(e.X)
	a.HilAt(Token, e.Lbrack, 1)
	if e.Low  != nil { a.Expr(e.Low ) }
	if e.High != nil { a.Expr(e.High) }
	if e.Max  != nil { a.Expr(e.Max ) }
	a.HilAt(Token, e.Rbrack, 1)
	return a.ModeTab.Err
}

func (a *Analyzer) ParenExpr (e *ast.ParenExpr) env.Mode {
	a.HilAt(Token, e.Lparen, 1)
	a.Expr(e.X)
	a.HilAt(Token, e.Rparen, 1)
	return a.ModeTab.Err
}

func (a *Analyzer) SelectorExpr (e *ast.SelectorExpr) env.Mode {
	m := a.Expr(e.X)
	res := env.ModeNil

	switch mm := m.(type) {
	case *env.IdType: m = env.DerefOpt (mm.Base)
	case *env.Ref   : m = env.DeId     (mm.Mode)
	}
	switch m := m.(type) {
	case *env.Struct:
		a.WithEnv(m.Flds, func () { res = a.Expr(e.Sel) })
	case *env.Builtin:
		if *m != env.KErr && *m != env.KAny {
			a.Error(e.X.Pos(), e.X.End(), "struct type required instead of '" + m.Head() + "'")
		}
		a.WithEnv(env.Any, func () { a.Expr(e.Sel) })
		res = a.ModeTab.Err
	default:
		a.Error(e.X.Pos(), e.X.End(), "struct type required instead of '" + m.Head () + "'")
		a.WithEnv(env.Any, func () { a.Expr(e.Sel) })
		res = a.ModeTab.Err
	}
	return res
}

func (a *Analyzer) KeyValueExpr (e *ast.KeyValueExpr) env.Mode {
	a.Expr(e.Key)
	a.HilAt(Separator, e.Colon, 1)
	a.Expr(e.Value)
	return a.ModeTab.Err
}

func (a *Analyzer) TypeAssertExpr (e *ast.TypeAssertExpr) env.Mode {
	a.Expr(e.X)
	a.HilAt(Token, e.Lparen, 1)
	a.Type(e.Type)
	a.HilAt(Token, e.Rparen, 1)
	return a.ModeTab.Err
}

func (a *Analyzer) UnaryExpr (e *ast.UnaryExpr) env.Mode {
	a.HilAt(Unary, e.OpPos, len(e.Op.String()))
	a.Expr(e.X)
	return a.ModeTab.Err
}

func (a *Analyzer) BinaryExpr (e *ast.BinaryExpr) env.Mode {
	a.Expr(e.X)
	a.HilAt(Binary, e.OpPos, len(e.Op.String()))
	a.Expr(e.Y)
	return a.ModeTab.Err
}

func (a *Analyzer) FuncLit (f *ast.FuncLit) env.Mode {
	a.FuncType (f.Type)
	a.BlockStmt(f.Body)
	return a.ModeTab.Err
}

func (a *Analyzer) BasicLit (c *ast.BasicLit) env.Mode {
	switch c.Kind {
	case token.STRING: a.Hil(String   , c); return a.ModeTab.String
	case token.  CHAR: a.Hil(Char     , c); return a.ModeTab.Int
	case token.   INT: a.Hil(Number   , c); return a.ModeTab.Int
	case token. FLOAT: a.Hil(Number   , c); return a.ModeTab.Float64
	case token.  IMAG: a.Hil(Number   , c); return a.ModeTab.Complex128
	}
	return a.ModeTab.Err
}

func (a *Analyzer) CompositeLit (e *ast.CompositeLit) env.Mode {
	a.Type(e.Type)
	a.HilAt(Token, e.Lbrace, 1)
	for _, elt := range e.Elts { a.Expr(elt) }
	a.HilAt(Token, e.Rbrace, 1)
	return a.ModeTab.Err
}

//===================================================================

func (a *Analyzer) Type (t ast.Expr) env.Mode {
	if m, ok := a.TryType(t); !ok {
		a.Error(t.Pos(), t.End(), "invalid type node")
		return a.Expr(t)
	} else {
		return m
	}
}

func (a *Analyzer) TryType (t ast.Expr) (env.Mode, bool) {
	res := env.ModeNil
	ok := true;
	a.WithNode (InType, t, func () {
		switch t := t.(type) {
		case *ast.        Ident: res = a.    IdentType(t)
		case *ast.      BadExpr: a.  Hil(Error,  t); res = a.ModeTab.Err
		case *ast.      MapType: res = a.      MapType(t)
		case *ast.     StarExpr: res = a.     StarType(t)
		case *ast.     FuncType: res = a.     FuncType(t)
		case *ast.     ChanType: res = a.     ChanType(t)
		case *ast.    ArrayType: res = a.    ArrayType(t)
		case *ast.   StructType: res = a.   StructType(t)
		case *ast.     Ellipsis: res = a. EllipsisType(t)
		case *ast.InterfaceType: res = a.InterfaceType(t)
		case *ast. SelectorExpr: res = a. SelectorType(t)
		default: ok = (t == nil); res = a.ModeTab.Err
		}
	})
	return res, ok
}

func (a *Analyzer) IdentType (id *ast.Ident) env.Mode {
    if d, any := a.CurrSearch.Find (names.Put(id.Name)); !any && d == nil {
	a.Error(id.Pos(), id.End(), "type `" + id.Name + "' undefined")
	return a.ModeTab.Err
    } else if d == nil {
	a.  Hil(TypRef, id)
	return a.ModeTab.Err
    } else {
	switch d.Kind {
 	case env.KVar    : a.Hil(VarRef, id)
	case env.KType   : a.Hil(TypRef, id)
	case env.KFunc   : a.Hil(FunRef, id)
	case env.KConst  : a.Hil(ConRef, id)
	case env.KPackage: a.Hil(PkgRef, id)
	default: a.Hil(Error, id)
	}
	if d.Kind != env.KType && d.Kind != env.KPackage {
		a.Error(id.Pos(), id.End(), "type ident required")
	}
	return d.Mode
    }
}

func (a *Analyzer) EllipsisType (t *ast.Ellipsis) env.Mode {
	a.HilAt(Token, t.Ellipsis, 3)
	a.Type(t.Elt)
	return a.ModeTab.Err
}

func (a *Analyzer) StarType (t *ast.StarExpr) env.Mode {
	a.HilAt(Operator, t.Star, 1)
	m := a.Type(t.X)
	return a.ModeTab.Ref (m)
}

func (a *Analyzer) ChanType (t *ast.ChanType) env.Mode {
	if t.Arrow == token.NoPos || t.Begin < t.Arrow {
	  a.HilAt(Keyword, t.Begin, 4)
	  if  t.Arrow != token.NoPos {
	    a.HilAt(Separator, t.Arrow, 2)
	  }
	} else {
	  a.HilAt(Separator, t.Begin, 2)
	}
	a.Type(t.Value)
	return a.ModeTab.Err
}

func (a *Analyzer) SelectorType (t *ast.SelectorExpr) env.Mode {
	a.Type(t.X)
	a.WithEnv(env.Any, func () { a.Type(t.Sel) })
	return a.ModeTab.Err

}

func (a *Analyzer) FuncType (t *ast.FuncType) env.Mode {
	a.HilAt(Keyword, t.Func, 4)
	a.FieldList(t.Params)
	a.FieldList(t.Results)
	return a.ModeTab.Err
}

func (a *Analyzer) StructType (t *ast.StructType) env.Mode {
	a.HilAt(Keyword, t.Struct, 6)
	bldr := env.NewBldr ()
	a.WithEnvDecl (bldr, func () {
		a.FieldList(t.Fields)
	})
	return a.ModeTab.Struct(bldr.Close ())
}

func (a *Analyzer) InterfaceType (t *ast.InterfaceType) env.Mode {
	a.HilAt(Keyword, t.Interface, 9)
	a.FieldList(t.Methods)
	return a.ModeTab.Err
}

func (a *Analyzer) ArrayType (t *ast.ArrayType) env.Mode {
	a.HilAt(Token, t.Lbrack, 1)
	var size *int = nil
	if t.Len != nil { /*sizeV := a.Expr(t.Len); size = &sizeV*/ } // TODO: size
	elem := a.Type(t.Elt)
	return a.ModeTab.Array (size, elem)
}

func (a *Analyzer) MapType (t *ast.MapType) env.Mode {
	a.HilAt(Keyword, t.Map, 3)
	a.Type(t.Key)
	a.Type(t.Value)
	return a.ModeTab.Err
}

//===================================================================

func (a *Analyzer) Declare (k env.Kind, n *names.Name, t ast.Expr, m env.Mode, v ast.Node) {
	m = a.ModeTab.Find (m);
	if n.Gbl () {
		a.CurrGbl.Declare(k, n, t, m, v)
	} else {
		a.CurrLcl.Declare(k, n, t, m, v)
	}
}

func (a *Analyzer) Decl (d ast.Decl) {
	a.WithNode (InDecl, d, func () {
		switch d := d.(type) {
		case *ast. BadDecl: a.Hil(Error, d)
		case *ast. GenDecl: a. GenDecl(d)
		case *ast.FuncDecl: a.FuncDecl(d)
		default: if d != nil { a.Error(d.Pos(), d.End(), "unknown decl node") }
		}
	})
}

func (a *Analyzer) FuncDecl (d *ast.FuncDecl) {
	// make types
	//a.FuncType(d.Type)
	//if d.Recv != nil { a.FieldList(d.Recv) }
	a.Hil(FunDef, d.Name)
	a.Declare(env.KFunc, names.Put(d.Name.Name), d.Type, a.ModeTab.Any, d)
}

func (a *Analyzer) GenDecl (d *ast.GenDecl) {
	switch d.Tok {
	case token.IMPORT: a.HilAt(Keyword, d.Pos(), 6)
	case token.CONST : a.HilAt(Keyword, d.Pos(), 5)
	case token.TYPE  : a.HilAt(Keyword, d.Pos(), 4)
	case token.VAR   : a.HilAt(Keyword, d.Pos(), 3)
	}

	if d.Lparen != 0 { a.HilAt(Token, d.Lparen, 1) }

	for _, s := range d.Specs {
		switch s := s.(type) {
		case *ast.ImportSpec: a.ImportSpec(s)
		case *ast. ValueSpec: a. ValueSpec(s)
		case *ast.  TypeSpec: a.  TypeSpec(s)
		default: if s != nil { a.Error(s.Pos(), s.End(), "unknown spec node") }
		}
	}

	if d.Rparen != 0 { a.HilAt(Token, d.Rparen, 1) }
}

func (a *Analyzer) Field (f *ast.Field) {
	m := a.Type (f.Type)
	for _, id := range f.Names {
		a.CurrLcl.Declare(env.KVar, names.Put(id.Name), f.Type, m, nil)
		//fmt.Println (id)
		a.Hil (VarDef, id)
	}
	//fmt.Println(f.Type)
	//ast.Fprint(os.Stdout, nil, f, nil)
	if f.Tag != nil { a.BasicLit (f.Tag) }
	if f.Comment != nil { a.Hil (Comment, f.Comment) }
}

func (a *Analyzer) FieldList (fl *ast.FieldList) {
	if fl != nil {
		if fl.Opening != 0 { a.HilAt(Token, fl.Opening, 1) }
		for _, f := range fl.List {
			a.Field(f)
		}
		if fl.Closing != 0 { a.HilAt(Token, fl.Closing, 1) }
	}
}

//===================================================================

func (a *Analyzer) ImportSpec (imp *ast.ImportSpec) {
	if imp.Name != nil {
		a.Hil (PkgDef, imp.Name)
		a.Declare(env.KPackage, names.Put(imp.Name.Name), nil, a.ModeTab.Any, imp)
	} else {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			a.Error(imp.Path.Pos(), imp.Path.End(), "invalid path syntax")
		} else {
			a.Declare(env.KPackage, names.Put (path.Base(p)), nil, a.ModeTab.Any, imp) // TODO: get file name
		}
	}
	a.BasicLit(imp.Path)
	if imp.Comment != nil { a.Hil (Comment, imp.Comment) }
}

func (a *Analyzer) ValueSpec (v *ast.ValueSpec) {
	for i, id := range v.Names {
		a.Hil (ValDef, id)
		if v.Values == nil {
			a.Declare(env.KConst, names.Put(id.Name), v.Type, a.ModeTab.Any, nil)
		} else {
			a.Declare(env.KConst, names.Put(id.Name), v.Type, a.ModeTab.Any, v.Values[i])
		}
	}
	if v.Comment != nil { a.Hil (Comment, v.Comment) }
}

func (a *Analyzer) TypeSpec (t *ast.TypeSpec) {
	a.Hil (TypDef, t.Name)
	a.Declare(env.KType, names.Put(t.Name.Name), t.Type, a.ModeTab.Any, nil)

	if t.Comment != nil { a.Hil (Comment, t.Comment) }
}

//===================================================================

func (a *Analyzer) AnalyzeFileIntr (f *ast.File) {
	a.CurrGbl = env.NewBldr()
	a.CurrGbl.Nested (a.ModeTab.BEnv)
	a.CurrLcl = env.NewBldr()
	a.CurrSearch = a.CurrLcl
	a.HilAt (Keyword, f.Package, 7)
	a.Hil   (PkgDef , f.Name)
	a.Declare (env.KPackage, names.Put(f.Name.Name), nil, a.ModeTab.Any, nil)
	if f.Doc != nil { a.Hil (Comment, f.Doc) }
 	for _, c    := range f.Comments   { a.Hil   (Comment, c) }
	for _, decl := range f.Decls      { a.Decl (decl) }
	a.Gbl = a.CurrGbl.Close()
	a.CurrLcl.Nested (a.Gbl)
	a.Lcl = a.CurrLcl.Close()
}

//-----------------------------------------------------------------

func (a *Analyzer) AnalyzeFileBody (f *ast.File) {
	a.CurrLcl = env.NewBldr ()
	a.CurrGbl = a.CurrLcl
	a.CurrGbl.Nested(a.Lcl)
	a.CurrSearch = a.CurrLcl

	a.CurrLcl.Scan(func (d *env.Decl) bool {
		if d.Kind == env.KType {
		        //fmt.Printf ("d.Name=%v, mode=%s\n", d.Name, d.Mode.Repr ())
			d.Mode = a.ModeTab.IdType (d, a.Type (d.Type))
		}
		return true
	})

	scan := func (ftr func (d *env.Decl) bool) {

		a.CurrLcl.Scan (func (d *env.Decl) bool {
			if (!ftr (d)) { return  true }
			m := d.Mode
			if d.Type != nil { m = a.Type(d.Type) }
			if d.Kind == env.KType {
				m = a.ModeTab.IdType (d, m)
			}
			switch n := d.Value.(type) {
			case *ast.FuncDecl:
				a.FuncType(n.Type)
				if n.Recv != nil { a.FieldList(n.Recv) }
				a.BlockStmt(n.Body)
			case  ast.Expr:
				em := a.Expr (n)
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

	//for _, decl := range f.Decls      { a.Decl (decl) }
 	//for _, id   := range f.Unresolved { a.Hil   (Error, id)  }
}

//===================================================================

func (a *Analyzer) SetOuterErrors (e scanner.ErrorList) {
	for _, err := range e {
		length := 1
		prefix := "expected '.', found 'IDENT' "
		if res, _ := regexp.MatchString (prefix + "[_A-Za-z]+", err.Msg); res {
			length = len (err.Msg) - len (prefix)
		}
		pos := a.file.Pos(err.Pos.Offset)
		a.Error (pos, token.Pos (int (pos)+length), err.Msg)
	}
}

//===================================================================

func (a *Analyzer) SetTokenFile (file *token.File) {
	a.file = file
}

//===================================================================

func (a *Analyzer) Analyze (file *token.File, f *ast.File) {
	a.SetTokenFile(file)
	a.AnalyzeFileIntr(f)
	a.AnalyzeFileBody(f)
}

//===================================================================

func (a *Analyzer) GetErrors (fname string, no int) *results.Errors {
	return &results.Errors {fname, no, a.errs}
}

func (a *Analyzer) GetFonts (fname string, bname string, no int) *results.Fontify {
	return &results.Fontify{fname, bname, no, 0, a.file.Size(), a.hils}
}
