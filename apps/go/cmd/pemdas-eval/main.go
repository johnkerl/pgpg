package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/johnkerl/pgpg/manual/go/pkg/asts"

	generatedlexers "github.com/johnkerl/pgpg/generated/go/pkg/lexers"
	generatedparsers "github.com/johnkerl/pgpg/generated/go/pkg/parsers"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] [-e | -l] [file ...]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  -e: arguments are expressions to parse (at least one required).\n")
	fmt.Fprintf(os.Stderr, "  -l: read stdin line-by-line, evaluate each line, print result (REPL mode).\n")
	fmt.Fprintf(os.Stderr, "  With -l and stdin a TTY, -p sets the prompt (default \"> \"); use -p \"\" to disable.\n")
	fmt.Fprintf(os.Stderr, "  Without -e/-l: zero arguments = read from stdin; one or more = read from those files.\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var verbose bool
	var exprMode bool
	var lineMode bool
	var prompt string
	flag.BoolVar(&verbose, "v", false, "Print AST before evaluation")
	flag.BoolVar(&exprMode, "e", false, "Arguments are expressions to parse (at least one required)")
	flag.BoolVar(&lineMode, "l", false, "Read stdin line-by-line, evaluate each, print result (REPL)")
	flag.StringVar(&prompt, "p", "> ", "In -l mode with TTY stdin, prompt string (default \"> \"; use \"\" to disable)")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if lineMode {
		if exprMode {
			fmt.Fprintln(os.Stderr, "pemdas-eval: -e and -l are mutually exclusive")
			os.Exit(1)
		}
		runREPL(verbose, prompt)
		return
	}

	if exprMode {
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "pemdas-eval: -e requires at least one argument")
			os.Exit(1)
		}
		for _, arg := range args {
			if err := runParserOnce(arg, verbose); err != nil {
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
			if err := runParserOnce(string(content), verbose); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		} else if err := runParserOnFiles(args, verbose); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func stdinIsTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func runREPL(verbose bool, prompt string) {
	usePrompt := stdinIsTTY() && prompt != ""
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if usePrompt {
			fmt.Fprint(os.Stdout, prompt)
			os.Stdout.Sync()
		}
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := runParserOnce(line, verbose); err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runParserOnFiles(filenames []string, verbose bool) error {
	for _, filename := range filenames {
		content, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if err := runParserOnce(string(content), verbose); err != nil {
			return err
		}
	}
	return nil
}

func runParserOnce(input string, verbose bool) error {
	lexer := generatedlexers.NewPEMDASLexer(input)
	parser := generatedparsers.NewPEMDASParser()
	ast, err := parser.Parse(lexer, "")
	if err != nil {
		return err
	}
	v, err := evaluateAST(ast, verbose)
	if err != nil {
		return err
	}
	fmt.Printf("%d\n", v)
	return nil
}

func evaluateAST(ast *asts.AST, verbose bool) (int, error) {
	if verbose {
		ast.Print()
	}

	if ast.RootNode == nil {
		fmt.Println("(nil AST)")
		return -1, nil
	}

	return evaluateASTNode(ast.RootNode)
}

func evaluateASTNode(node *asts.ASTNode) (int, error) {
	switch node.Type {

	case "int_literal":
		v, err := evaluateLiteralNode(node)
		if err != nil {
			return -1, err
		}
		return v, nil

	case "operator":
		v, err := evaluateBinaryOperatorNode(node)
		if err != nil {
			return -1, err
		}
		return v, nil

	case "unary":
		v, err := evaluateUnaryOperatorNode(node)
		if err != nil {
			return -1, err
		}
		return v, nil

	}

	return -1, fmt.Errorf("Unhandled node type \"%s\"", node.Type)
}

func evaluateLiteralNode(node *asts.ASTNode) (int, error) {
	if node.Token == nil {
		return -1, fmt.Errorf("Literal node has no token")
	}
	v, err := strconv.Atoi(string(node.Token.Lexeme))
	if err != nil {
		return -1, err
	}
	return v, nil
}

func intPower(base, exp int) int {
	if exp < 0 {
		return 0
	}
	out := 1
	for i := 0; i < exp; i++ {
		out *= base
	}
	return out
}

func evaluateBinaryOperatorNode(node *asts.ASTNode) (int, error) {
	op := string(node.Token.Lexeme)
	if len(node.Children) != 2 {
		return -1, fmt.Errorf("Expected two operands for operator \"%s\"; got %d",
			op, len(node.Children),
		)
	}

	c1, e1 := evaluateASTNode(node.Children[0])
	if e1 != nil {
		return -1, e1
	}
	c2, e2 := evaluateASTNode(node.Children[1])
	if e2 != nil {
		return -1, e2
	}

	switch op {
	case "+":
		return c1 + c2, nil
	case "-":
		return c1 - c2, nil
	case "*":
		return c1 * c2, nil
	case "/":
		return c1 / c2, nil
	case "%":
		return c1 % c2, nil
	case "**":
		return intPower(c1, c2), nil
	default:
		return -1, fmt.Errorf("Unhandled operator \"%s\"", op)
	}
}

func evaluateUnaryOperatorNode(node *asts.ASTNode) (int, error) {
	op := string(node.Token.Lexeme)
	if len(node.Children) != 1 {
		return -1, fmt.Errorf("Expected one operand for unary \"%s\"; got %d",
			op, len(node.Children),
		)
	}

	v, err := evaluateASTNode(node.Children[0])
	if err != nil {
		return -1, err
	}

	switch op {
	case "+":
		return v, nil
	case "-":
		return -v, nil
	default:
		return -1, fmt.Errorf("Unhandled unary operator \"%s\"", op)
	}
}
