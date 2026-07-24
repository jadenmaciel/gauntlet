package crap

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// ComplexityOfFile parses one Go source file (given as bytes, using a
// virtual filename so no disk I/O is required) and returns the
// cyclomatic complexity of every top-level function and method.
//
// Complexity counting follows the standard gocyclo algorithm:
//   - starts at 1 per function/method declaration
//   - +1 for each if, for (including range), case/comm clause, && and ||
//   - else alone and the switch/select keyword itself do not add
func ComplexityOfFile(src []byte) ([]FunctionComplexity, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		return nil, err
	}

	var results []FunctionComplexity
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		pos := fset.Position(fn.Pos())
		results = append(results, FunctionComplexity{
			File:       pos.Filename,
			Line:       pos.Line,
			Func:       fn.Name.Name,
			Complexity: complexityOf(fn.Body),
		})
	}
	return results, nil
}

func complexityOf(body ast.Node) int {
	complexity := 1
	ast.Inspect(body, func(n ast.Node) bool {
		complexity += decisionComplexity(n)
		return true
	})
	return complexity
}

func decisionComplexity(n ast.Node) int {
	switch node := n.(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
		return 1
	case *ast.CaseClause:
		return boolToInt(node.List != nil)
	case *ast.CommClause:
		return boolToInt(node.Comm != nil)
	case *ast.BinaryExpr:
		return boolToInt(node.Op == token.LAND || node.Op == token.LOR)
	default:
		return 0
	}
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
