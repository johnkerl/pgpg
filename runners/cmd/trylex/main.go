package main

import (
	"fmt"
	"os"
	"sort"

	generatedlexers "github.com/johnkerl/pgpg/generated/pkg"
	"github.com/johnkerl/pgpg/manual/pkg/lexers"
)

type lexerMaker func(string) lexers.AbstractLexer
type lexerInfoT struct {
	maker lexerMaker
	help  string
}

var lexerMakerTable = map[string]lexerInfoT{
	"m:canned": lexerInfoT{lexers.NewCannedTextLexer, "Does string-split on the input at startup."},
	"m:rune":   lexerInfoT{lexers.NewRuneLexer, "Each UTF-8 character is its own token."},
	"m:line":   lexerInfoT{lexers.NewLineLexer, "Each line of text is its own token. Carriage returns are not delivered."},
	"m:word":   lexerInfoT{lexers.NewWordLexer, "Each run of non-whitespace text is its own token. Whitespace is not delivered."},
	"m:seng":   lexerInfoT{lexers.NewSENGLexer, "SENG lexicon."},
	"m:am":     lexerInfoT{lexers.NewAMLexer, "Integers with + and *."},
	"m:pemdas": lexerInfoT{lexers.NewPEMDASLexer, "Arithmetic with parentheses and PEMDAS operators."},
	"m:vic":    lexerInfoT{lexers.NewVICLexer, "Arithmetic with identifiers, assignments, and PEMDAS operators."},
	"m:vbc":    lexerInfoT{lexers.NewVBCLexer, "Boolean expressions with identifiers and AND/OR/NOT."},
	"m:ebnf":   lexerInfoT{lexers.NewEBNFLexer, "EBNF grammar with identifiers, literals, and operators."},
	"g:arith":  lexerInfoT{generatedlexers.NewArithLexLexer, "Generated arithmetic lexer from generated/pkg/arith-lex.go."},
	"g:arithw": lexerInfoT{generatedlexers.NewArithLexWhitespaceLexer, "Generated arithmetic lexer from generated/arithw.go."},
	"g:signd":  lexerInfoT{generatedlexers.NewSignDigitLexLexer, "Generated sign/digit lexer from generated/pkg/sign-digit.go."},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s {lexer name} {one or more strings to lex ...}\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Lexer names:\n")
	names := make([]string, 0, len(lexerMakerTable))
	for name := range lexerMakerTable {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		maker := lexerMakerTable[name]
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
