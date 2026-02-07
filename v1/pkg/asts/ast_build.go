// ================================================================
// AST-build methods
// ================================================================

package asts

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
)

// NewAST constructs a new root for the abstract syntax tree.
// The argument-typing is interface{} as a holdover from my experience with GOCC; this may change.
func NewAST(iroot interface{}) *AST {
	return &AST{
		RootNode: iroot.(*ASTNode),
	}
}

// NewASTNode constructs a new node for the abstract syntax tree.
//
// The argument-typing is interface{}, rather than tokens.Token, as a holdover from my experience
// with GOCC; this will probably change.
//
// If children is non-nil and length 0, a zary node is created. (Example: a function call with zero
// arguments.) If children is nil, a terminal node is created. (Example: a string or integer
// literal.)
func NewASTNode(
	itok interface{},
	// nodeType TNodeType,
	children []interface{},
) *ASTNode {

	var tok *tokens.Token = nil
	if itok != nil {
		tok = itok.(*tokens.Token)
	}

	node := &ASTNode{
		Token: tok,
		// Type:     nodeType,
		Children: nil,
	}

	if children == nil {
		return node
	}

	n := len(children)
	node.Children = make([]*ASTNode, n)
	for i, child := range children {
		node.Children[i] = child.(*ASTNode)
	}
	return node
}

func NewASTNodeTerminal(itok interface{} /*nodeType TNodeType*/) *ASTNode {
	var tok *tokens.Token = nil
	if itok != nil {
		tok = itok.(*tokens.Token)
	}
	return &ASTNode{
		Token: tok,
		// Type:     nodeType,
		Children: nil,
	}
}

func WithChildPrepended(iparent interface{}, ichild interface{}) (*ASTNode, error) {
	parent := iparent.(*ASTNode)
	child := ichild.(*ASTNode)
	if parent.Children == nil {
		parent.Children = []*ASTNode{child}
	} else {
		parent.Children = append([]*ASTNode{child}, parent.Children...)
	}
	return parent, nil
}

func WithTwoChildrenPreprended(iparent interface{}, ichildA, ichildB interface{}) (*ASTNode, error) {
	parent := iparent.(*ASTNode)
	childA := ichildA.(*ASTNode)
	childB := ichildB.(*ASTNode)
	if parent.Children == nil {
		parent.Children = []*ASTNode{childA, childB}
	} else {
		parent.Children = append([]*ASTNode{childA, childB}, parent.Children...)
	}
	return parent, nil
}

func WithChildAppended(iparent interface{}, child interface{}) (*ASTNode, error) {
	parent := iparent.(*ASTNode)
	if parent.Children == nil {
		parent.Children = []*ASTNode{child.(*ASTNode)}
	} else {
		parent.Children = append(parent.Children, child.(*ASTNode))
	}
	return parent, nil
}

func WithChildrenAdopted(iparent interface{}, ichild interface{}) (*ASTNode, error) {
	parent := iparent.(*ASTNode)
	child := ichild.(*ASTNode)
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
