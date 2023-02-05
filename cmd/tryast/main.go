package main

import (
	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

func main() {
	node := asts.NewASTNodeBinaryNestable(
		tokens.NewToken([]rune("+"), tokens.NewTokenLocation(1, 2)),
		asts.NewASTNode(tokens.NewToken([]rune("a"), tokens.NewTokenLocation(1, 1))),
		asts.NewASTNode(tokens.NewToken([]rune("b"), tokens.NewTokenLocation(1, 3))),
	)
	ast := asts.NewAST(node)
	ast.Print()
	ast.PrintParex()
}
