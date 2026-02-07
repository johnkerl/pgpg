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
	// TODO
	// Type     TNodeType
	Children []*ASTNode[T]
}

//type TNodeType string
//const (
//	NodeTypeTBD TNodeType = "TBD"
//
//	// A special token which causes a panic when evaluated.  This is for testing.
//	NodeTypePanic TNodeType = "panic token"
//)
