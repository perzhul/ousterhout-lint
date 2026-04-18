// Package passthrough flags parameters that flow into a deeper call without
// branching, transformation, or inspection — a shape that signals the
// layer adds no abstraction value over the downstream function.
package passthrough

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/perzhul/ousterhout-lint/internal/astutil"
)

const analyzerName = "passthrough"

var Analyzer = &analysis.Analyzer{
	Name:     analyzerName,
	Doc:      "reports function parameters that are forwarded to a deeper call without inspection",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if fn.Body == nil || len(fn.Body.List) < 2 {
			// Single-statement bodies are shallowmethod's job.
			return
		}
		if astutil.HasNolintComment(fn.Doc, analyzerName) {
			return
		}
		if fn.Type.Params == nil {
			return
		}
		for _, field := range fn.Type.Params.List {
			if isExemptType(pass, field.Type) {
				continue
			}
			for _, name := range field.Names {
				if name.Name == "" || name.Name == "_" {
					continue
				}
				obj := pass.TypesInfo.Defs[name]
				if obj == nil {
					continue
				}
				if call, ok := singleDirectCallArgUsage(pass, fn.Body, obj); ok {
					pass.Reportf(name.Pos(),
						"passthrough: parameter %q is forwarded to %s without inspection; consider whether this layer adds value",
						name.Name, astutil.RenderCallee(pass.Fset, call))
				}
			}
		}
	})
	return nil, nil
}

// Cross-package forwarding is the conventional adapter pattern — excluded
// here to keep the signal high.
func singleDirectCallArgUsage(pass *analysis.Pass, body *ast.BlockStmt, obj types.Object) (*ast.CallExpr, bool) {
	var totalRefs int
	var callArg *ast.CallExpr
	var stack []ast.Node

	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if id, ok := n.(*ast.Ident); ok {
			if pass.TypesInfo.ObjectOf(id) == obj {
				totalRefs++
				if len(stack) > 0 {
					if call, ok := stack[len(stack)-1].(*ast.CallExpr); ok && isDirectArg(call, id) {
						callArg = call
					}
				}
			}
		}
		stack = append(stack, n)
		return true
	})

	if totalRefs != 1 || callArg == nil {
		return nil, false
	}
	if !isLocalCall(pass, callArg) {
		return nil, false
	}
	return callArg, true
}

func isLocalCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	var obj types.Object
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		obj = pass.TypesInfo.Uses[fn]
	case *ast.SelectorExpr:
		obj = pass.TypesInfo.Uses[fn.Sel]
	}
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg() == pass.Pkg
}

func isDirectArg(call *ast.CallExpr, id *ast.Ident) bool {
	for _, a := range call.Args {
		if a == id {
			return true
		}
	}
	return false
}

// Conventional plumbing types — flagging them as pass-through is noise.
var exemptTypes = map[string]bool{
	"context.Context": true,
	"testing.T":       true,
	"testing.B":       true,
	"testing.F":       true,
	"testing.TB":      true,
}

// Unwraps pointer types so *testing.T matches the "testing.T" entry.
func isExemptType(pass *analysis.Pass, expr ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(expr)
	if t == nil {
		return false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return exemptTypes[obj.Pkg().Path()+"."+obj.Name()]
}
