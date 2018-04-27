package env

//import "fmt"
import "go/ast"
import "github.com/kouzdra/go-analyzer/names"
import "github.com/kouzdra/go-analyzer/paths"


type Kind int

const (
        KType = Kind (iota)
        KFunc
        KVar
        KPackage
        KConst
)

type Decl struct {
        Kind  Kind
        Name *names.Name
	Path *paths.Path
        Type ast.Expr
        Mode  Mode
        Value ast.Node
}


type DeclList struct {
        Decl
        Next *DeclList
}


type Subs struct {
        Sub []*Env
        Any bool
}

type Env struct {
        Subs
        Decls []*DeclList
}

func (d *DeclList) Scan (fn func (d *Decl) bool) bool {
       for d != nil {
                if !fn (&d.Decl) { return false }
                d = d.Next
       }
       return true
}

func (e *Subs) Scan (fn func (d *Decl) bool) bool {
        for _, e := range e.Sub {
                if !e.Scan(fn) { return false }
        }
        return true
}

func (e *Env) Scan (fn func (d *Decl) bool) bool {
        for _, d := range e.Decls {
                if !d.Scan(fn) { return false }
        }
        return e.Subs.Scan(fn)
}

func (e *EnvBldr) Scan (fn func (d *Decl) bool) bool {
        return e.Decls.Scan (fn) && e.Subs.Scan(fn)
}


var Empty  = &Env{Subs{[]*Env{}, false}, []*DeclList{nil}}
var Any    = &Env{Subs{[]*Env{}, true }, []*DeclList{nil}}

type Mark struct {
    nSubs int
    decls *DeclList
}

type EnvBldr struct {
        Subs
        Decls *DeclList
}

func (e *EnvBldr) Mark () Mark {
        return Mark{len(e.Subs.Sub), e.Decls}
}

func (e *EnvBldr) Reset (m Mark) {
        e.Decls = m.decls
        //len(e.Subs.Sub) = m.nSubs
}

func (e *EnvBldr) Close () *Env {
        hsize := uint (1)
        for d := e.Decls; d != nil; d = d.Next { hsize ++ }
        res := &Env{e.Subs, make ([]*DeclList, hsize, hsize)}
        for d := e.Decls; d != nil; d = d.Next {
                h := d.Name.Hash % hsize
                res.Decls [h] = &DeclList{Decl{d.Kind, d.Name, d.Path, d.Type, d.Mode, d.Value}, res.Decls[h]}
        }
        return res
}

func NewBldr () *EnvBldr { return &EnvBldr{Subs{make([]*Env, 0, 40), false}, nil} }

func (e *EnvBldr) Declare (k Kind, n *names.Name, p *paths.Path, t ast.Expr, m Mode, v ast.Node) *Decl {
        if n == names.Dummy {
                return nil
        }
        if m == nil { panic ("nil mode decl") }
        e.Decls = &DeclList{Decl{k, n, p, t, m, v}, e.Decls}
        return &e.Decls.Decl
}

func (e *EnvBldr) Nested (s * Env) {
        e.Sub = append(e.Sub, s)
}

func (e *EnvBldr) With (fn func ()) {
        mark := e.Mark()
        fn ()
        e.Reset(mark)
}

func (e *Subs) Find (n *names.Name) (*Decl, bool) {
        //fmt.Printf("Sub.Find: %s %b\n", n.Name, e.Any)
        for _, s := range e.Sub {
                if d, any := s.Find(n); any || d != nil { return d, any }
        }
        return nil, e.Any
}

func (e *Env) Find (n *names.Name) (*Decl, bool) {
        //fmt.Printf("%#v name=%d\n", e.Decls, n.Hash)
        h := n.Hash % uint (len(e.Decls))
        for d := e.Decls[h]; d != nil; d = d.Next {
                if d.Name == n { return &d.Decl, false }
        }
        return e.Subs.Find(n)
}

func (e *EnvBldr) Find (n *names.Name) (*Decl, bool) {
        //fmt.Printf("+  name=%s\n", n.Name)
        for d := e.Decls; d != nil; d = d.Next {
                //fmt.Printf("- name=%s\n", d.Name.Name)
                if d.Name == n { return &d.Decl, false }
        }
        return e.Subs.Find(n)
}
