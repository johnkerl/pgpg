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
	fmt.Fprintf(os.Stderr, "Usage: %s [-o output.go] [-package name] [-type name] tables.json\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	var outputPath string
	var packageName string
	var typeName string
	var debug bool
	flag.StringVar(&outputPath, "o", "", "Output Go file (default stdout)")
	flag.StringVar(&packageName, "package", "lexers", "Package name for generated lexer")
	flag.StringVar(&typeName, "type", "GeneratedLexer", "Lexer type name")
	flag.BoolVar(&debug, "debug", false, "Write unformatted code to stderr")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}
	tablesPath := flag.Arg(0)

	opts := lexgen.LexCodegenOptions{
		Package: packageName,
		Type:    typeName,
		Format:  true,
	}
	if debug {
		if err := dumpRaw(tablesPath, opts); err != nil {
			die(err)
		}
	}
	if err := run.LexgenCode(context.Background(), tablesPath, outputPath, opts); err != nil {
		die(err)
	}
}

// dumpRaw writes the unformatted generated code to stderr for debugging.
func dumpRaw(tablesPath string, opts lexgen.LexCodegenOptions) error {
	data, err := os.ReadFile(tablesPath)
	if err != nil {
		return err
	}
	tables, err := lexgen.DecodeTables(data)
	if err != nil {
		return err
	}
	opts.Format = false
	raw, err := lexgen.GenerateCode(tables, opts)
	if err != nil {
		return err
	}
	_, _ = os.Stderr.Write(raw)
	_, _ = os.Stderr.Write([]byte("\n"))
	return nil
}
