package main

import (
	"fmt"
	"os"

	"github.com/johnkerl/pgpg/pkg/lexers"
)

func main() {
	lexer := lexers.NewCannedTextLexer("the quick brown fox jumped over the lazy dogs")
	err := lexers.Run(lexer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
