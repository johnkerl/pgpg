package lexers

import (
	"fmt"
)

func Run(lexer AbstractLexer) error {
	for {
		token := lexer.Scan()
		fmt.Printf(
			"Line %4d column %4d type %-16s token <<%s>>\n",
			token.Location.LineNumber,
			token.Location.ColumnNumber,
			token.Type,
			string(token.Lexeme),
		)
		if token.IsEOF() || token.IsError() {
			break
		}
	}
	return nil
}
