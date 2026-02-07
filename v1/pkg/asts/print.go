// ================================================================
// Print routines for AST and ASTNode
// ================================================================

package asts

import (
	"fmt"
)

// ================================================================
// Print is indent-style multiline print.
//
// Example given parse of "a + b":
// "+"
//     "a"
//     "b"

func (ast *AST[T]) Print() {
	ast.RootNode.Print()
}

// PrintParex is parenthesized-expression print.
//
// Example, given parse of 'a + b':
// (+ a b)
func (ast *AST[T]) PrintParex() {
	ast.RootNode.PrintParex()
}

// PrintParexOneLine is parenthesized-expression print, all on one line.
//
// Example, given parse of 'a + b':
// (+ a b)
func (ast *AST[T]) PrintParexOneLine() {
	ast.RootNode.PrintParexOneLine()
}

// Print is indent-style multiline print.
func (node *ASTNode[T]) Print() {
	node.printAux(0)
}

// printAux is a recursion-helper for Print.
func (node *ASTNode[T]) printAux(depth int) {
	// Indent
	for i := 0; i < depth; i++ {
		fmt.Print("    ")
	}

	// Token text (if non-nil) and token type
	tok := node.Token
	// TODO fmt.Print("* " + node.Type)
	if tok != nil {
		fmt.Printf("\"%s\" (%s)", (*tok).LexemeText(), (*tok).TokenTypeText())
	}
	fmt.Println()

	// Children, indented one level further
	if node.Children != nil {
		for _, child := range node.Children {
			child.printAux(depth + 1)
		}
	}
}

// PrintParex is parenthesized-expression print.
func (node *ASTNode[T]) PrintParex() {
	node.printParexAux(0)
}

// printParexAux is a recursion-helper for PrintParex.
func (node *ASTNode[T]) printParexAux(depth int) {
	if node.IsLeaf() {
		for i := 0; i < depth; i++ {
			fmt.Print("    ")
		}
		fmt.Println(node.Text())

	} else if node.ChildrenAreAllLeaves() {
		// E.g. (= sum 0) or (+ 1 2)
		for i := 0; i < depth; i++ {
			fmt.Print("    ")
		}
		fmt.Print("(")
		fmt.Print(node.Text())

		for _, child := range node.Children {
			fmt.Print(" ")
			fmt.Print(child.Text())
		}
		fmt.Println(")")

	} else {
		// Parent and opening parenthesis on first line
		for i := 0; i < depth; i++ {
			fmt.Print("    ")
		}
		fmt.Print("(")
		fmt.Println(node.Text())

		// Children on their own lines
		for _, child := range node.Children {
			child.printParexAux(depth + 1)
		}

		// Closing parenthesis on last line
		for i := 0; i < depth; i++ {
			fmt.Print("    ")
		}
		fmt.Println(")")
	}
}

// PrintParexOneLine is parenthesized-expression print, all on one line.
func (node *ASTNode[T]) PrintParexOneLine() {
	node.printParexOneLineAux()
	fmt.Println()
}

// printParexOneLineAux is a recursion-helper for PrintParexOneLine.
func (node *ASTNode[T]) printParexOneLineAux() {
	if node.IsLeaf() {
		fmt.Print(node.Text())
	} else {
		fmt.Print("(")
		fmt.Print(node.Text())
		for _, child := range node.Children {
			fmt.Print(" ")
			child.printParexOneLineAux()
		}
		fmt.Print(")")
	}
}

// IsLeaf determines if an AST node is a leaf node.
func (node *ASTNode[T]) IsLeaf() bool {
	return len(node.Children) == 0
}

// ChildrenAreAllLeaves determines if an AST node's children are all leaf nodes.
func (node *ASTNode[T]) ChildrenAreAllLeaves() bool {
	for _, child := range node.Children {
		if !child.IsLeaf() {
			return false
		}
	}
	return true
}

// Text makes a human-readable, whitespace-free name for an AST node. Some
// nodes have non-nil tokens; other, nil. And token-types can have spaces in
// them. In this method we use custom mappings to always get a whitespace-free
// representation of the content of a single AST node.
func (node *ASTNode[T]) Text() string {
	tokenText := ""
	if node.Token != nil {
		tokenText = (*node.Token).LexemeText()
	}

	// TODO
	//	switch node.Type {
	//
	//	case NodeTypeStringLiteral:
	//		return "\"" + strings.ReplaceAll(tokenText, "\"", "\\\"") + "\""
	//	case NodeTypeIntLiteral:
	//		return tokenText
	//	}

	return tokenText
}
