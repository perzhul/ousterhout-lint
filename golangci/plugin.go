// Package main is the golangci-lint analyzer plugin entry point.
// Build with: go build -buildmode=plugin -o ousterhout.so ./golangci
package main

import (
	"golang.org/x/tools/go/analysis"

	"github.com/perzhul/ousterhout-lint/passes/shallowmethod"
)

type analyzerPlugin struct{}

func (analyzerPlugin) GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{shallowmethod.Analyzer}
}

// AnalyzerPlugin is the symbol golangci-lint looks up via plugin.Lookup.
// Renaming or removing it breaks downstream configs.
var AnalyzerPlugin analyzerPlugin

func main() {}
