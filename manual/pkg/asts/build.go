// ================================================================
// AST-build methods
// ================================================================

package asts

import (
	"fmt"

	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

// NewAST constructs a new root for the abstract syntax tree.
func NewAST(root *ASTNode) *AST {
	return &AST{
		RootNode: root,
	}
}

// NewASTNode constructs a new node for the abstract syntax tree.
//
// If children is non-nil and length 0, a zary node is created. (Example: a function call with zero
// arguments.) If children is nil, a terminal node is created. (Example: a string or integer
// literal.)
func NewASTNode(
	tok *tokens.Token,
	nodeType NodeType,
	children []*ASTNode,
) *ASTNode {
	node := &ASTNode{
		Token:    tok,
		Type:     nodeType,
		Children: nil,
	}

	if children == nil {
		return node
	}

	node.Children = children
	return node
}

func NewASTNodeTerminal(tok *tokens.Token, nodeType NodeType) *ASTNode {
	return &ASTNode{
		Token:    tok,
		Type:     nodeType,
		Children: nil,
	}
}

func WithChildPrepended(parent *ASTNode, child *ASTNode) (*ASTNode, error) {
	if parent.Children == nil {
		parent.Children = []*ASTNode{child}
	} else {
		parent.Children = append([]*ASTNode{child}, parent.Children...)
	}
	return parent, nil
}

func WithTwoChildrenPreprended(parent *ASTNode, childA, childB *ASTNode) (*ASTNode, error) {
	if parent.Children == nil {
		parent.Children = []*ASTNode{childA, childB}
	} else {
		parent.Children = append([]*ASTNode{childA, childB}, parent.Children...)
	}
	return parent, nil
}

func WithChildAppended(parent *ASTNode, child *ASTNode) (*ASTNode, error) {
	if parent.Children == nil {
		parent.Children = []*ASTNode{child}
	} else {
		parent.Children = append(parent.Children, child)
	}
	return parent, nil
}

func WithChildrenAdopted(parent *ASTNode, child *ASTNode) (*ASTNode, error) {
	parent.Children = child.Children
	child.Children = nil
	return parent, nil
}

func (node *ASTNode) CheckArity(
	arity int,
) error {
	if len(node.Children) != arity {
		return fmt.Errorf("expected AST node arity %d, got %d", arity, len(node.Children))
	} else {
		return nil
	}
}

// Tokens are produced by GOCC. However there is an exception: for the ternary
// operator I want the AST to have a "?:" token, which GOCC doesn't produce
// since nothing is actually spelled like that in the DSL.
func NewASTToken(iliteral interface{}, iclonee interface{}) *tokens.Token {
	literal := iliteral.(string)
	// clonee := iclonee.(*tokens.Token)
	return &tokens.Token{
		// Type: clonee.Type,
		Lexeme: []rune(literal),
		// Position: clonee.Position,
	}
}
