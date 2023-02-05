package lexers

import (
	"fmt"
)

func Run(lxr AbstractLexer) error {
	for {
		token, err := lxr.Scan()
		if err != nil {
			return err
		}
		if token == nil {
			break // EOF
		}
		fmt.Printf("TOKEN: %s\n", token.Lexeme) // TODO: token.String()
	}
	return nil
}
