package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/johnkerl/pgpg/generator_go/pkg/lexgen"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [-o output.json] input.bnf\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var outputPath string
	flag.StringVar(&outputPath, "o", "", "Output JSON file (default stdout)")
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

	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tables, err := lexgen.GenerateTablesFromEBNFWithSourceName(string(inputBytes), absPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	jsonBytes, err := lexgen.EncodeTables(tables)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if outputPath == "" || outputPath == "-" {
		_, _ = os.Stdout.Write(jsonBytes)
		_, _ = os.Stdout.Write([]byte("\n"))
		return
	}

	if err := os.WriteFile(outputPath, jsonBytes, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
