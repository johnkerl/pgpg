/*
go build github.com/johnkerl/pgpg/cmd/tmp
*/

package main

import (
	"github.com/johnkerl/pgpg/pkg/ast" // XXX RENAME
	"github.com/johnkerl/pgpg/pkg/types" // XXX RENAME
)

func main() {

	node := ast.NewASTNodeBinaryNestable(
		types.NewToken("+", types.NewTokenLocation(1, 1)),
		ast.NewASTNode(types.NewToken("a", types.NewTokenLocation(1, 1))),
		ast.NewASTNode(types.NewToken("b", types.NewTokenLocation(1, 3))),
	)
	a := ast.NewAST(node)
	a.Print()
	a.PrintParex()
}
