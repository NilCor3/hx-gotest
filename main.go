package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: hx-gotest <file> <line> [cursor|func|file]")
		os.Exit(1)
	}

	file := os.Args[1]
	line, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid line number: %s\n", os.Args[2])
		os.Exit(1)
	}

	mode := "cursor"
	if len(os.Args) >= 4 {
		mode = os.Args[3]
	}

	pattern, ok := FindTestPattern(file, line, mode)
	if !ok {
		fmt.Fprintln(os.Stderr, "cursor is not inside a test function")
		os.Exit(1)
	}
	if pattern == "" {
		fmt.Fprintln(os.Stderr, "no test functions found")
		os.Exit(1)
	}

	fmt.Println(pattern)
}
