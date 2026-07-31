package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/johnkerl/pgpg/go/generators/pkg/lexgen"
	"github.com/johnkerl/pgpg/go/generators/pkg/run"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-o output.json] input.bnf\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var outputPath string
	var nosort bool
	flag.StringVar(&outputPath, "o", "", "Output JSON file (default stdout)")
	flag.BoolVar(&nosort, "nosort", false, "Skip sorting for faster output with nondeterministic ordering")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}

	opts := &run.LexgenTablesOptions{
		Encode: &lexgen.EncodeOptions{Sort: !nosort},
	}
	if err := run.LexgenTables(context.Background(), flag.Arg(0), outputPath, opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
