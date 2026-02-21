# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PGPG (Pretty Good Parser Generator) is a parser generator written in Go. It produces lexers (via
Thompson NFA→DFA construction) and LR(1) parsers from BNF grammar files. The project includes both
hand-written recursive-descent parsers and a full generator pipeline.

## Build Commands

```bash
# Build everything (manual, generator, apps/generated, apps/go) and run tests
make
make -C manual test
make -C generators/go test

# Build and test individual modules
make -C manual          # Build manual module (core libraries)
make -C manual test     # Run manual tests
make -C generators/go     # Build generator executables
make -C generators/go test  # Run generator tests
make -C apps/generated  # Generate lexers and parsers from BNF source
make -C apps/go         # Build CLI runner tools

# Format code
make -C manual fmt
make -C generators/go fmt
make -C apps/generated fmt
make -C apps/go fmt

# Static analysis (requires: go install honnef.co/go/tools/cmd/staticcheck@latest)
make -C generators/go staticcheck

# Pre-push check (fmt + build + test)
make -C manual dev
make -C generators/go dev
```

## Running a Single Test

```bash
cd manual    && go test ./go/pkg/lexers/ -run TestPEMDASLexer
cd generators/go && go test ./pkg/lexgen/ -run TestCodegen
```

## Testing Parsers Interactively

```bash
# Manual (hand-written) parsers: prefix "m:"
./apps/go/tryparse -e m:pemdas '1*2+3'
./apps/go/tryparse -e m:vic 'x = x + 1'

# Generated parsers: prefix "g:"
./apps/go/tryparse -e g:pemdas '1+2*3'
./apps/go/tryparse -e g:json '{"a": [1, 2, 3]}'
./apps/go/tryparse -e g:lisp '(+ 1 (* 2 3))'

# Debug flags (flags before parser name)
./apps/go/tryparse -tokens -states -stack -e g:pemdas '1+2'

# Test lexers
./apps/go/trylex -e m:pemdas '1+2*3'
./apps/go/trylex -e g:pemdas '1+2*3'
```

## Architecture

The repo is a Go monorepo with four separate Go modules connected via `replace` directives:

```
manual/         → Core libraries (tokens, lexers, parsers, AST). No external deps except testify.
generators/go/  → Code generation tools. Depends on manual.
apps/generated/ → Output of generators/go (pre-generated lexers/parsers from BNF grammars). Depends on manual.
apps/go/        → CLI tools (trylex, tryparse, tryast). Depends on manual + generated.
```

### Generator Pipeline

```
BNF grammar file (.bnf)
    → lexgen-tables / parsegen-tables → JSON tables (intermediate, language-independent)
    → lexgen-code / parsegen-code     → Generated Go source files
```

The JSON intermediate format is intentionally language-independent to allow future code generation targets beyond Go. The same pipeline can be driven in process: see **Using the generators as a library** below.

### Key Packages

- **`manual/go/pkg/tokens/`** — Token type, location tracking
- **`manual/go/pkg/lexers/`** — `AbstractLexer` interface + hand-written lexers (pemdas, vic, vbc, seng, ebnf, etc.)
- **`manual/go/pkg/parsers/`** — `AbstractParser` interface + hand-written recursive-descent parsers
- **`manual/go/pkg/asts/`** — AST node structure (Type, Token, Children), constructors, pretty-printing
- **`generators/go/pkg/lexgen/`** — NFA→DFA lexer table generation + Go code generation (uses `templates/lexer.go.tmpl`)
- **`generators/go/pkg/parsegen/`** — LR(1) parser table generation + Go code generation (uses `templates/parser.go.tmpl`)
- **`generators/go/pkg/run/`** — File I/O wrappers for one-call-per-step: `LexgenTables`, `LexgenCode`, `ParsegenTables`, `ParsegenCode`
- **`apps/bnfs/`** — Grammar files to have lexers/parsers generated from
- **`apps/generated/go/pkg/lexers/`** — Auto-generated lexers from `apps/bnfs/`
- **`apps/generated/go/pkg/parsers/`** — Auto-generated parsers from `apps/bnfs/`
- **`apps/go/cmd/`** — CLIs to interactively test-drive the manual and generated lexers and parsers.

### BNF Grammars

Grammar files live in `apps/bnfs/` (pemdas, lisp, json, seng, statements, pascal, etc.).

### Using the generators as a library

Other packages can use the generators in process (e.g. from `go generate`) instead of calling the CLI binaries. Add `github.com/johnkerl/pgpg/generators/go` to your module’s dependencies (and `replace` for local dev). The library surface is:

- **`pkg/lexgen`** and **`pkg/parsegen`**: `GenerateTables(grammar, opts)`, `EncodeTables(tables, opts)`, `DecodeTables(data)`, `GenerateCode(tables, opts)`. All behavior is controlled by options structs; no globals.
- **`pkg/run`**: `LexgenTables`, `LexgenCode`, `ParsegenTables`, `ParsegenCode` — each does read → generate → write for one pipeline step; pass `""` or `"-"` as output path to write to stdout.

See **`generators/go/LIBRARY.md`** for dependency setup, option types, and examples.

## Profiling

```bash
./generators/go/parsegen-tables \
  -cpuprofile cpu.pprof -memprofile mem.pprof -trace trace.out \
  -o output.json grammar.bnf
go tool pprof -http=:8082 cpu.pprof
```
