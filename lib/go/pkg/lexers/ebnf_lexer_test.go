package lexers

import (
	"testing"

	"github.com/johnkerl/pgpg/lib/pkg/tokens"
	"github.com/stretchr/testify/assert"
)

// ebnfExpectedToken represents one expected token for table-driven EBNF lexer tests.
// Use lexeme "" and type tokens.TokenTypeEOF for the final EOF token.
type ebnfExpectedToken struct {
	lexeme string
	typ    tokens.TokenType
}

func TestEBNFLexerTableDriven(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []ebnfExpectedToken
	}{
		{
			name:  "empty",
			input: "",
			want:  []ebnfExpectedToken{{"", tokens.TokenTypeEOF}},
		},
		{
			name:  "rule with alternation",
			input: "rule ::= \"a\" | 'b' ;",
			want: []ebnfExpectedToken{
				{"rule", EBNFLexerTypeIdentifier},
				{"::=", EBNFLexerTypeAssign},
				{"\"a\"", EBNFLexerTypeString},
				{"|", EBNFLexerTypeOr},
				{"'b'", EBNFLexerTypeString},
				{";", EBNFLexerTypeSemicolon},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "brackets and equals",
			input: "[ { ( ) } ] =",
			want: []ebnfExpectedToken{
				{"[", EBNFLexerTypeLBracket},
				{"{", EBNFLexerTypeLBrace},
				{"(", EBNFLexerTypeLParen},
				{")", EBNFLexerTypeRParen},
				{"}", EBNFLexerTypeRBrace},
				{"]", EBNFLexerTypeRBracket},
				{"=", EBNFLexerTypeAssign},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "standalone colon",
			input: "x :=",
			want: []ebnfExpectedToken{
				{"x", EBNFLexerTypeIdentifier},
				{":", EBNFLexerTypeColon},
				{"=", EBNFLexerTypeAssign},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "arrow and hint block",
			input: `A ::= B -> { "parent" : 0 }`,
			want: []ebnfExpectedToken{
				{"A", EBNFLexerTypeIdentifier},
				{"::=", EBNFLexerTypeAssign},
				{"B", EBNFLexerTypeIdentifier},
				{"->", EBNFLexerTypeArrow},
				{"{", EBNFLexerTypeLBrace},
				{`"parent"`, EBNFLexerTypeString},
				{":", EBNFLexerTypeColon},
				{"0", EBNFLexerTypeInteger},
				{"}", EBNFLexerTypeRBrace},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "dash in range",
			input: `"a"-"z"`,
			want: []ebnfExpectedToken{
				{`"a"`, EBNFLexerTypeString},
				{"-", EBNFLexerTypeDash},
				{`"z"`, EBNFLexerTypeString},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "comma and integers",
			input: "0, 12, 345",
			want: []ebnfExpectedToken{
				{"0", EBNFLexerTypeInteger},
				{",", EBNFLexerTypeComma},
				{"12", EBNFLexerTypeInteger},
				{",", EBNFLexerTypeComma},
				{"345", EBNFLexerTypeInteger},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "dot wildcard",
			input: "x . y",
			want: []ebnfExpectedToken{
				{"x", EBNFLexerTypeIdentifier},
				{".", EBNFLexerTypeDot},
				{"y", EBNFLexerTypeIdentifier},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "comment ignored",
			input: "# comment\nrule ::= x",
			want: []ebnfExpectedToken{
				{"rule", EBNFLexerTypeIdentifier},
				{"::=", EBNFLexerTypeAssign},
				{"x", EBNFLexerTypeIdentifier},
				{"", tokens.TokenTypeEOF},
			},
		},
		{
			name:  "identifier with leading underscore",
			input: "_foo ::= bar",
			want: []ebnfExpectedToken{
				{"_foo", EBNFLexerTypeIdentifier},
				{"::=", EBNFLexerTypeAssign},
				{"bar", EBNFLexerTypeIdentifier},
				{"", tokens.TokenTypeEOF},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewEBNFLexer(tt.input)
			for i, want := range tt.want {
				tok := lexer.Scan()
				if tok == nil {
					t.Fatalf("token %d: Scan() returned nil", i)
				}
				if want.typ == tokens.TokenTypeEOF {
					assert.True(t, tok.IsEOF(), "token %d: expected EOF", i)
					continue
				}
				assert.Equal(t, want.lexeme, tok.LexemeText(), "token %d lexeme", i)
				assert.Equal(t, want.typ, tok.Type, "token %d type", i)
			}
		})
	}
}

func TestEBNFLexer1(t *testing.T) {
	lexer := NewEBNFLexer("")

	token := lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestEBNFLexer2(t *testing.T) {
	lexer := NewEBNFLexer("rule ::= \"a\" | 'b' ;")

	token := lexer.Scan()
	assert.Equal(t, "rule", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "::=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "\"a\"", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "|", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeOr, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "'b'", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ";", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeSemicolon, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestEBNFLexer3(t *testing.T) {
	lexer := NewEBNFLexer("[ { ( ) } ] =")

	token := lexer.Scan()
	assert.Equal(t, "[", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLBracket, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "{", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLBrace, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "(", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLParen, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ")", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRParen, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "}", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRBrace, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "]", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRBracket, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)
}

func TestEBNFLexer4(t *testing.T) {
	// Standalone colon is now valid (used in AST hint blocks)
	lexer := NewEBNFLexer("x :=")

	token := lexer.Scan()
	assert.Equal(t, "x", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ":", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeColon, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestEBNFLexerArrow(t *testing.T) {
	lexer := NewEBNFLexer(`A ::= B -> { "parent" : 0 }`)

	token := lexer.Scan()
	assert.Equal(t, "A", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "::=", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeAssign, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "B", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeIdentifier, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "->", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeArrow, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "{", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeLBrace, token.Type)

	token = lexer.Scan()
	assert.Equal(t, `"parent"`, token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ":", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeColon, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "0", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "}", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeRBrace, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}

func TestEBNFLexerDashStillWorks(t *testing.T) {
	// Dash without > should still produce dash token (used in ranges)
	lexer := NewEBNFLexer(`"a"-"z"`)

	token := lexer.Scan()
	assert.Equal(t, `"a"`, token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "-", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeDash, token.Type)

	token = lexer.Scan()
	assert.Equal(t, `"z"`, token.LexemeText())
	assert.Equal(t, EBNFLexerTypeString, token.Type)
}

func TestEBNFLexerCommaAndIntegers(t *testing.T) {
	lexer := NewEBNFLexer("0, 12, 345")

	token := lexer.Scan()
	assert.Equal(t, "0", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ",", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeComma, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "12", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.Equal(t, ",", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeComma, token.Type)

	token = lexer.Scan()
	assert.Equal(t, "345", token.LexemeText())
	assert.Equal(t, EBNFLexerTypeInteger, token.Type)

	token = lexer.Scan()
	assert.True(t, token.IsEOF())
}
