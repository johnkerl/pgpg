package parsegen

import (
	"bytes"
	"testing"
)

func TestGenerateTablesFromReader(t *testing.T) {
	grammar := `!ws ::= " " ; num ::= "0" | "1" ; Root ::= num ;`
	r := bytes.NewBufferString(grammar)
	tables, err := GenerateTablesFromReader(r, &ParseTableOptions{SourceName: "test.bnf"})
	if err != nil {
		t.Fatalf("GenerateTablesFromReader: %v", err)
	}
	if tables == nil {
		t.Fatal("tables is nil")
	}
	if tables.StartSymbol != "Root" {
		t.Errorf("StartSymbol: got %q", tables.StartSymbol)
	}
	if len(tables.Productions) == 0 {
		t.Error("expected at least one production")
	}
}
