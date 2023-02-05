package lexers

import (
	"fmt"
)

func Run(lexer AbstractLexer) error {
	for {
		token := lexer.Scan()
		if token.IsEOF() {
			break
		}
		// TODO: token.String()
		fmt.Printf(
			"Line %d column %d type %d token <<%s>>\n",
			token.Location.LineNumber,
			token.Location.ColumnNumber,
			token.Type, // TODO: somewhere in the API, retain a mapping between code and human-friendly type names
			string(token.Lexeme),
		)
		if token.IsError() {
			break
		}
	}
	return nil
}
