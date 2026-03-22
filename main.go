package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: compiler <source.pas>")

		os.Exit(1)
	}

	src, err := os.ReadFile(os.Args[1])

	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read '%s': %v\n", os.Args[1], err)

		os.Exit(1)
	}

	scn = initScanner(src)

	next()
	
	initCodegen()

	module()

	if scn.errcnt != 0 {
		fmt.Fprintf(os.Stderr, "%d error(s), compilation failed\n", scn.errcnt)

		os.Exit(1)
	}

	base := os.Args[1]

	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}

	close(base)
}