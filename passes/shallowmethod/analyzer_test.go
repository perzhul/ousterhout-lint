package shallowmethod_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/perzhul/ousterhout-lint/passes/shallowmethod"
)

func TestShallowMethod(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), shallowmethod.Analyzer, "p")
}
