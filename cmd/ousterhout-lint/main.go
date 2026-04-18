// Command ousterhout-lint bundles the shallowmethod and passthrough
// analyzers into a single CLI.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/perzhul/ousterhout-lint/passes/passthrough"
	"github.com/perzhul/ousterhout-lint/passes/shallowmethod"
)

func main() {
	multichecker.Main(
		shallowmethod.Analyzer,
		passthrough.Analyzer,
	)
}
