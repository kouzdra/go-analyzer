package env

import "fmt"

func (e *EnvBldr) Print () {
  for d := e.Decls; d != nil; d = d.Next {
        switch d.Kind {
        case KVar    : fmt.Printf("KVar: %s\n"     , d.Name.Name)
        case KType   : fmt.Printf("KType: %s\n"    , d.Name.Name)
        case KFunc   : fmt.Printf("KFunc: %s\n "   , d.Name.Name)
        case KConst  : fmt.Printf("KConst: %s\n"   , d.Name.Name)
        case KPackage: fmt.Printf("KPackage: %s\n" , d.Name.Name)
        }
  }
}
