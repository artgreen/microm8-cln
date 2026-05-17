//go:build !remint
// +build !remint

package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	gotest "testing"
)

func TestNoUpdatesFollowsNoUpdateFlagOnly(t *gotest.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", nil, 0)
	if err != nil {
		t.Fatalf("parse main.go: %v", err)
	}

	var rhs string
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}

		for i, lhs := range assign.Lhs {
			if !isSettingsNoUpdates(lhs) || i >= len(assign.Rhs) {
				continue
			}
			rhs = nodeSource(t, fset, assign.Rhs[i])
		}

		return true
	})

	if rhs == "" {
		t.Fatal("settings.NoUpdates assignment not found")
	}
	if strings.Contains(rhs, "mcpMode") {
		t.Fatalf("settings.NoUpdates must not depend on mcpMode; got %q", rhs)
	}
	if rhs != "*noUpdate" {
		t.Fatalf("settings.NoUpdates should be assigned from *noUpdate only; got %q", rhs)
	}
}

func isSettingsNoUpdates(expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "NoUpdates" {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "settings"
}

func nodeSource(t *gotest.T, fset *token.FileSet, node ast.Node) string {
	t.Helper()

	var out strings.Builder
	if err := format.Node(&out, fset, node); err != nil {
		t.Fatalf("format assignment expression: %v", err)
	}
	return out.String()
}
