// Package main provides the golangci-lint "analyzer plugin" entry point.
//
// Build with:
//
//	go build -buildmode=plugin -o ousterhout.so ./golangci
//
// Then reference the resulting .so in your .golangci.yml under
// `linters-settings.custom`.
package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/perzhul/ousterhout-lint/passes/passthrough"
	"github.com/perzhul/ousterhout-lint/passes/shallowmethod"
)

// analyzerPlugin holds the set of analyzers exposed to golangci-lint via
// the legacy plugin loader.
type analyzerPlugin struct{}

// GetAnalyzers returns the analyzers loaded when this package is opened as
// a golangci-lint custom plugin.
func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		shallowmethod.Analyzer,
		passthrough.Analyzer,
	}
}

// AnalyzerPlugin is the symbol golangci-lint looks up via plugin.Lookup
// when loading a custom linter. Renaming or removing it breaks downstream
// golangci-lint configurations.
var AnalyzerPlugin analyzerPlugin

// main is required so this package compiles as package main. The plugin is
// loaded via -buildmode=plugin, which picks up AnalyzerPlugin, not main.
func main() {}
