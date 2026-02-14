package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/johnkerl/pgpg/manual/go/pkg/lexers"

	generatedlexers "github.com/johnkerl/pgpg/generated/go/pkg/lexers"
)

type lexerMaker func(string) lexers.AbstractLexer
type lexerInfoT struct {
	maker lexerMaker
	help  string
}

var lexerMakerTable = map[string]lexerInfoT{
	"m:canned":       lexerInfoT{lexers.NewCannedTextLexer, "Does string-split on the input at startup."},
	"m:rune":         lexerInfoT{lexers.NewRuneLexer, "Each UTF-8 character is its own token."},
	"m:line":         lexerInfoT{lexers.NewLineLexer, "Each line of text is its own token. Carriage returns are not delivered."},
	"m:word":         lexerInfoT{lexers.NewWordLexer, "Each run of non-whitespace text is its own token. Whitespace is not delivered."},
	"m:seng":         lexerInfoT{lexers.NewSENGLexer, "SENG lexicon."},
	"m:am":           lexerInfoT{lexers.NewAMLexer, "Integers with + and *."},
	"m:pemdas":       lexerInfoT{lexers.NewPEMDASLexer, "Arithmetic with parentheses and PEMDAS operators."},
	"m:vic":          lexerInfoT{lexers.NewVICLexer, "Arithmetic with identifiers, assignments, and PEMDAS operators."},
	"m:vbc":          lexerInfoT{lexers.NewVBCLexer, "Boolean expressions with identifiers and AND/OR/NOT."},
	"m:ebnf":         lexerInfoT{lexers.NewEBNFLexer, "EBNF grammar with identifiers, literals, and operators."},
	"g:signd":        lexerInfoT{generatedlexers.NewSignDigitLexer, "Generated sign/digit lexer from bnfs/sign-digit.bnf."},
	"g:pemdas-plain": lexerInfoT{generatedlexers.NewPEMDASPlainLexer, "Generated PEMDAS lexer from bnfs/pemdas-plain.bnf."},
	"g:pemdas":       lexerInfoT{generatedlexers.NewPEMDASLexer, "Generated PEMDAS hinted lexer from bnfs/pemdas.bnf."},
	"g:stmts":        lexerInfoT{generatedlexers.NewStatementsLexer, "Generated statements lexer from bnfs/statements.bnf."},
	"g:seng":         lexerInfoT{generatedlexers.NewSENGLexer, "Generated statements lexer from bnfs/seng.bnf."},
	"g:lisp":         lexerInfoT{generatedlexers.NewLISPLexer, "Generated LISP lexer from bnfs/lisp.bnf."},
	"g:json":         lexerInfoT{generatedlexers.NewJSONLexer, "Generated JSON lexer from bnfs/json.bnf."},
	"g:json-plain":   lexerInfoT{generatedlexers.NewJSONPlainLexer, "Generated JSON lexer from bnfs/json_plain.bnf."},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s {lexer name} expr {one or more strings to lex ...}\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s {lexer name} file {one or more filenames}\n", os.Args[0])
	flag.PrintDefaults()
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
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 3 {
		usage()
	}
	lexerName := flag.Arg(0)
	mode := flag.Arg(1)
	args := flag.Args()[2:]

	lexerInfo, ok := lexerMakerTable[lexerName]
	if !ok {
		usage()
	}
	lexerMaker := lexerInfo.maker

	switch mode {
	case "expr":
		for _, arg := range args {
			if err := runLexerOnce(lexerMaker, arg); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	case "file":
		if err := runLexerOnFiles(lexerMaker, args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		usage()
	}
}

func runLexerOnce(lexerMaker lexerMaker, input string) error {
	lexer := lexerMaker(input)
	return lexers.Run(lexer)
}

func runLexerOnFiles(lexerMaker lexerMaker, filenames []string) error {
	for _, filename := range filenames {
		handle, err := os.Open(filename)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(handle)
		for scanner.Scan() {
			if err := runLexerOnce(lexerMaker, scanner.Text()); err != nil {
				_ = handle.Close()
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			_ = handle.Close()
			return err
		}
		if err := handle.Close(); err != nil {
			return err
		}
	}
	return nil
}
