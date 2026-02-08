package lexers

import "testing"

func TestArithWhitespaceLexerEmitsWhitespaceTokens(t *testing.T) {
	lexer := NewArithWhitespaceLexer("1 + 2")

	cases := []struct {
		lexeme string
		typ    string
	}{
		{lexeme: "1", typ: "int_literal"},
		{lexeme: " ", typ: "whitespace"},
		{lexeme: "+", typ: "plus"},
		{lexeme: " ", typ: "whitespace"},
		{lexeme: "2", typ: "int_literal"},
	}

	for i, tc := range cases {
		token := lexer.Scan()
		if token == nil {
			t.Fatalf("case %d: expected token, got nil", i)
		}
		if string(token.Lexeme) != tc.lexeme {
			t.Fatalf("case %d: expected lexeme %q, got %q", i, tc.lexeme, string(token.Lexeme))
		}
		if string(token.Type) != tc.typ {
			t.Fatalf("case %d: expected type %q, got %q", i, tc.typ, string(token.Type))
		}
	}

	if token := lexer.Scan(); token == nil || !token.IsEOF() {
		t.Fatal("expected EOF token")
	}
}
