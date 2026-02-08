package main

import (
	"fmt"
	"os"

	"github.com/johnkerl/pgpg/pkg/lexers"
)

type lexerMaker func(string) lexers.AbstractLexer
type lexerInfoT struct {
	maker lexerMaker
	help  string
}

var lexerMakerTable = map[string]lexerInfoT{
	"canned": lexerInfoT{lexers.NewCannedTextLexer, "Does string-split on the input at startup."},
	"rune":   lexerInfoT{lexers.NewRuneLexer, "Each UTF-8 character is its own token."},
	"line":   lexerInfoT{lexers.NewLineLexer, "Each line of text is its own token. Carriage returns are not delivered."},
	"word":   lexerInfoT{lexers.NewWordLexer, "Each run of non-whitespace text is its own token. Whitespace is not delivered."},
	"seng":   lexerInfoT{lexers.NewSENGLexer, "SENG lexicon."},
	"am":     lexerInfoT{lexers.NewAMLexer, "Integers with + and *."},
	"pemdas": lexerInfoT{lexers.NewPEMDASLexer, "Arithmetic with parentheses and PEMDAS operators."},
	"vic":    lexerInfoT{lexers.NewVICLexer, "Arithmetic with identifiers, assignments, and PEMDAS operators."},
	"vbc":    lexerInfoT{lexers.NewVBCLexer, "Boolean expressions with identifiers and AND/OR/NOT."},
	"ebnf":   lexerInfoT{lexers.NewEBNFLexer, "EBNF grammar with identifiers, literals, and operators."},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s {lexer name} {one or more strings to lex ...}\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Lexer names:\n")
	// TODO: this prints in random hashmap order :(
	// Use sort-keys to determinize.
	for name, maker := range lexerMakerTable {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", name, maker.help)
	}
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}
	lexerName := os.Args[1]

	lexerInfo, ok := lexerMakerTable[lexerName]
	if !ok {
		usage()
	}
	lexerMaker := lexerInfo.maker

	for _, arg := range os.Args[2:] {
		lexer := lexerMaker(arg)
		err := lexers.Run(lexer)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
