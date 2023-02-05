// ================================================================
// AST-build methods
// ================================================================

package asts

import (
	"github.com/johnkerl/pgpg/pkg/tokens"
)

// ----------------------------------------------------------------
func NewAST(iroot interface{}) *AST {
	return &AST{
		RootNode: iroot.(*ASTNode),
	}
}

// ----------------------------------------------------------------
func NewASTNode(itok interface{}) *ASTNode {
	return NewASTNodeNestable(itok)
}

func NewASTNodeNestable(itok interface{}) *ASTNode {
	var tok *tokens.Token = nil
	if itok != nil {
		tok = itok.(*tokens.Token)
	}
	return &ASTNode{
		Token: tok,
		Children: nil,
	}
}

// Signature: Token Node Node Type
func NewASTNodeBinaryNestable(itok, childA, childB interface{}) *ASTNode {
	parent := NewASTNodeNestable(itok)
	convertToBinary(parent, childA, childB)
	return parent
}

// Signature: Token Node Node Type
func NewASTNodeBinary(
	itok, childA, childB interface{},
) (*ASTNode, error) {
	return NewASTNodeBinaryNestable(itok, childA, childB), nil
}

func convertToBinary(iparent interface{}, childA, childB interface{}) {
	parent := iparent.(*ASTNode)
	children := make([]*ASTNode, 2)
	children[0] = childA.(*ASTNode)
	children[1] = childB.(*ASTNode)
	parent.Children = children
}
