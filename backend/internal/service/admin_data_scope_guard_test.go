package service

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminServiceFunctionsDeclareDataScopeGuard(t *testing.T) {
	var files []string
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		t.Fatalf("scan service files: %v", err)
	}
	var missing []string
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, file, src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, decl := range parsed.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if !isAdminScopedServiceFunction(fn) {
				continue
			}
			body := string(src[fset.Position(fn.Body.Pos()).Offset:fset.Position(fn.Body.End()).Offset])
			if adminFunctionHasDataScopeGuard(body) {
				continue
			}
			missing = append(missing, file+":"+fn.Name.Name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("admin service functions must use access data-scope helpers or visible-resource guards: %s", strings.Join(missing, ", "))
	}
}

func isAdminScopedServiceFunction(fn *ast.FuncDecl) bool {
	if !strings.Contains(fn.Name.Name, "ForAdminContext") {
		return false
	}
	if fn.Type == nil || fn.Type.Params == nil {
		return false
	}
	for _, field := range fn.Type.Params.List {
		if !isUintIdent(field.Type) {
			continue
		}
		for _, name := range field.Names {
			if name.Name == "adminID" {
				return true
			}
		}
	}
	return false
}

func isUintIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "uint"
}

func adminFunctionHasDataScopeGuard(body string) bool {
	markers := []string{
		"access.ScopedResourceQuery",
		"access.DataScopeFilter",
		"access.UserDataScopeFilter",
		"access.UserIDDataScopeFilter",
		"access.VisibleDeptIDs",
		"VisibleQueryContext",
		"ForAdminContext(ctx",
		"ensure",
		"scoped",
	}
	for _, marker := range markers {
		if strings.Contains(body, marker) {
			return true
		}
	}
	return false
}
