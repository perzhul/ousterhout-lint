// Command ousterhout-lint is a deep-module enforcer for Go, derived from
// John Ousterhout's "A Philosophy of Software Design". It bundles two
// analyzers — shallowmethod and passthrough — that catch the mechanically
// detectable shape of shallow abstraction: trivial wrapper methods and
// parameters that flow straight through without adding value.
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
