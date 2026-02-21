package run

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnkerl/pgpg/generators/go/pkg/lexgen"
	"github.com/johnkerl/pgpg/generators/go/pkg/parsegen"
)

const minimalLexerBNF = `
!ws ::= " " | "\t" ;
num ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
plus ::= "+" ;
`

func TestLexgenTablesAndCode(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	bnfPath := filepath.Join(dir, "grammar.bnf")
	jsonPath := filepath.Join(dir, "tables.json")
	goPath := filepath.Join(dir, "lexer.go")

	if err := os.WriteFile(bnfPath, []byte(minimalLexerBNF), 0o644); err != nil {
		t.Fatalf("write BNF: %v", err)
	}

	if err := LexgenTables(ctx, bnfPath, jsonPath, nil); err != nil {
		t.Fatalf("LexgenTables: %v", err)
	}
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}
	if !strings.Contains(string(jsonData), "start_state") {
		t.Errorf("tables JSON missing start_state")
	}

	opts := lexgen.LexCodegenOptions{Package: "lexers", Type: "TestLexer", Format: true}
	if err := LexgenCode(ctx, jsonPath, goPath, opts); err != nil {
		t.Fatalf("LexgenCode: %v", err)
	}
	goData, err := os.ReadFile(goPath)
	if err != nil {
		t.Fatalf("read Go: %v", err)
	}
	if !strings.Contains(string(goData), "package lexers") {
		t.Errorf("generated Go missing package lexers")
	}
	if !strings.Contains(string(goData), "TestLexer") {
		t.Errorf("generated Go missing type TestLexer")
	}
}

// Minimal unambiguous grammar: one parser rule Root ::= num, plus lexer rules.
const minimalParserBNF = `
!ws ::= " " | "\t" ;
num ::= "0" | "1" ;
Root ::= num ;
`

func TestParsegenTablesAndCode(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	bnfPath := filepath.Join(dir, "grammar.bnf")
	jsonPath := filepath.Join(dir, "tables.json")
	goPath := filepath.Join(dir, "parser.go")

	if err := os.WriteFile(bnfPath, []byte(minimalParserBNF), 0o644); err != nil {
		t.Fatalf("write BNF: %v", err)
	}

	if err := ParsegenTables(ctx, bnfPath, jsonPath, nil); err != nil {
		t.Fatalf("ParsegenTables: %v", err)
	}
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("read JSON: %v", err)
	}
	if !strings.Contains(string(jsonData), "start_symbol") {
		t.Errorf("tables JSON missing start_symbol")
	}

	opts := parsegen.ParseCodegenOptions{Package: "parsers", Type: "TestParser", Format: true}
	if err := ParsegenCode(ctx, jsonPath, goPath, opts); err != nil {
		t.Fatalf("ParsegenCode: %v", err)
	}
	goData, err := os.ReadFile(goPath)
	if err != nil {
		t.Fatalf("read Go: %v", err)
	}
	if !strings.Contains(string(goData), "package parsers") {
		t.Errorf("generated Go missing package parsers")
	}
	if !strings.Contains(string(goData), "TestParser") {
		t.Errorf("generated Go missing type TestParser")
	}
}

func TestWriteOutputToStdout(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	bnfPath := filepath.Join(dir, "grammar.bnf")
	if err := os.WriteFile(bnfPath, []byte(minimalLexerBNF), 0o644); err != nil {
		t.Fatalf("write BNF: %v", err)
	}
	// LexgenTables with outputPath "-" should not create a file; we just check it doesn't error.
	if err := LexgenTables(ctx, bnfPath, "-", nil); err != nil {
		t.Fatalf("LexgenTables to stdout: %v", err)
	}
}
