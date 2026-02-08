package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/johnkerl/pgpg/manual/pkg/parsers"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"

	generatedlexers "github.com/johnkerl/pgpg/generated/pkg/lexers"
	generatedparsers "github.com/johnkerl/pgpg/generated/pkg/parsers"
)

type parserInfoT struct {
	run  func(string, traceOptions) (*asts.AST, error)
	help string
}

type traceOptions struct {
	tokens bool
	states bool
	stack  bool
}

var parserMakerTable = map[string]parserInfoT{
	"m:ame":    {run: runManualParser(parsers.NewAMEParser), help: "Integers with + and * at equal precedence."},
	"m:amne":   {run: runManualParser(parsers.NewAMNEParser), help: "Integers with + and * at unequal precedence."},
	"m:pemdas": {run: runManualParser(parsers.NewPEMDASParser), help: "Arithmetic with parentheses and PEMDAS precedence."},
	"m:vic":    {run: runManualParser(parsers.NewVICParser), help: "Arithmetic with identifiers, assignments, and PEMDAS precedence."},
	"m:vbc":    {run: runManualParser(parsers.NewVBCParser), help: "Boolean expressions with identifiers and AND/OR/NOT."},
	"m:ebnf":   {run: runManualParser(parsers.NewEBNFParser), help: "EBNF grammar with identifiers, literals, and operators."},
	"g:arith":  {run: runGeneratedArithParser, help: "Generated arithmetic parser from generated/pkg/arith-parse.go."},
	"g:arithw": {run: runGeneratedArithWhitespaceParser, help: "Generated arithmetic parser from generated/pkg/arithw-parse.go."},
	"g:stmts":  {run: runGeneratedStatementsParser, help: "Generated statements parser from generated/pkg/parsers/statements-parse.go."},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] {parser name} {one or more strings to parse ...}\n", os.Args[0])
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
	flag.BoolVar(&traceTokens, "tokens", false, "Print tokens as they're read")
	flag.BoolVar(&traceStates, "states", false, "Show parser state transitions")
	flag.BoolVar(&traceStack, "stack", false, "Show parser stack after each action")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 2 {
		usage()
	}
	parserName := flag.Arg(0)

	parserInfo, ok := parserMakerTable[parserName]
	if !ok {
		usage()
	}
	run := parserInfo.run
	opts := traceOptions{
		tokens: traceTokens,
		states: traceStates,
		stack:  traceStack,
	}

	for _, arg := range flag.Args()[1:] {
		ast, err := run(arg, opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// TODO: CLI option
		ast.Print()
		// ast.PrintParex()
	}
}

func runManualParser(maker func() parsers.AbstractParser) func(string, traceOptions) (*asts.AST, error) {
	return func(input string, _ traceOptions) (*asts.AST, error) {
		parser := maker()
		return parser.Parse(input)
	}
}

func runGeneratedArithParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewArithLexer(input)
	parser := generatedparsers.NewArithParser()
	attachArithTrace(parser, opts)
	return parser.Parse(lexer)
}

func runGeneratedArithWhitespaceParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewArithWhitespaceLexer(input)
	parser := generatedparsers.NewArithWhitespaceParser()
	attachArithWhitespaceTrace(parser, opts)
	return parser.Parse(lexer)
}

func runGeneratedStatementsParser(input string, opts traceOptions) (*asts.AST, error) {
	lexer := generatedlexers.NewStatementsLexer(input)
	parser := generatedparsers.NewStatementsParser()
	attachStatementsTrace(parser, opts)
	return parser.Parse(lexer)
}

func attachArithTrace(parser *generatedparsers.ArithParser, opts traceOptions) {
	if !opts.tokens && !opts.states && !opts.stack {
		return
	}
	parser.Trace = &generatedparsers.ArithParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !opts.tokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatToken(tok))
		},
		OnAction: func(state int, action generatedparsers.ArithParserAction, lookahead *tokens.Token) {
			if !opts.states {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n", state, formatArithAction(action), tokenTypeName(lookahead), tokenLexeme(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !opts.stack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n", formatIntStack(stateStack), formatNodeStack(nodeStack))
		},
	}
}

func attachArithWhitespaceTrace(parser *generatedparsers.ArithWhitespaceParser, opts traceOptions) {
	if !opts.tokens && !opts.states && !opts.stack {
		return
	}
	parser.Trace = &generatedparsers.ArithWhitespaceParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !opts.tokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatToken(tok))
		},
		OnAction: func(state int, action generatedparsers.ArithWhitespaceParserAction, lookahead *tokens.Token) {
			if !opts.states {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n", state, formatArithWhitespaceAction(action), tokenTypeName(lookahead), tokenLexeme(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !opts.stack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n", formatIntStack(stateStack), formatNodeStack(nodeStack))
		},
	}
}

func attachStatementsTrace(parser *generatedparsers.StatementsParser, opts traceOptions) {
	if !opts.tokens && !opts.states && !opts.stack {
		return
	}
	parser.Trace = &generatedparsers.StatementsParserTraceHooks{
		OnToken: func(tok *tokens.Token) {
			if !opts.tokens {
				return
			}
			fmt.Fprintln(os.Stderr, formatToken(tok))
		},
		OnAction: func(state int, action generatedparsers.StatementsParserAction, lookahead *tokens.Token) {
			if !opts.states {
				return
			}
			fmt.Fprintf(os.Stderr, "STATE %d %s on %s(%q)\n", state, formatStatementsAction(action), tokenTypeName(lookahead), tokenLexeme(lookahead))
		},
		OnStack: func(stateStack []int, nodeStack []*asts.ASTNode) {
			if !opts.stack {
				return
			}
			fmt.Fprintf(os.Stderr, "STACK states=%s nodes=%s\n", formatIntStack(stateStack), formatNodeStack(nodeStack))
		},
	}
}

func formatToken(tok *tokens.Token) string {
	if tok == nil {
		return "TOK <nil>"
	}
	return fmt.Sprintf("TOK type=%s lexeme=%q line=%d col=%d", tok.Type, string(tok.Lexeme), tok.Location.LineNumber, tok.Location.ColumnNumber)
}

func tokenTypeName(tok *tokens.Token) string {
	if tok == nil {
		return "<nil>"
	}
	return string(tok.Type)
}

func tokenLexeme(tok *tokens.Token) string {
	if tok == nil {
		return ""
	}
	return string(tok.Lexeme)
}

func formatIntStack(stack []int) string {
	parts := make([]string, len(stack))
	for i, v := range stack {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatNodeStack(stack []*asts.ASTNode) string {
	parts := make([]string, len(stack))
	for i, node := range stack {
		if node == nil {
			parts[i] = "<nil>"
			continue
		}
		parts[i] = string(node.Type)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func formatArithAction(action generatedparsers.ArithParserAction) string {
	switch action.Kind {
	case generatedparsers.ArithParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case generatedparsers.ArithParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case generatedparsers.ArithParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

func formatArithWhitespaceAction(action generatedparsers.ArithWhitespaceParserAction) string {
	switch action.Kind {
	case generatedparsers.ArithWhitespaceParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case generatedparsers.ArithWhitespaceParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case generatedparsers.ArithWhitespaceParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}

func formatStatementsAction(action generatedparsers.StatementsParserAction) string {
	switch action.Kind {
	case generatedparsers.StatementsParserActionShift:
		return fmt.Sprintf("shift(%d)", action.Target)
	case generatedparsers.StatementsParserActionReduce:
		return fmt.Sprintf("reduce(%d)", action.Target)
	case generatedparsers.StatementsParserActionAccept:
		return "accept"
	default:
		return "unknown"
	}
}
