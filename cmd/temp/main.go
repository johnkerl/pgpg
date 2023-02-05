/*
go build github.com/johnkerl/pgpg/cmd/tmp
*/

package main

import (
	"github.com/johnkerl/pgpg/pkg/asts"
	"github.com/johnkerl/pgpg/pkg/tokens"
)

func main() {

	node := asts.NewASTNodeBinaryNestable(
		tokens.NewToken("+", tokens.NewTokenLocation(1, 1)),
		asts.NewASTNode(tokens.NewToken("a", tokens.NewTokenLocation(1, 1))),
		asts.NewASTNode(tokens.NewToken("b", tokens.NewTokenLocation(1, 3))),
	)
	ast := asts.NewAST(node)
	ast.Print()
	ast.PrintParex()
}
