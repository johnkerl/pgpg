// ================================================================
// AST-build methods
// ================================================================

package asts

import (
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
// The argument-typing is interface{}, rather than tokens.Token, as a holdover from my experience
// with GOCC; this will probably change.
func NewASTNode(itok interface{}) *ASTNode {
	return NewASTNodeNestable(itok)
}

// Holdover from my experience with GOCC; will almost certainly change.
func NewASTNodeNestable(itok interface{}) *ASTNode {
	var tok *tokens.Token = nil
	if itok != nil {
		tok = itok.(*tokens.Token)
	}
	return &ASTNode{
		Token:    tok,
		Children: nil,
	}
}

// Signature: Token
// Holdover from my experience with GOCC; will almost certainly change.
func NewASTNodeZaryNestable(itok interface{}) *ASTNode {
	parent := NewASTNodeNestable(itok)
	convertToZary(parent)
	return parent
}

// Signature: Token
// Holdover from my experience with GOCC; will almost certainly change.
func NewASTNodeZary(
	itok interface{},
) (*ASTNode, error) {
	return NewASTNodeZaryNestable(itok), nil
}

// Signature: Token Node Node Type
// Holdover from my experience with GOCC; will almost certainly change.
func NewASTNodeBinaryNestable(itok, childA, childB interface{}) *ASTNode {
	parent := NewASTNodeNestable(itok)
	convertToBinary(parent, childA, childB)
	return parent
}

// Signature: Token Node Node Type
// Holdover from my experience with GOCC; will almost certainly change.
func NewASTNodeBinary(
	itok, childA, childB interface{},
) (*ASTNode, error) {
	return NewASTNodeBinaryNestable(itok, childA, childB), nil
}

// Holdover from my experience with GOCC; will almost certainly change.
func convertToZary(iparent interface{}) {
	parent := iparent.(*ASTNode)
	children := make([]*ASTNode, 0)
	parent.Children = children
}

// Holdover from my experience with GOCC; will almost certainly change.
func convertToBinary(iparent interface{}, childA, childB interface{}) {
	parent := iparent.(*ASTNode)
	children := make([]*ASTNode, 2)
	children[0] = childA.(*ASTNode)
	children[1] = childB.(*ASTNode)
	parent.Children = children
}
