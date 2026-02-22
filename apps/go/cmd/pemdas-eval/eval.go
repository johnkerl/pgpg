package main

import (
	"fmt"

	"github.com/johnkerl/pgpg/lib/go/pkg/asts"
)

// evaluateAST walks the AST using the given Numeric numeric and returns the result.
func evaluateAST[T, E any](ast *asts.AST, numeric Numeric[T, E], verbose bool) (T, error) {
	var zero T
	if verbose {
		ast.Print()
	}
	if ast.RootNode == nil {
		fmt.Println("(nil AST)")
		return zero, nil
	}
	return evaluateNode(ast.RootNode, numeric)
}

func evaluateNode[T, E any](node *asts.ASTNode, numeric Numeric[T, E]) (T, error) {
	var zero T
	switch node.Type {
	case "int_literal", "hex_literal", "float_literal", "bin_literal":
		return evaluateLiteralNode(node, numeric)
	case "operator":
		return evaluateBinaryOperatorNode(node, numeric)
	case "unary":
		return evaluateUnaryOperatorNode(node, numeric)
	default:
		return zero, fmt.Errorf("unhandled node type %q", node.Type)
	}
}

func isLiteralNode(node *asts.ASTNode) bool {
	return node != nil && node.Token != nil &&
		(node.Type == "int_literal" || node.Type == "hex_literal" || node.Type == "float_literal" || node.Type == "bin_literal")
}

func evaluateLiteralNode[T, E any](node *asts.ASTNode, numeric Numeric[T, E]) (T, error) {
	var zero T
	if node.Token == nil {
		return zero, fmt.Errorf("literal node has no token")
	}
	return numeric.FromString(string(node.Token.Lexeme))
}

func evaluateBinaryOperatorNode[T, E any](node *asts.ASTNode, numeric Numeric[T, E]) (T, error) {
	var zero T
	op := string(node.Token.Lexeme)
	if len(node.Children) != 2 {
		return zero, fmt.Errorf("expected two operands for operator %q; got %d", op, len(node.Children))
	}
	a, err := evaluateNode(node.Children[0], numeric)
	if err != nil {
		return zero, err
	}
	b, err := evaluateNode(node.Children[1], numeric)
	if err != nil {
		return zero, err
	}
	switch op {
	case "+":
		return numeric.Add(a, b), nil
	case "-":
		return numeric.Subtract(a, b), nil
	case "*":
		return numeric.Multiply(a, b), nil
	case "/":
		return numeric.Divide(a, b)
	case "%":
		return numeric.Mod(a, b)
	case "**":
		var exp E
		var err error
		if isLiteralNode(node.Children[1]) {
			exp, err = numeric.ParseExponent(string(node.Children[1].Token.Lexeme))
		} else {
			exp, err = numeric.ToExponent(b)
		}
		if err != nil {
			return zero, err
		}
		return numeric.Exponentiate(a, exp)
	default:
		return zero, fmt.Errorf("unhandled operator %q", op)
	}
}

func evaluateUnaryOperatorNode[T, E any](node *asts.ASTNode, numeric Numeric[T, E]) (T, error) {
	var zero T
	op := string(node.Token.Lexeme)
	if len(node.Children) != 1 {
		return zero, fmt.Errorf("expected one operand for unary %q; got %d", op, len(node.Children))
	}
	v, err := evaluateNode(node.Children[0], numeric)
	if err != nil {
		return zero, err
	}
	switch op {
	case "+":
		return v, nil
	case "-":
		return numeric.Negate(v), nil
	default:
		return zero, fmt.Errorf("unhandled unary operator %q", op)
	}
}
