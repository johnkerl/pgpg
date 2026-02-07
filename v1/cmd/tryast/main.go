package main

import (
	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

func main() {
	node := asts.NewASTNode(
		tokens.NewToken([]rune("+"), 1, tokens.NewNonDefaultTokenLocation(1, 2)),
		[]interface{}{
			asts.NewASTNode(tokens.NewToken([]rune("a"), 2, tokens.NewNonDefaultTokenLocation(1, 1)), nil),
			asts.NewASTNode(tokens.NewToken([]rune("b"), 2, tokens.NewNonDefaultTokenLocation(1, 3)), nil),
		},
	)
	ast := asts.NewAST(node)
	ast.Print()
	ast.PrintParex()
}
