// ================================================================
// Print routines for AST and ASTNode
// ================================================================

package asts

import (
	"fmt"
)

// Print is indent-style multiline print.
//
// Example given parse of "a + b":
// "+"
//     "a"
//     "b"

func (a *AST) Print() {
	a.RootNode.Print()
}

// PrintParex is parenthesized-expression print.
//
// Example, given parse of 'a + b':
// (+ a b)
func (a *AST) PrintParex() {
	a.RootNode.PrintParex()
}

// PrintParexOneLine is parenthesized-expression print, all on one line.
//
// Example, given parse of 'a + b':
// (+ a b)
func (a *AST) PrintParexOneLine() {
	a.RootNode.PrintParexOneLine()
}

// Print is indent-style multiline print.
func (n *ASTNode) Print() {
	n.printAux(0)
}

// printAux is a recursion-helper for Print.
func (n *ASTNode) printAux(depth int) {
	// Indent
	for i := 0; i < depth; i++ {
		fmt.Print("    ")
	}

	// Token text (if non-nil) and token type
	tok := n.Token
	if tok != nil {
		fmt.Printf("\"%s\" [tt:%s] [nt:%s]", tok.LexemeText(), tok.TokenTypeText(), n.Type)
	} else {
		fmt.Printf("[nt:%s]", n.Type)
	}
	fmt.Println()

	// Children, indented one level further
	if n.Children != nil {
		for _, child := range n.Children {
			child.printAux(depth + 1)
		}
	}
}

// PrintParex is parenthesized-expression print.
func (n *ASTNode) PrintParex() {
	n.printParexAux(0)
}

// printParexAux is a recursion-helper for PrintParex.
func (n *ASTNode) printParexAux(depth int) {
	if n.IsLeaf() {
		for i := 0; i < depth; i++ {
			fmt.Print("    ")
		}
		fmt.Println(n.Text())

	} else if n.ChildrenAreAllLeaves() {
		// E.g. (= sum 0) or (+ 1 2)
		for i := 0; i < depth; i++ {
			fmt.Print("    ")
		}
		fmt.Print("(")
		fmt.Print(n.Text())

		for _, child := range n.Children {
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
		fmt.Println(n.Text())

		// Children on their own lines
		for _, child := range n.Children {
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
func (n *ASTNode) PrintParexOneLine() {
	n.printParexOneLineAux()
	fmt.Println()
}

// printParexOneLineAux is a recursion-helper for PrintParexOneLine.
func (n *ASTNode) printParexOneLineAux() {
	if n.IsLeaf() {
		fmt.Print(n.Text())
		return
	}
	fmt.Print("(")
	fmt.Print(n.Text())
	for _, child := range n.Children {
		fmt.Print(" ")
		child.printParexOneLineAux()
	}
	fmt.Print(")")
}

// IsLeaf determines if an AST node is a leaf node.
func (n *ASTNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// ChildrenAreAllLeaves determines if an AST node's children are all leaf nodes.
func (n *ASTNode) ChildrenAreAllLeaves() bool {
	for _, child := range n.Children {
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
func (n *ASTNode) Text() string {
	tokenText := ""
	if n.Token != nil {
		tokenText = n.Token.LexemeText()
	}

	return tokenText
}
