// ================================================================
// AST and ASTNode data structures
// ================================================================

package ast

import (
	"github.com/johnkerl/pgpg/pkg/types"
)

// ----------------------------------------------------------------
type AST struct {
	RootNode *ASTNode
}

// ----------------------------------------------------------------
type ASTNode struct {
	Token    *types.Token // Nil for tokenless/structural nodes
	// TODO
	// Type     TNodeType
	Children []*ASTNode
}

//// ----------------------------------------------------------------
//type TNodeType string
//const (
//	NodeTypeTBD TNodeType = "TBD"
//
//	// A special token which causes a panic when evaluated.  This is for testing.
//	NodeTypePanic TNodeType = "panic token"
//)
