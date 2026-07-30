package run

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnkerl/pgpg/go/generators/pkg/lexgen"
	"github.com/johnkerl/pgpg/go/generators/pkg/parsegen"
)

// TestGoldenTables regenerates lexer and parser tables for every grammar with
// checked-in JSON under apps/jsons and requires byte-identical output. This
// guards the JSON interchange format: table-encoding refactors must not change
// a single byte, and generator behavior changes must come with regenerated
// goldens. The Makefile alone cannot catch generator-induced staleness, since
// it only rebuilds a JSON file when its .bnf is newer.
func TestGoldenTables(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	jsonsDir := filepath.Join(repoRoot, "apps", "jsons")
	bnfsDir := filepath.Join(repoRoot, "apps", "bnfs")

	entries, err := os.ReadDir(jsonsDir)
	if err != nil {
		// The apps/ tree is absent when this module is tested outside the
		// pgpg repo (e.g. as a downloaded module); there is nothing to check.
		t.Skipf("skipping: cannot read golden dir: %v", err)
	}

	checked := 0
	for _, entry := range entries {
		name := entry.Name()
		var kind string
		var base string
		switch {
		case strings.HasSuffix(name, "-lex.json"):
			kind = "lex"
			base = strings.TrimSuffix(name, "-lex.json")
		case strings.HasSuffix(name, "-parse.json"):
			kind = "parse"
			base = strings.TrimSuffix(name, "-parse.json")
		default:
			continue
		}
		checked++

		t.Run(name, func(t *testing.T) {
			grammarBytes, err := os.ReadFile(filepath.Join(bnfsDir, base+".bnf"))
			if err != nil {
				t.Fatalf("read grammar: %v", err)
			}
			golden, err := os.ReadFile(filepath.Join(jsonsDir, name))
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}

			var generated []byte
			if kind == "lex" {
				tables, err := lexgen.GenerateTables(string(grammarBytes), nil)
				if err != nil {
					t.Fatalf("generate lexer tables: %v", err)
				}
				generated, err = lexgen.EncodeTables(tables, nil)
				if err != nil {
					t.Fatalf("encode lexer tables: %v", err)
				}
			} else {
				tables, err := parsegen.GenerateTables(string(grammarBytes), nil)
				if err != nil {
					t.Fatalf("generate parser tables: %v", err)
				}
				generated, err = parsegen.EncodeTables(tables, nil)
				if err != nil {
					t.Fatalf("encode parser tables: %v", err)
				}
			}

			if string(generated) != string(golden) {
				t.Errorf("generated tables differ from checked-in %s; "+
					"if the generator change is intentional, regenerate with "+
					"make -C apps/go/generated after touching the .bnf files",
					name)
			}
		})
	}

	if checked == 0 {
		t.Fatalf("no golden files found in %s", jsonsDir)
	}
}
