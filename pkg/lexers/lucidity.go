package lexers

import (
	"fmt"
)

func Run(lexer AbstractLexer) error {
	for {
		token := lexer.Scan()
		desc, err := lexer.DecodeType(token.Type)
		if err != nil {
			return err
		}
		// TODO: token.String()
		fmt.Printf(
			"Line %4d column %4d type %-10s (%2d) token <<%s>>\n",
			token.Location.LineNumber,
			token.Location.ColumnNumber,
			desc,
			token.Type,
			string(token.Lexeme),
		)
		if token.IsEOF() || token.IsError() {
			break
		}
	}
	return nil
}
