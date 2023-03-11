// Run analyzer: myanalyzer [directory].
package main

import (
	"myanalyzer/os_exit_analyzer"
	"strings"

	"github.com/gostaticanalysis/elseless"
	"github.com/gostaticanalysis/nakedreturn"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
)

func main() {

	var myChecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// Checks all "SA" checks from staticcheck.
		if strings.Contains(v.Analyzer.Name, "SA") {
			myChecks = append(myChecks, v.Analyzer)
		}
	}

	for _, v := range simple.Analyzers {
		// Checks 'S1002' check from staticcheck.
		// Omit comparison with boolean constant.
		if v.Analyzer.Name == "S1002" {
			myChecks = append(myChecks, v.Analyzer)
			break
		}
	}

	// Checks printf.Analyzer, shadow.Analyzer, os_exit_analyzer.OsExitAnalyzer.
	myChecks = append(myChecks, printf.Analyzer, shadow.Analyzer, os_exit_analyzer.OsExitAnalyzer)

	// Nakedreturn analyzer. Checks if there are naked returns in code.
	myChecks = append(myChecks, nakedreturn.Analyzer)

	// Elseless finds unnecessary else.
	myChecks = append(myChecks, elseless.Analyzer)

	multichecker.Main(myChecks...)
}
