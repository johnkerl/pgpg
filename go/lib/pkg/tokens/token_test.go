package tokens

import (
	"testing"
)

func TestNewToken(t *testing.T) {
	loc := NewTokenLocation()
	tok := NewToken([]rune("abc"), TokenType("test"), loc)
	if tok == nil {
		t.Fatal("NewToken returned nil")
	}
	if string(tok.Lexeme) != "abc" || tok.Type != TokenType("test") {
		t.Errorf("NewToken: got lexeme=%q type=%s", string(tok.Lexeme), tok.Type)
	}
	if tok.Location.LineNumber != 1 || tok.Location.ColumnNumber != 1 {
		t.Errorf("NewToken: location = %+v", tok.Location)
	}
}

func TestNewEOFToken(t *testing.T) {
	loc := NewTokenLocation()
	tok := NewEOFToken(loc)
	if tok == nil {
		t.Fatal("NewEOFToken returned nil")
	}
	if !tok.IsEOF() || tok.Lexeme != nil || tok.Type != TokenTypeEOF {
		t.Errorf("NewEOFToken: IsEOF=%v lexeme=%v type=%s", tok.IsEOF(), tok.Lexeme, tok.Type)
	}
}

func TestNewErrorToken(t *testing.T) {
	loc := NewTokenLocation()
	tok := NewErrorToken("syntax error", loc)
	if tok == nil {
		t.Fatal("NewErrorToken returned nil")
	}
	if !tok.IsError() || string(tok.Lexeme) != "syntax error" || tok.Type != TokenTypeError {
		t.Errorf("NewErrorToken: IsError=%v lexeme=%q type=%s", tok.IsError(), string(tok.Lexeme), tok.Type)
	}
}

func TestIsEOF_IsError(t *testing.T) {
	loc := NewTokenLocation()
	eof := NewEOFToken(loc)
	err := NewErrorToken("err", loc)
	norm := NewToken([]rune("x"), TokenType("x"), loc)
	if !eof.IsEOF() || eof.IsError() {
		t.Error("EOF token: IsEOF should be true, IsError false")
	}
	if err.IsEOF() || !err.IsError() {
		t.Error("Error token: IsError should be true, IsEOF false")
	}
	if norm.IsEOF() || norm.IsError() {
		t.Error("Normal token: IsEOF and IsError should be false")
	}
}
