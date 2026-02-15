package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/johnkerl/pgpg/manual/go/pkg/asts"
	"github.com/johnkerl/pgpg/manual/go/pkg/parsers"

	generatedlexers "github.com/johnkerl/pgpg/generated/go/pkg/lexers"
	generatedparsers "github.com/johnkerl/pgpg/generated/go/pkg/parsers"
)

type parserInfoT struct {
	run  func(string, traceOptions) (*asts.AST, error)
	help string
}

type traceOptions struct {
	tokens  bool
	states  bool
	stack   bool
	astMode string // "", "noast", or "fullast"
}

var parserMakerTable = map[string]parserInfoT{
	"m:ame":          {run: runManualParser(parsers.NewAMEParser), help: "Integers with + and * at equal precedence."},
	"m:amne":         {run: runManualParser(parsers.NewAMNEParser), help: "Integers with + and * at unequal precedence."},
	"m:pemdas":       {run: runManualParser(parsers.NewPEMDASParser), help: "Arithmetic with parentheses and PEMDAS precedence."},
	"m:vic":          {run: runManualParser(parsers.NewVICParser), help: "Arithmetic with identifiers, assignments, and PEMDAS precedence."},
	"m:vbc":          {run: runManualParser(parsers.NewVBCParser), help: "Boolean expressions with identifiers and AND/OR/NOT."},
	"m:ebnf":         {run: runManualParser(parsers.NewEBNFParser), help: "EBNF grammar with identifiers, literals, and operators."},
	"g:pemdas-plain": {run: runGeneratedPEMDASPlainParser, help: "Generated arithmetic parser from bnfs/pemdas-plain.bnf."},
	"g:pemdas":       {run: runGeneratedPEMDASParser, help: "Generated arithmetic parser with AST hints from bnfs/pemdas.bnf."},
	"g:stmts":        {run: runGeneratedStatementsParser, help: "Generated statements parser from generated/bnffs/statements.bnf."},
	"g:seng":         {run: runGeneratedSENGParser, help: "Generated SENG parser from generated/bnffs/seng.bnf."},
	"g:lisp":         {run: runGeneratedLISPParser, help: "Generated LISP parser from bnfs/lisp.bnf."},
	"g:json":         {run: runGeneratedJSONParser, help: "Generated JSON parser from bnfs/json.bnf."},
	"g:json-plain":   {run: runGeneratedJSONPlainParser, help: "Generated JSON parser from bnfs/json_plain.bnf."},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] {parser name} [file ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  With -e (before parser name): one or more positional args are expressions (error if none).\n")
	fmt.Fprintf(os.Stderr, "  Without -e: zero args = read from stdin; one or more = read from those files.\n")
	flag.PrintDefaults()
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
	var traceTokens bool
	var traceStates bool
	var traceStack bool
	var noast bool
	var fullast bool
	var exprMode bool
	flag.BoolVar(&traceTokens, "tokens", false, "Print tokens as they're read")
	flag.BoolVar(&traceStates, "states", false, "Show parser state transitions")
	flag.BoolVar(&traceStack, "stack", false, "Show parser stack after each action")
	flag.BoolVar(&noast, "noast", false, "Syntax-only: do not build or print AST (generated parsers only)")
	flag.BoolVar(&fullast, "fullast", false, "Ignore AST hints and build full parse tree (generated parsers only)")
	flag.BoolVar(&exprMode, "e", false, "Arguments are expressions to parse (at least one required)")
	flag.Usage = usage
	flag.Parse()

	if noast && fullast {
		fmt.Fprintln(os.Stderr, "cannot use -noast and -fullast together")
		os.Exit(1)
	}
	astMode := ""
	if noast {
		astMode = "noast"
	} else if fullast {
		astMode = "fullast"
	}

	if flag.NArg() < 1 {
		usage()
	}
	parserName := flag.Arg(0)
	args := flag.Args()[1:]

	parserInfo, ok := parserMakerTable[parserName]
	if !ok {
		usage()
	}
	run := parserInfo.run
	opts := traceOptions{
		tokens:  traceTokens,
		states:  traceStates,
		stack:   traceStack,
		astMode: astMode,
	}

	if exprMode {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "tryparse: -e requires at least one argument")
			os.Exit(1)
		}
		for _, arg := range args {
			if err := runParserOnce(run, arg, opts); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	} else {
		if len(args) == 0 {
			content, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err := runParserOnce(run, string(content), opts); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else if err := runParserOnFiles(run, args, opts); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func runManualParser(maker func() parsers.AbstractParser) func(string, traceOptions) (*asts.AST, error) {
	return func(input string, _ traceOptions) (*asts.AST, error) {
		parser := maker()
		return parser.Parse(input)
	}
}

func runGeneratedPEMDASPlainParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewPEMDASPlainLexer(input)
	parser := generatedparsers.NewPEMDASPlainParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	return parser.Parse(lexer, opts.astMode)
}

func runGeneratedPEMDASParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewPEMDASLexer(input)
	parser := generatedparsers.NewPEMDASParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	return parser.Parse(lexer, opts.astMode)
}

func runGeneratedStatementsParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewStatementsLexer(input)
	parser := generatedparsers.NewStatementsParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	return parser.Parse(lexer, opts.astMode)
}

func runGeneratedSENGParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewSENGLexer(input)
	parser := generatedparsers.NewSENGParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	return parser.Parse(lexer, opts.astMode)
}

func runGeneratedLISPParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewLISPLexer(input)
	parser := generatedparsers.NewLISPParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	return parser.Parse(lexer, opts.astMode)
}

func runGeneratedJSONParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewJSONLexer(input)
	parser := generatedparsers.NewJSONParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	return parser.Parse(lexer, opts.astMode)
}

func runGeneratedJSONPlainParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewJSONPlainLexer(input)
	parser := generatedparsers.NewJSONPlainParser()
	parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
	return parser.Parse(lexer, opts.astMode)
}

func runParserOnce(run func(string, traceOptions) (*asts.AST, error), input string, opts traceOptions) error {
	// TODO: CLI option
	fmt.Println(input)
	ast, err := run(input, opts)
	if err != nil {
		return err
	}
	if ast != nil && opts.astMode != "noast" {
		ast.Print()
	}
	return nil
}

func runParserOnFiles(run func(string, traceOptions) (*asts.AST, error), filenames []string, opts traceOptions) error {
	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := runParserOnce(run, string(content), opts); err != nil {
			return err
		}
	}
	return nil
}
