package passthrough_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/perzhul/ousterhout-lint/passes/passthrough"
)

func TestPassthrough(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), passthrough.Analyzer, "p")
}
