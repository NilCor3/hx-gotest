package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// knownNameFields are struct field names commonly used as table-driven test names.
var knownNameFields = map[string]bool{
	"name": true, "testName": true, "desc": true, "description": true,
	"scenario": true, "test": true, "caseName": true, "label": true,
}

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

		// cursor mode: find deepest t.Run subtest at this line
		subPath := findSubtest(fset, fn.Body, line)
		if subPath != "" {
			return "^" + fn.Name.Name + "$/" + subPath, true
		}

		// Fallback: table-driven test — find struct member named like a test name
		if hint := tableTestNameOnLine(fset, f, line); hint != "" {
			return "^" + fn.Name.Name + "$/" + "^" + hint + "$", true
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

// tableTestNameOnLine finds a struct composite literal element on the given line
// and returns the value of its name-like field (keyed or first string).
// Spaces are replaced with "_" to match Go's t.Run sanitization.
func tableTestNameOnLine(fset *token.FileSet, f *ast.File, line int) string {
	var result string

	ast.Inspect(f, func(n ast.Node) bool {
		if result != "" {
			return false
		}
		outer, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		for _, elt := range outer.Elts {
			eltStart := fset.Position(elt.Pos()).Line
			eltEnd := fset.Position(elt.End()).Line
			if line < eltStart || line > eltEnd {
				continue
			}

			// Element is a struct composite literal (the row)
			row, ok := elt.(*ast.CompositeLit)
			if !ok {
				// Could be a single-value element — check for string literal
				if bl, ok := elt.(*ast.BasicLit); ok && bl.Kind == token.STRING {
					result = testNameFromLit(bl)
				}
				return false
			}

			// First pass: look for keyed name field
			for _, field := range row.Elts {
				kv, ok := field.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok || !knownNameFields[key.Name] {
					continue
				}
				if bl, ok := kv.Value.(*ast.BasicLit); ok && bl.Kind == token.STRING {
					result = testNameFromLit(bl)
					return false
				}
			}

			// Second pass: use first string BasicLit in unkeyed struct
			for _, field := range row.Elts {
				if bl, ok := field.(*ast.BasicLit); ok && bl.Kind == token.STRING {
					result = testNameFromLit(bl)
					return false
				}
			}

			return false
		}
		return true
	})

	return result
}

// testNameFromLit strips quotes and replaces spaces with _ (Go's t.Run sanitization).
func testNameFromLit(bl *ast.BasicLit) string {
	v := strings.Trim(bl.Value, "`\"")
	return strings.ReplaceAll(v, " ", "_")
}
