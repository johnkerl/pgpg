package lexgen

import (
	"bytes"
	"testing"
)

func TestGenerateTablesFromReader(t *testing.T) {
	grammar := `!ws ::= " " ; num ::= "0" | "1" ;`
	r := bytes.NewBufferString(grammar)
	tables, err := GenerateTablesFromReader(r, &LexTableOptions{SourceName: "test.bnf"})
	if err != nil {
		t.Fatalf("GenerateTablesFromReader: %v", err)
	}
	if tables == nil {
		t.Fatal("tables is nil")
	}
	if tables.StartState != 0 {
		t.Errorf("StartState: got %d", tables.StartState)
	}
	if len(tables.Actions) == 0 {
		t.Error("expected at least one action")
	}
}
