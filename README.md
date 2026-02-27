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
go build -o ~/.local/bin/hx-gotest .
```

## Edge cases

- `t.Run` with variable name → uses `.*` wildcard at that level
- Deeply nested subtests → `TestA/SubB/SubC`
- Cursor on `t.Run(...)` line → matches that subtest
