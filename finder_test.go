package main

import (
	"os"
	"path/filepath"
	"testing"
)

const testFile = `package mypackage

import "testing"

func TestSimple(t *testing.T) {
	t.Log("hello")
}

func TestWithSub(t *testing.T) {
	t.Run("subA", func(t *testing.T) {
		t.Log("in subA")
	})
	t.Run("subB", func(t *testing.T) {
		t.Log("in subB")
	})
}

func TestNested(t *testing.T) {
	t.Run("outer", func(t *testing.T) {
		t.Run("inner", func(t *testing.T) {
			t.Log("deep")
		})
	})
}

func TestVarSubtest(t *testing.T) {
	name := "dynamic"
	t.Run(name, func(t *testing.T) {
		t.Log("var")
	})
}
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	f := filepath.Join(dir, "example_test.go")
	if err := os.WriteFile(f, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFindTestPattern(t *testing.T) {
	file := writeTemp(t, testFile)

	cases := []struct {
		name    string
		line    int
		mode    string
		want    string
		wantOK  bool
	}{
		// cursor mode
		{"cursor: top-level func body", 6, "cursor", `^TestSimple$`, true},
		{"cursor: on func signature", 5, "cursor", `^TestSimple$`, true},
		{"cursor: in subA", 11, "cursor", `^TestWithSub$/^subA$`, true},
		{"cursor: in subB", 14, "cursor", `^TestWithSub$/^subB$`, true},
		{"cursor: on t.Run line (subA)", 10, "cursor", `^TestWithSub$/^subA$`, true},
		{"cursor: nested inner", 22, "cursor", `^TestNested$/^outer$/^inner$`, true},
		{"cursor: nested outer (on outer t.Run line)", 19, "cursor", `^TestNested$/^outer$`, true},
		{"cursor: variable subtest", 30, "cursor", `^TestVarSubtest$/.*`, true},
		{"cursor: outside any test", 1, "cursor", ``, false},

		// func mode
		{"func: ignores subtests", 11, "func", `^TestWithSub$`, true},
		{"func: simple", 6, "func", `^TestSimple$`, true},

		// file mode
		{"file: all tests", 1, "file", `^(TestSimple|TestWithSub|TestNested|TestVarSubtest)$`, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := FindTestPattern(file, tc.line, tc.mode)
			if ok != tc.wantOK {
				t.Errorf("ok = %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Errorf("pattern = %q, want %q", got, tc.want)
			}
		})
	}
}
