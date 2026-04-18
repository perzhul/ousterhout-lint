// Package astutil provides small AST helpers shared across passes.
package astutil

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

// IsConstructor reports whether the function name follows the New/NewXxx
// convention. The rune after "New" must be an uppercase letter — so that
// Newline, Newer, NewsFeed are *not* considered constructors.
func IsConstructor(name string) bool {
	if name == "New" {
		return true
	}
	if !strings.HasPrefix(name, "New") {
		return false
	}
	rest := name[len("New"):]
	r, _ := utf8.DecodeRuneInString(rest)
	return unicode.IsUpper(r)
}

// ParamCount returns the total parameter count of a FuncType, expanding groups.
func ParamCount(ft *ast.FuncType) int {
	if ft == nil || ft.Params == nil {
		return 0
	}
	n := 0
	for _, field := range ft.Params.List {
		if len(field.Names) == 0 {
			n++
			continue
		}
		n += len(field.Names)
	}
	return n
}

// RenderCallee returns the textual form of the callee in call, suffixed
// with "(...)". Used in diagnostic messages to point at the wrapped call
// without dumping its arguments. Returns a stable fallback if printing
// fails.
func RenderCallee(fset *token.FileSet, call *ast.CallExpr) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, call.Fun); err != nil {
		return "the wrapped function"
	}
	return fmt.Sprintf("%s(...)", buf.String())
}

// HasNolintComment reports whether any comment in cg contains //nolint:<name>.
func HasNolintComment(cg *ast.CommentGroup, name string) bool {
	if cg == nil {
		return false
	}
	needle := "nolint:" + name
	for _, c := range cg.List {
		if strings.Contains(c.Text, needle) {
			return true
		}
	}
	return false
}
