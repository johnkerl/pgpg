package parsers

import (
	"testing"

	"github.com/johnkerl/pgpg/generated/pkg/lexers"
)

func TestArithParserParsesSimpleExpression(t *testing.T) {
	lexer := lexers.NewArithLexer("1+2")
	parser := NewArithParser()
	ast, err := parser.Parse(lexer)
	if err != nil {
		t.Fatalf("expected parse success, got error: %v", err)
	}
	if ast == nil || ast.RootNode == nil {
		t.Fatal("expected non-nil AST")
	}
}

func TestArithWhitespaceParserParsesSimpleExpression(t *testing.T) {
	lexer := lexers.NewArithWhitespaceLexer("1+2")
	parser := NewArithWhitespaceParser()
	ast, err := parser.Parse(lexer)
	if err != nil {
		t.Fatalf("expected parse success, got error: %v", err)
	}
	if ast == nil || ast.RootNode == nil {
		t.Fatal("expected non-nil AST")
	}
}
