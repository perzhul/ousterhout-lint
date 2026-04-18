// Package shallowmethod reports methods whose body is a trivial pass-through
// of the outer parameters. Ousterhout's central thesis is that modules should
// be deep — a simple interface hiding significant complexity. A method that
// only forwards its arguments to another call adds call-stack depth without
// adding abstraction value, and is a strong signal that the abstraction
// boundary is in the wrong place.
package shallowmethod

import (
	"go/ast"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/perzhul/ousterhout-lint/internal/astutil"
)

const analyzerName = "shallowmethod"

// Analyzer is the shallowmethod pass — see the package doc for the rule it
// enforces.
var Analyzer = &analysis.Analyzer{
	Name:     analyzerName,
	Doc:      "reports methods whose body is a trivial pass-through of the outer parameters",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// implementsMarker matches a doc-comment line whose text is, optionally
// prefixed by the method name, "implements <InterfaceName>" — e.g.
//
//	// GetUser implements Fetcher.
//	// implements Fetcher
//
// The anchor and single-word slot before "implements" prevent accidental
// suppression from sentences that merely contain the word ("this layer
// doesn't implement anything useful").
var implementsMarker = regexp.MustCompile(`^//\s*(?:\w+\s+)?implements\s+\w+`)

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		fn := n.(*ast.FuncDecl)
		if !isCandidate(fn) {
			return
		}
		if astutil.HasNolintComment(fn.Doc, analyzerName) {
			return
		}
		if hasImplementsMarker(fn.Doc) {
			return
		}
		if fn.Recv != nil && implementsAnyInterface(pass, fn) {
			return
		}
		call, ok := shallowPassThroughCall(pass.TypesInfo, fn)
		if !ok {
			return
		}
		pass.Reportf(fn.Pos(),
			"shallowmethod: %s is a trivial pass-through to %s; remove this layer or let it add real value",
			fn.Name.Name, astutil.RenderCallee(pass.Fset, call))
	})
	return nil, nil
}

func isCandidate(fn *ast.FuncDecl) bool {
	if fn.Name == nil || fn.Body == nil {
		return false
	}
	if astutil.ParamCount(fn.Type) == 0 {
		return false
	}
	if astutil.IsConstructor(fn.Name.Name) {
		return false
	}
	return true
}

// shallowPassThroughCall returns the inner call if fn's body is a single
// statement consisting of a return/expression wrapping a call whose every
// argument is an identifier that names one of fn's parameters.
//
// Builtins (len, cap, make) and type conversions (int(x), UserID(s)) are
// rejected because they aren't wrappers — they're the operation itself.
func shallowPassThroughCall(info *types.Info, fn *ast.FuncDecl) (*ast.CallExpr, bool) {
	if len(fn.Body.List) != 1 {
		return nil, false
	}
	var call *ast.CallExpr
	switch s := fn.Body.List[0].(type) {
	case *ast.ReturnStmt:
		if len(s.Results) != 1 {
			return nil, false
		}
		c, ok := s.Results[0].(*ast.CallExpr)
		if !ok {
			return nil, false
		}
		call = c
	case *ast.ExprStmt:
		c, ok := s.X.(*ast.CallExpr)
		if !ok {
			return nil, false
		}
		call = c
	default:
		return nil, false
	}

	if tv, ok := info.Types[call.Fun]; ok && tv.IsType() {
		return nil, false
	}
	if id, ok := call.Fun.(*ast.Ident); ok {
		if obj := info.Uses[id]; obj != nil {
			if _, isBuiltin := obj.(*types.Builtin); isBuiltin {
				return nil, false
			}
		}
	}

	paramObjs := map[types.Object]bool{}
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, ident := range field.Names {
				if ident.Name == "" || ident.Name == "_" {
					continue
				}
				if obj := info.Defs[ident]; obj != nil {
					paramObjs[obj] = true
				}
			}
		}
	}
	if len(paramObjs) == 0 {
		return nil, false
	}
	if len(call.Args) == 0 {
		return nil, false
	}
	for _, arg := range call.Args {
		id, ok := arg.(*ast.Ident)
		if !ok {
			return nil, false
		}
		if !paramObjs[info.ObjectOf(id)] {
			return nil, false
		}
	}
	return call, true
}

// implementsAnyInterface reports whether fn's receiver type satisfies any
// interface declared in the current package or in a directly-imported
// package, where that interface includes a method with fn's name. Adapter
// and bridge implementations of io.Writer, http.Handler, fmt.Stringer, and
// similar widely-used interfaces are legitimate thin wrappers.
func implementsAnyInterface(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	obj := pass.TypesInfo.Defs[fn.Name]
	if obj == nil {
		return false
	}
	fnObj, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	sig, ok := fnObj.Type().(*types.Signature)
	if !ok {
		return false
	}
	recv := sig.Recv()
	if recv == nil {
		return false
	}
	recvType := recv.Type()
	method := fn.Name.Name

	if pkgHasSatisfyingInterface(pass.Pkg, recvType, method) {
		return true
	}
	for _, imp := range pass.Pkg.Imports() {
		if pkgHasSatisfyingInterface(imp, recvType, method) {
			return true
		}
	}
	return false
}

func pkgHasSatisfyingInterface(pkg *types.Package, recvType types.Type, method string) bool {
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		iface, ok := tn.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}
		hasMethod := false
		for i := 0; i < iface.NumMethods(); i++ {
			if iface.Method(i).Name() == method {
				hasMethod = true
				break
			}
		}
		if !hasMethod {
			continue
		}
		if types.Implements(recvType, iface) {
			return true
		}
	}
	return false
}

func hasImplementsMarker(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if implementsMarker.MatchString(c.Text) {
			return true
		}
	}
	return false
}
