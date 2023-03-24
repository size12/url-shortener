// Package os_exit_analyzer is an analyzer that check that there are not os.Exit in main function.
package os_exit_analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer analyzes is there are os.Exit() func in main package main func.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osExit",
	Doc:  "check for os exit in main function",
	Run:  run,
}

// run runs analyzer.Analyzer implementation of OsExitAnalyzer
func run(pass *analysis.Pass) (interface{}, error) {

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				// If package isn't main, skip it.
				if x.Name.String() != "main" {
					return false
				}
			case *ast.FuncDecl:
				// If func name isn't main, skip it.
				if x.Name.String() != "main" {
					return false
				}
			case *ast.CallExpr:
				{
					if fun, ok := x.Fun.(*ast.SelectorExpr); ok {
						// If func name isn't 'Exit', skip it.
						if fun.Sel.String() != "Exit" {
							return true
						}

						name, ok := fun.X.(*ast.Ident)

						if !ok {
							return true
						}

						// If package name is "os", report error.
						if name.String() == "os" {
							pass.Reportf(name.Pos(), "you shouldn't use os.Exit in function main")
						}
					}
				}
			}
			return true
		})
	}

	return nil, nil
}
