package main

import (
	"github.com/johnkerl/pgpg/manual/pkg/asts"
	"github.com/johnkerl/pgpg/manual/pkg/tokens"
)

func main() {
	node := asts.NewASTNode(
		tokens.NewToken([]rune("+"), tokens.TokenType("+"), tokens.NewNonDefaultTokenLocation(1, 2)),
		asts.NodeType("operator"),
		[]*asts.ASTNode{
			asts.NewASTNode(
				tokens.NewToken([]rune("a"), tokens.TokenType("word"), tokens.NewNonDefaultTokenLocation(1, 1)),
				asts.NodeType("word"),
				nil,
			),
			asts.NewASTNode(
				tokens.NewToken([]rune("b"), tokens.TokenType("word"), tokens.NewNonDefaultTokenLocation(1, 3)),
				asts.NodeType("word"),
				nil,
			),
		},
	)
	ast := asts.NewAST(node)
	ast.Print()
	ast.PrintParex()
}
