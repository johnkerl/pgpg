package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/johnkerl/pgpg/pkg/parsers"
)

type parserMaker func() parsers.AbstractParser
type parserInfoT struct {
	maker parserMaker
	help  string
}

var parserMakerTable = map[string]parserInfoT{
	"ame":  parserInfoT{parsers.NewAMEParser, "Integers with + and * at equal precedence."},
	"amne": parserInfoT{parsers.NewAMNEParser, "Integers with + and * at unequal precedence."},
	"pemdas": parserInfoT{
		parsers.NewPEMDASParser,
		"Arithmetic with parentheses and PEMDAS precedence.",
	},
	"vic": parserInfoT{
		parsers.NewVICParser,
		"Arithmetic with identifiers, assignments, and PEMDAS precedence.",
	},
	"vbc": parserInfoT{
		parsers.NewVBCParser,
		"Boolean expressions with identifiers and AND/OR/NOT.",
	},
	"ebnf": parserInfoT{
		parsers.NewEBNFParser,
		"EBNF grammar with identifiers, literals, and operators.",
	},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s {parser name} {one or more strings to lex ...}\n", os.Args[0])
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
	parserMaker := parserInfo.maker

	for _, arg := range os.Args[2:] {
		parser := parserMaker()
		ast, err := parser.Parse(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// TODO: CLI option
		ast.Print()
		// ast.PrintParex()
	}
}
