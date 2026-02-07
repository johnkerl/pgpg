// ================================================================
// AST and ASTNode data structures
// ================================================================

package asts

// TokenLike is a minimal interface for tokens used in the AST.
type TokenLike interface {
	LexemeText() string
	TokenTypeText() string
}

type AST[T TokenLike] struct {
	RootNode *ASTNode[T]
}

type ASTNode[T TokenLike] struct {
	Token *T // Nil for tokenless/structural nodes
	Type     TNodeType
	Children []*ASTNode[T]
}

type NodeType string
