package myanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitExistsInMain is my analyzer for multichecker.
var OsExitExistsInMain = &analysis.Analyzer{
	Name: "osexitexistsinmain",
	Doc:  "check for os.Exit in the main file",
	Run:  run,
}

// run - analysis of code, checks for os.Exit in the main file.
func run(pass *analysis.Pass) (interface{}, error) {
	// iterate .go files.
	for _, file := range pass.Files {
		// iterate over all AST nodes.
		ast.Inspect(file, func(n ast.Node) bool {
			// look for main function.
			if v, ok := n.(*ast.FuncDecl); ok && v.Name.Name == `main` {
				// iterate over AST nodes in main.
				for _, stmt := range v.Body.List {
					// look for expression statements.
					if ex, ok := stmt.(*ast.ExprStmt); ok {
						// look for function call expressions.
						if call, ok := ex.X.(*ast.CallExpr); ok {
							// look for selector expressions.
							if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
								// checking the selector for equality with 'os' and then whit 'Exit'.
								if i, ok := selector.X.(*ast.Ident); ok && i.Name == `os` {
									if selector.Sel.Name == `Exit` {
										pass.Reportf(selector.Pos(), "os.Exit exists in main body")
									}
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
