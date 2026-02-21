package main

import (
	"strconv"
	"testing"

	"github.com/johnkerl/goffl/pkg/f2poly"
	"github.com/johnkerl/goffl/pkg/f2polymod"
)

// TestF2PolyModPow2Mod13 verifies that in F2[x]/(x^4+x+1), the powers of x (hex poly 2)
// from 1 through 15 yield the expected residues (mod hex poly 13 = 0x13).
func TestF2PolyModPow2Mod13(t *testing.T) {
	modulus := f2poly.New(0x13) // x^4 + x + 1
	backend, err := NewF2PolyModNumeric(modulus)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"2", "4", "8", "3", "6", "c", "b", "5", "a", "7", "e", "f", "d", "9", "1"}

	for exp := 1; exp <= 15; exp++ {
		expr := "2**" + strconv.Itoa(exp)
		ast, err := parseWithMode(expr, "f2polymod")
		if err != nil {
			t.Fatalf("parse %q: %v", expr, err)
		}
		result, err := evaluateAST[*f2polymod.F2PolyMod, int](ast, backend, false)
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		got := backend.String(result)
		if got != want[exp-1] {
			t.Errorf("exponent %d: got %q, want %q", exp, got, want[exp-1])
		}
	}
}
