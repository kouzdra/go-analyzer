package analyzer

import "go/ast"

type IProcessor interface {
        Before (n ast.Node) bool
        After  (n ast.Node) bool
}

type Processor struct {
        Analyzer
}

func (p *Processor) Before (n ast.Node) bool { return true }
func (p *Processor) After  (n ast.Node) bool { return true }
