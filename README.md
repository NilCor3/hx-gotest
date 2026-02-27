# hx-gotest

A small Go CLI tool that finds the correct `go test -run` pattern at a given file/line using Go's AST parser. Used to run tests from the Helix editor via WezTerm.

## Usage

```sh
hx-gotest <file> <line> [cursor|func|file]
```

### Modes

| Mode | Description | Example output |
|------|-------------|----------------|
| `cursor` | Deepest test/subtest at cursor, falls back to whole func | `^TestFoo$/^subBar$` |
| `func` | Enclosing top-level `Test*` func (all subtests) | `^TestFoo$` |
| `file` | All `Test*` funcs in the file | `^(TestFoo\|TestBar)$` |

Exits with code 1 if cursor is not inside a test function.

## Helix keybindings (`Space , `)

| Key | Action |
|-----|--------|
| `t` | Run test under cursor |
| `T` | Run whole test func |
| `F` | Run all tests in file |

Output streams into a bottom WezTerm pane.

## Build & install

```sh
cd ~/source/hx-gotest
go build -o ~/.local/bin/hx-gotest .
```

Requires `~/.local/bin` on `$PATH` (added to `~/.zshrc`).

## Development

```sh
# Run all tests
go test -v ./...

# Run a specific test
go test -v -run TestFindTestPattern/table:_unkeyed_first_row

# Rebuild after changes
go build -o ~/.local/bin/hx-gotest .
```

The core logic is in `finder.go`. Add new test cases to `finder_test.go` in the
`testFile` constant (embedded Go source) and the `cases` table in `TestFindTestPattern`.

Note: line numbers in test cases refer to lines within the embedded `testFile` string,
not lines in `finder_test.go` itself (`testFile` starts at line 1 of the parsed temp file).

## Edge cases

- `t.Run` with variable name → uses `.*` wildcard at that level
- Deeply nested subtests → `TestA/SubB/SubC`
- Cursor on `t.Run(...)` line → matches that subtest
