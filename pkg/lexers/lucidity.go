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
		fmt.Printf(
			"Line %d column %d token <<%s>>\n",
			token.Location.LineNumber,
			token.Location.ColumnNumber,
			string(token.Lexeme),
		) // TODO: token.String()
	}
	return nil
}
