package main

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/go/pkg/asts"
)

// evaluateAST walks the AST using the given Numeric backend and returns the result.
func evaluateAST[T, E any](ast *asts.AST, backend Numeric[T, E], verbose bool) (T, error) {
	var zero T
	if verbose {
		ast.Print()
	}
	if ast.RootNode == nil {
		fmt.Println("(nil AST)")
		return zero, nil
	}
	return evaluateNode(ast.RootNode, backend)
}

func evaluateNode[T, E any](node *asts.ASTNode, backend Numeric[T, E]) (T, error) {
	var zero T
	switch node.Type {
	case "int_literal":
		return evaluateLiteralNode(node, backend)
	case "operator":
		return evaluateBinaryOperatorNode(node, backend)
	case "unary":
		return evaluateUnaryOperatorNode(node, backend)
	default:
		return zero, fmt.Errorf("unhandled node type %q", node.Type)
	}
}

func evaluateLiteralNode[T, E any](node *asts.ASTNode, backend Numeric[T, E]) (T, error) {
	var zero T
	if node.Token == nil {
		return zero, fmt.Errorf("literal node has no token")
	}
	return backend.FromString(string(node.Token.Lexeme))
}

func evaluateBinaryOperatorNode[T, E any](node *asts.ASTNode, backend Numeric[T, E]) (T, error) {
	var zero T
	op := string(node.Token.Lexeme)
	if len(node.Children) != 2 {
		return zero, fmt.Errorf("expected two operands for operator %q; got %d", op, len(node.Children))
	}
	a, err := evaluateNode(node.Children[0], backend)
	if err != nil {
		return zero, err
	}
	b, err := evaluateNode(node.Children[1], backend)
	if err != nil {
		return zero, err
	}
	switch op {
	case "+":
		return backend.Add(a, b), nil
	case "-":
		return backend.Subtract(a, b), nil
	case "*":
		return backend.Multiply(a, b), nil
	case "/":
		return backend.Divide(a, b)
	case "%":
		return backend.Mod(a, b)
	case "**":
		exp, err := backend.ToExponent(b)
		if err != nil {
			return zero, err
		}
		return backend.Exponentiate(a, exp)
	default:
		return zero, fmt.Errorf("unhandled operator %q", op)
	}
}

func evaluateUnaryOperatorNode[T, E any](node *asts.ASTNode, backend Numeric[T, E]) (T, error) {
	var zero T
	op := string(node.Token.Lexeme)
	if len(node.Children) != 1 {
		return zero, fmt.Errorf("expected one operand for unary %q; got %d", op, len(node.Children))
	}
	v, err := evaluateNode(node.Children[0], backend)
	if err != nil {
		return zero, err
	}
	switch op {
	case "+":
		return v, nil
	case "-":
		return backend.Negate(v), nil
	default:
		return zero, fmt.Errorf("unhandled unary operator %q", op)
	}
}
