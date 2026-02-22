package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/johnkerl/pgpg/apps/go/manual/parsers"
	"github.com/johnkerl/pgpg/lib/go/pkg/asts"
	liblexers "github.com/johnkerl/pgpg/lib/go/pkg/lexers"
	libparsers "github.com/johnkerl/pgpg/lib/go/pkg/parsers"

	generatedlexers "github.com/johnkerl/pgpg/generated/go/pkg/lexers"
	generatedparsers "github.com/johnkerl/pgpg/generated/go/pkg/parsers"
)

// generatedParser is the common interface for all generated parsers (AttachCLITrace + Parse(lexer, astMode)).
type generatedParser interface {
	AttachCLITrace(traceTokens, traceStates, traceStack bool)
	Parse(lexer liblexers.AbstractLexer, astMode string) (*asts.AST, error)
}

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
	"m:ame":    {run: runManualParser(parsers.NewAMEParser), help: "Integers with + and * at equal precedence."},
	"m:amne":   {run: runManualParser(parsers.NewAMNEParser), help: "Integers with + and * at unequal precedence."},
	"m:pemdas": {run: runManualParser(parsers.NewPEMDASParser), help: "Arithmetic with parentheses and PEMDAS precedence."},
	"m:vic":    {run: runManualParser(parsers.NewVICParser), help: "Arithmetic with identifiers, assignments, and PEMDAS precedence."},
	"m:vbc":    {run: runManualParser(parsers.NewVBCParser), help: "Boolean expressions with identifiers and AND/OR/NOT."},
	"m:ebnf":   {run: runManualParser(libparsers.NewEBNFParser), help: "EBNF grammar with identifiers, literals, and operators."},

	"g:pemdas-plain": {
		run:  runGeneratedParser(generatedlexers.NewPEMDASPlainLexer, func() generatedParser { return generatedparsers.NewPEMDASPlainParser() }),
		help: "Generated arithmetic parser from apps/bnfs/pemdas-plain.bnf.",
	},
	"g:pemdas": {
		run:  runGeneratedParser(generatedlexers.NewPEMDASLexer, func() generatedParser { return generatedparsers.NewPEMDASParser() }),
		help: "Generated arithmetic parser with AST hints from apps/bnfs/pemdas.bnf.",
	},
	"g:stmts": {
		run:  runGeneratedParser(generatedlexers.NewStatementsLexer, func() generatedParser { return generatedparsers.NewStatementsParser() }),
		help: "Generated statements parser from apps/bnfs/statements.bnf.",
	},
	"g:seng": {
		run:  runGeneratedParser(generatedlexers.NewSENGLexer, func() generatedParser { return generatedparsers.NewSENGParser() }),
		help: "Generated SENG parser from apps/bnfs/seng.bnf.",
	},
	"g:lisp": {
		run:  runGeneratedParser(generatedlexers.NewLISPLexer, func() generatedParser { return generatedparsers.NewLISPParser() }),
		help: "Generated LISP parser from apps/bnfs/lisp.bnf.",
	},
	"g:json": {
		run:  runGeneratedParser(generatedlexers.NewJSONLexer, func() generatedParser { return generatedparsers.NewJSONParser() }),
		help: "Generated JSON parser from apps/bnfs/json.bnf.",
	},
	"g:json-plain": {
		run:  runGeneratedParser(generatedlexers.NewJSONPlainLexer, func() generatedParser { return generatedparsers.NewJSONPlainParser() }),
		help: "Generated JSON parser from apps/bnfs/json_plain.bnf.",
	},
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
		return
	}
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

func runManualParser(maker func() libparsers.AbstractParser) func(string, traceOptions) (*asts.AST, error) {
	return func(input string, _ traceOptions) (*asts.AST, error) {
		parser := maker()
		return parser.Parse(input)
	}
}

func runGeneratedParser(
	newLexer func(string) liblexers.AbstractLexer,
	newParser func() generatedParser,
) func(string, traceOptions) (*asts.AST, error) {
	return func(input string, opts traceOptions) (*asts.AST, error) {
		lexer := newLexer(input)
		parser := newParser()
		parser.AttachCLITrace(opts.tokens, opts.states, opts.stack)
		return parser.Parse(lexer, opts.astMode)
	}
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
