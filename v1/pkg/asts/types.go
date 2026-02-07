// ================================================================
// AST and ASTNode data structures
// ================================================================

package asts

import "github.com/johnkerl/pgpg/pkg/tokens"

type AST struct {
	RootNode *ASTNode
}

type ASTNode struct {
	Token    *tokens.Token // Nil for tokenless/structural nodes
	Type     NodeType
	Children []*ASTNode
}

type NodeType string
