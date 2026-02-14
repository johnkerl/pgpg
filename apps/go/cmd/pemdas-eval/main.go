package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/johnkerl/pgpg/manual/go/pkg/asts"

	generatedlexers "github.com/johnkerl/pgpg/generated/go/pkg/lexers"
	generatedparsers "github.com/johnkerl/pgpg/generated/go/pkg/parsers"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] expr {one or more strings to parse ...}\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s [options] file [one or more filenames]  (none = stdin)\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "Parser names:\n")
	os.Exit(1)
}

func main() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Print AST before evaluation")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}
	mode := flag.Arg(0)
	args := flag.Args()[1:]

	switch mode {
	case "expr":
		if len(args) == 0 {
			usage()
		}
		for _, arg := range args {
			if err := runParserOnce(arg, verbose); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	case "file":
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
	default:
		usage()
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
