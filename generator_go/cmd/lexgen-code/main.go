package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/johnkerl/pgpg/generator_go/pkg/lexgen"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-o output.go] [-package name] [-type name] tables.json\n", os.Args[0])
	flag.PrintDefaults()
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
	inputPath := flag.Arg(0)

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tables, err := lexgen.DecodeTables(inputBytes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if debug {
		raw, err := lexgen.GenerateGoLexerCodeRaw(tables, packageName, typeName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		_, _ = os.Stderr.Write(raw)
		_, _ = os.Stderr.Write([]byte("\n"))
	}

	code, err := lexgen.GenerateGoLexerCode(tables, packageName, typeName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if outputPath == "" || outputPath == "-" {
		_, _ = os.Stdout.Write(code)
		return
	}

	if err := os.WriteFile(outputPath, code, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
