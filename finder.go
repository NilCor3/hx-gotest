package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// FindTestPattern returns the go test -run pattern for the given file/line/mode.
// mode is one of: cursor, func, file.
// Returns ("", false) if the cursor is not inside a test function (for cursor/func modes).
func FindTestPattern(filename string, line int, mode string) (string, bool) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return "", false
	}

	if mode == "file" {
		return filePattern(fset, f), true
	}

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			continue
		}
		if !strings.HasPrefix(fn.Name.Name, "Test") {
			continue
		}
		start := fset.Position(fn.Pos()).Line
		end := fset.Position(fn.End()).Line
		if line < start || line > end {
			continue
		}

		if mode == "func" {
			return "^" + fn.Name.Name + "$", true
		}

		// cursor mode: find deepest subtest at this line
		subPath := findSubtest(fset, fn.Body, line)
		if subPath != "" {
			return "^" + fn.Name.Name + "$/" + subPath, true
		}
		return "^" + fn.Name.Name + "$", true
	}

	return "", false
}

// findSubtest recursively searches for t.Run/b.Run calls containing the line.
// Returns the joined subtest path e.g. "^SubA$/^SubB$", or "" if none found.
func findSubtest(fset *token.FileSet, node ast.Node, line int) string {
	var result string

	ast.Inspect(node, func(n ast.Node) bool {
		if result != "" {
			return false // already found deepest
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Match t.Run(...) or b.Run(...)
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Run" {
			return true
		}
		if len(call.Args) < 2 {
			return true
		}

		start := fset.Position(call.Pos()).Line
		end := fset.Position(call.End()).Line
		if line < start || line > end {
			return true
		}

		// Get subtest name
		var name string
		if lit, ok := call.Args[0].(*ast.BasicLit); ok {
			// String literal: strip quotes
			name = "^" + strings.Trim(lit.Value, `"`) + "$"
		} else {
			// Variable or expression: use wildcard
			name = ".*"
		}

		// Recurse into the function body for deeper nesting
		if fn, ok := call.Args[1].(*ast.FuncLit); ok {
			deeper := findSubtest(fset, fn.Body, line)
			if deeper != "" {
				result = name + "/" + deeper
				return false
			}
		}

		result = name
		return false
	})

	return result
}

// filePattern returns a -run pattern matching all Test* functions in the file.
func filePattern(fset *token.FileSet, f *ast.File) string {
	var names []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			continue
		}
		if strings.HasPrefix(fn.Name.Name, "Test") {
			names = append(names, fn.Name.Name)
		}
	}
	if len(names) == 0 {
		return ""
	}
	return "^(" + strings.Join(names, "|") + ")$"
}
