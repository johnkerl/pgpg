# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PGPG (Pretty Good Parser Generator) is a parser generator written in Go. It produces lexers (via Thompson NFA→DFA construction) and LR(1) parsers from BNF grammar files. The project includes both hand-written recursive-descent parsers and a full generator pipeline.

## Build Commands

```bash
# Build everything (manual, generator, generated, runners) and run tests
./build-all.sh

# Build and test individual modules
make -C manual          # Build manual module (core libraries)
make -C manual test     # Run manual tests
make -C generator       # Build generator executables
make -C generator test  # Run generator tests
make -C runners         # Build CLI runner tools

# Regenerate lexers/parsers from BNF grammars
cd generated && ./try-lexgen.sh && ./try-parsegen.sh

# Format code
make -C manual fmt
make -C generator fmt

# Static analysis (requires: go install honnef.co/go/tools/cmd/staticcheck@latest)
make -C generator staticcheck

# Pre-push check (fmt + build + test)
make -C generator dev
make -C manual dev
```

## Running a Single Test

```bash
cd manual && go test ./pkg/lexers/ -run TestPEMDASLexer
cd generator && go test ./pkg/lexgen/ -run TestCodegen
```

## Testing Parsers Interactively

```bash
# Manual (hand-written) parsers: prefix "m:"
./runners/tryparse m:pemdas expr '1*2+3'
./runners/tryparse m:vic expr 'x = x + 1'

# Generated parsers: prefix "g:"
./runners/tryparse g:pemdas expr '1+2*3'
./runners/tryparse g:json expr '{"a": [1, 2, 3]}'
./runners/tryparse g:lisp expr '(+ 1 (* 2 3))'

# Debug flags
./runners/tryparse -tokens -states -stack g:pemdas expr '1+2'

# Test lexers
./runners/trylex m:pemdas expr '1+2*3'
./runners/trylex g:pemdas expr '1+2*3'
```

## Architecture

The repo is a Go monorepo with four separate Go modules connected via `replace` directives:

```
manual/       → Core libraries (tokens, lexers, parsers, AST). No external deps except testify.
generator/    → Code generation tools. Depends on manual.
generated/    → Output of generator (pre-generated lexers/parsers from BNF grammars). Depends on manual.
runners/      → CLI tools (trylex, tryparse, tryast). Depends on manual + generated.
```

### Generator Pipeline

```
BNF grammar file (.bnf)
    → lexgen-tables / parsegen-tables → JSON tables (intermediate, language-independent)
    → lexgen-code / parsegen-code     → Generated Go source files
```

The JSON intermediate format is intentionally language-independent to allow future code generation targets beyond Go.

### Key Packages

- **`manual/pkg/tokens/`** — Token type, location tracking
- **`manual/pkg/lexers/`** — `AbstractLexer` interface + hand-written lexers (pemdas, vic, vbc, seng, ebnf, etc.)
- **`manual/pkg/parsers/`** — `AbstractParser` interface + hand-written recursive-descent parsers
- **`manual/pkg/asts/`** — AST node structure (Type, Token, Children), constructors, pretty-printing
- **`generator/pkg/lexgen/`** — NFA→DFA lexer table generation + Go code generation (uses `templates/lexer.go.tmpl`)
- **`generator/pkg/parsegen/`** — LR(1) parser table generation + Go code generation (uses `templates/parser.go.tmpl`)

### BNF Grammars

Grammar files live in `generated/bnfs/` (pemdas, lisp, json, seng, statements, pascal, etc.) and `grammar-check/bnfs/` (for validation with GOCC).

## Profiling

```bash
./generator/parsegen-tables \
  -cpuprofile cpu.pprof -memprofile mem.pprof -trace trace.out \
  -o output.json grammar.bnf
go tool pprof -http=:8082 cpu.pprof
```
