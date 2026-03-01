// ================================================================
// Print routines for AST and ASTNode
// ================================================================

package asts

import (
	"fmt"
	"strings"
)

// String returns the indent-style multiline string representation.
func (a *AST) String() string {
	return a.RootNode.String()
}

// StringParex returns the parenthesized-expression string representation.
func (a *AST) StringParex() string {
	return a.RootNode.StringParex()
}

// StringParexOneLine returns the parenthesized-expression string, all on one line.
func (a *AST) StringParexOneLine() string {
	return a.RootNode.StringParexOneLine()
}

// Print is indent-style multiline print.
//
// Example given parse of "a + b":
// "+"
//
//	"a"
//	"b"
func (a *AST) Print() {
	fmt.Print(a.String())
}

// PrintParex is parenthesized-expression print.
//
// Example, given parse of 'a + b':
// (+ a b)
func (a *AST) PrintParex() {
	fmt.Print(a.StringParex())
}

// PrintParexOneLine is parenthesized-expression print, all on one line.
//
// Example, given parse of 'a + b':
// (+ a b)
func (a *AST) PrintParexOneLine() {
	fmt.Print(a.StringParexOneLine())
}

// String returns the indent-style multiline string representation.
func (n *ASTNode) String() string {
	var buf strings.Builder
	n.printAux(&buf, 0)
	return buf.String()
}

// Print is indent-style multiline print.
func (n *ASTNode) Print() {
	fmt.Print(n.String())
}

// printAux is a recursion-helper for Print.
func (n *ASTNode) printAux(buf *strings.Builder, depth int) {
	// Indent
	for i := 0; i < depth; i++ {
		buf.WriteString("    ")
	}

	// Token text (if non-nil) and token type
	tok := n.Token
	if tok != nil {
		buf.WriteString("\"")
		buf.WriteString(tok.LexemeText())
		buf.WriteString("\" [tt:")
		buf.WriteString(tok.TokenTypeText())
		buf.WriteString("] [nt:")
		buf.WriteString(string(n.Type))
		buf.WriteString("]")
	} else {
		buf.WriteString("[nt:")
		buf.WriteString(string(n.Type))
		buf.WriteString("]")
	}
	buf.WriteString("\n")

	// Children, indented one level further
	if n.Children != nil {
		for _, child := range n.Children {
			child.printAux(buf, depth+1)
		}
	}
}

// StringParex returns the parenthesized-expression string representation.
func (n *ASTNode) StringParex() string {
	var buf strings.Builder
	n.printParexAux(&buf, 0)
	return buf.String()
}

// PrintParex is parenthesized-expression print.
func (n *ASTNode) PrintParex() {
	fmt.Print(n.StringParex())
}

// printParexAux is a recursion-helper for PrintParex.
func (n *ASTNode) printParexAux(buf *strings.Builder, depth int) {
	if n.IsLeaf() {
		for i := 0; i < depth; i++ {
			buf.WriteString("    ")
		}
		buf.WriteString(n.Text())
		buf.WriteString("\n")

	} else if n.ChildrenAreAllLeaves() {
		// E.g. (= sum 0) or (+ 1 2)
		for i := 0; i < depth; i++ {
			buf.WriteString("    ")
		}
		buf.WriteString("(")
		buf.WriteString(n.Text())
		for _, child := range n.Children {
			buf.WriteString(" ")
			buf.WriteString(child.Text())
		}
		buf.WriteString(")\n")

	} else {
		// Parent and opening parenthesis on first line
		for i := 0; i < depth; i++ {
			buf.WriteString("    ")
		}
		buf.WriteString("(")
		buf.WriteString(n.Text())
		buf.WriteString("\n")

		// Children on their own lines
		for _, child := range n.Children {
			child.printParexAux(buf, depth+1)
		}

		// Closing parenthesis on last line
		for i := 0; i < depth; i++ {
			buf.WriteString("    ")
		}
		buf.WriteString(")\n")
	}
}

// StringParexOneLine returns the parenthesized-expression string, all on one line.
func (n *ASTNode) StringParexOneLine() string {
	var buf strings.Builder
	n.printParexOneLineAux(&buf)
	buf.WriteString("\n")
	return buf.String()
}

// PrintParexOneLine is parenthesized-expression print, all on one line.
func (n *ASTNode) PrintParexOneLine() {
	fmt.Print(n.StringParexOneLine())
}

// printParexOneLineAux is a recursion-helper for PrintParexOneLine.
func (n *ASTNode) printParexOneLineAux(buf *strings.Builder) {
	if n.IsLeaf() {
		buf.WriteString(n.Text())
		return
	}
	buf.WriteString("(")
	buf.WriteString(n.Text())
	for _, child := range n.Children {
		buf.WriteString(" ")
		child.printParexOneLineAux(buf)
	}
	buf.WriteString(")")
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
