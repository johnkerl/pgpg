# pgpg

PGPG is the Pretty Good Parser Generator.

As of Feburary 2023 it is very much a sketch and a work in progress.

As of February 2026 I'm picking this back up again, making significant use of Cursor.

## Goals

* Implement a few basic algorithms.
* Reuse code whenever possible
  * Across multiple algorithms like LALR/LR
* Make good use of classes---e.g. `lexer.match()` rather than global `match()` which are commonly used in intro-to-parsing textbooks.
* Be lucid above all else. Lexing/parsing is ubiquitous in the modern world, and forms a large part of our world. Yet sadly such tools are too often arcane and confusing. PGPG is transparent, inclusive, and explains itself openly.
* Offer choices.
  * Sometimes a parser-generator is overkill---for simpler grammars, a hand-written lexer and a hand-written recursive-descent parser are quite satisfactory. PGPG offers reusable, easy-to-understand examples here.
  * Sometimes a hand-written lexer/parser is underkill---yet parser-generators can be complex and intimidating. Here, too, PGPG offers reusable, easy-to-understand examples.
  * PGPG offers classes that reduce code-duplication for various lex/parse implementations: you can reuse what you want, and hand-write what you want.
  * PGPG offers grammar-to-parser all in one process invocation, or parser-generate to language-independent storage (probably JSON), or traditional parser-generate directly to implementation-language code.

## Languages

* Implementation initially in Go
  * Maybe Python and/or JavaScript and/or Rust later
  * Aim for non-clever abstraction and concept reuse
  * Try to use language-independent data structures when possible
* Generator initially in Go
  * Maybe Python and/or JavaScript and/or Rust later
  * Try to use language-independent data structures when possible

## Applications

* Self-education and experimentation
* Promotion of parser-generation knowledge
* I would like to ultimately use this in [Miller](https://github.com/johnkerl/miller)
* I'd love to get the latency lowered and flexibility increased to the point where I can
  simply play around with language design at will.

## Build commands

```bash
# Build everything (manual, generator, generated, apps/go) and run tests
make
make -C manual test
make -C generators/go test

# Build and test individual modules
make -C manual          # Build manual module (core libraries)
make -C manual test     # Run manual tests
make -C generators/go     # Build generator executables
make -C generators/go test  # Run generator tests
make -C generated       # Generate lexers and parsers from BNF source
make -C apps/go         # Build CLI runner tools

# Format code
make -C manual fmt
make -C generators/go fmt
make -C generated fmt
make -C apps/go fmt

# Static analysis (requires: go install honnef.co/go/tools/cmd/staticcheck@latest)
make -C generators/go staticcheck

# Pre-push check (fmt + build + test)
make -C manual dev
make -C generators/go dev
```

## Running a single test

```bash
cd manual    && go test ./go/pkg/lexers/ -run TestPEMDASLexer
cd generators/go && go test ./pkg/lexgen/ -run TestCodegen
```

## Testing parsers interactively

```bash
# Manual (hand-written) parsers: prefix "m:"
./apps/go/tryparse m:pemdas expr '1*2+3'
./apps/go/tryparse m:vic expr 'x = x + 1'

# Generated parsers: prefix "g:"
./apps/go/tryparse g:pemdas expr '1+2*3'
./apps/go/tryparse g:json expr '{"a": [1, 2, 3]}'
./apps/go/tryparse g:lisp expr '(+ 1 (* 2 3))'

# Debug flags
./apps/go/tryparse -tokens -states -stack g:pemdas expr '1+2'

# Test lexers
./apps/go/trylex m:pemdas expr '1+2*3'
./apps/go/trylex g:pemdas expr '1+2*3'
```
