// Command ousterhout-lint runs the shallowmethod analyzer.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/perzhul/ousterhout-lint/passes/shallowmethod"
)

func main() {
	multichecker.Main(shallowmethod.Analyzer)
}
