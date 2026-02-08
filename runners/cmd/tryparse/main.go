package main

import (
	"fmt"
	"os"
	"sort"

	generatedpkg "github.com/johnkerl/pgpg/generated/pkg"
	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/johnkerl/pgpg/manual/pkg/parsers"
)

type parserInfoT struct {
	run  func(string) (*asts.AST, error)
	help string
}

var parserMakerTable = map[string]parserInfoT{
	"m:ame": {run: runManualParser(parsers.NewAMEParser), help: "Integers with + and * at equal precedence."},
	"m:amne": {run: runManualParser(parsers.NewAMNEParser), help: "Integers with + and * at unequal precedence."},
	"m:pemdas": {run: runManualParser(parsers.NewPEMDASParser), help: "Arithmetic with parentheses and PEMDAS precedence."},
	"m:vic": {run: runManualParser(parsers.NewVICParser), help: "Arithmetic with identifiers, assignments, and PEMDAS precedence."},
	"m:vbc": {run: runManualParser(parsers.NewVBCParser), help: "Boolean expressions with identifiers and AND/OR/NOT."},
	"m:ebnf": {run: runManualParser(parsers.NewEBNFParser), help: "EBNF grammar with identifiers, literals, and operators."},
	"g:arith": {run: runGeneratedArithParser, help: "Generated arithmetic parser from generated/pkg/arith-parse.go."},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s {parser name} {one or more strings to parse ...}\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Parser names:\n")
	names := make([]string, 0, len(parserMakerTable))
	for name := range parserMakerTable {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		maker := parserMakerTable[name]
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", name, maker.help)
	}
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}
	parserName := os.Args[1]

	parserInfo, ok := parserMakerTable[parserName]
	if !ok {
		usage()
	}
	run := parserInfo.run

	for _, arg := range os.Args[2:] {
		ast, err := run(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// TODO: CLI option
		ast.Print()
		// ast.PrintParex()
	}
}

func runManualParser(maker func() parsers.AbstractParser) func(string) (*asts.AST, error) {
	return func(input string) (*asts.AST, error) {
		parser := maker()
		return parser.Parse(input)
	}
}

func runGeneratedArithParser(input string) (*asts.AST, error) {
	lexer := generatedpkg.NewArithLexLexer(input)
	parser := generatedpkg.NewArithParseParser()
	return parser.Parse(lexer)
}
