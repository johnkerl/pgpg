// ================================================================
// AST-build methods
// ================================================================

package asts

import (
	"fmt"

	"github.com/johnkerl/pgpg/pkg/tokens"
)

// NewAST constructs a new root for the abstract syntax tree.
// The AST is generic so parsers can use their own token types.
func NewAST[T TokenLike](root *ASTNode[T]) *AST[T] {
	return &AST[T]{
		RootNode: root,
	}
}

// NewASTNode constructs a new node for the abstract syntax tree.
//
// The token type is generic to allow different parser token representations.
//
// If children is non-nil and length 0, a zary node is created. (Example: a function call with zero
// arguments.) If children is nil, a terminal node is created. (Example: a string or integer
// literal.)
func NewASTNode[T TokenLike](
	tok *T,
	nodeType NodeType,
	children []*ASTNode[T],
) *ASTNode[T] {
	node := &ASTNode[T]{
		Token: tok,
		Type:     nodeType,
		Children: nil,
	}

	if children == nil {
		return node
	}

	node.Children = children
	return node
}

func NewASTNodeTerminal[T TokenLike](tok *T,nodeType NodeType) *ASTNode[T] {
	return &ASTNode[T]{
		Token: tok,
		Type:     nodeType,
		Children: nil,
	}
}

func WithChildPrepended[T TokenLike](parent *ASTNode[T], child *ASTNode[T]) (*ASTNode[T], error) {
	if parent.Children == nil {
		parent.Children = []*ASTNode[T]{child}
	} else {
		parent.Children = append([]*ASTNode[T]{child}, parent.Children...)
	}
	return parent, nil
}

func WithTwoChildrenPreprended[T TokenLike](parent *ASTNode[T], childA, childB *ASTNode[T]) (*ASTNode[T], error) {
	if parent.Children == nil {
		parent.Children = []*ASTNode[T]{childA, childB}
	} else {
		parent.Children = append([]*ASTNode[T]{childA, childB}, parent.Children...)
	}
	return parent, nil
}

func WithChildAppended[T TokenLike](parent *ASTNode[T], child *ASTNode[T]) (*ASTNode[T], error) {
	if parent.Children == nil {
		parent.Children = []*ASTNode[T]{child}
	} else {
		parent.Children = append(parent.Children, child)
	}
	return parent, nil
}

func WithChildrenAdopted[T TokenLike](parent *ASTNode[T], child *ASTNode[T]) (*ASTNode[T], error) {
	parent.Children = child.Children
	child.Children = nil
	return parent, nil
}

func (node *ASTNode[T]) CheckArity(
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
