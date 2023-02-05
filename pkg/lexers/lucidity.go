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
		desc, err := lexer.DecodeType(token.Type)
		if err != nil {
			return err
		}
		// TODO: token.String()
		fmt.Printf(
			"Line %d column %d type %s (%d) token <<%s>>\n",
			token.Location.LineNumber,
			token.Location.ColumnNumber,
			desc,
			token.Type,
			string(token.Lexeme),
		)
		if token.IsError() {
			break
		}
	}
	return nil
}
