package httpboundary

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestProductionHTTPHandlersDoNotReturnRawErrors(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatalf("resolve backend root: %v", err)
	}
	fset := token.NewFileSet()
	violations := make([]string, 0)
	err = filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		slashPath := filepath.ToSlash(rel)
		if !strings.Contains(slashPath, "/handler/") &&
			!strings.Contains(slashPath, "/transport/") &&
			!strings.Contains(slashPath, "/routes/") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", rel, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || !isResponseCall(call.Fun) {
				return true
			}
			for _, argument := range call.Args {
				if containsRawError(argument) {
					position := fset.Position(argument.Pos())
					violations = append(violations, fmt.Sprintf("%s:%d", slashPath, position.Line))
					break
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan HTTP boundary: %v", err)
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("production HTTP responses expose raw errors:\n%s", strings.Join(violations, "\n"))
	}
}

func isResponseCall(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || (selector.Sel.Name != "Fail" && selector.Sel.Name != "JSON") {
		return false
	}
	identifier, ok := selector.X.(*ast.Ident)
	return ok && identifier.Name == "response"
}

func containsRawError(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Error" && len(call.Args) == 0 {
			found = true
			return false
		}
		if identifier, ok := call.Fun.(*ast.Ident); ok && identifier.Name == "string" && len(call.Args) == 1 {
			if errIdentifier, ok := call.Args[0].(*ast.Ident); ok && errIdentifier.Name == "err" {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
