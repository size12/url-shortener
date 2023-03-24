package os_exit_analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestMyAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), OsExitAnalyzer, "./...")
}
