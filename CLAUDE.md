# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

PGPG (Pretty Good Parser Generator) is a parser generator written in Go. It produces lexers (via
Thompson NFA→DFA construction) and LR(1) parsers from BNF grammar files. The project includes both
hand-written recursive-descent parsers and a full generator pipeline.

## Build Commands

```bash
# Build everything (go lib+generators+bin, apps/go/generated, apps/go) and run tests
make
make -C go test
make -C apps/go test

# Build and test individual parts
make -C go              # Build lib, generators, and install binaries to go/bin/
make -C go test         # Run go (lib + generators) tests
make -C apps/go/generated  # Generate lexers and parsers from BNF (uses go/bin/*)
make -C apps/go         # Build CLI runner tools (uses go.work → local go/)

# Format code
make -C go fmt
make -C apps/go/generated fmt
make -C apps/go fmt

# Static analysis (requires: go install honnef.co/go/tools/cmd/staticcheck@latest)
make -C go staticcheck  # if added to go/Makefile

# Pre-push check (fmt + build + test)
make -C go dev          # if added to go/Makefile
```

## Running a Single Test

```bash
cd go && go test ./lib/pkg/lexers/ -run TestEBNFLexer
cd go && go test ./generators/pkg/lexgen/ -run TestCodegen
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

The repo uses two Go modules; no `replace` directives. External repos (e.g. pgpg-experiments) depend on `github.com/johnkerl/pgpg/go`.

- **`go/`** — One module: `module github.com/johnkerl/pgpg/go`. Contains:
  - **`go/lib/`** — Core libraries (tokens, asts, lexers, parsers, util). Used by generators and by apps. Import: `github.com/johnkerl/pgpg/go/lib/pkg/...`
  - **`go/generators/`** — Code generation tools (lexgen, parsegen). Depends on go/lib.
  - **`go/bin/`** — Generator binaries (lexgen-tables, lexgen-code, parsegen-tables, parsegen-code), built by `make -C go`. Used by `apps/go/generated` Makefile.
- **`apps/go/`** — One module: `module github.com/johnkerl/pgpg/apps/go`. Depends on `github.com/johnkerl/pgpg/go`. Contains:
  - **`apps/go/generated/`** — Generated lexers/parsers (from BNF); part of this module (no separate go.mod). Makefile invokes `go/bin/*`.
  - **`apps/go/manual/`** — Hand-written sample lexers/parsers.
  - **`apps/go/cmd/`** — CLIs (trylex, tryparse, tryast).
- **`apps/go/go.work`** — Optional: `use .` and `use ../../go` so that builds in apps/go use the local `go/` module (CI and local dev).
- **`apps/jsons/`** — JSON tables produced by lexgen-tables/parsegen-tables.

### Generator Pipeline

```
BNF grammar file (.bnf)
    → go/bin/lexgen-tables, go/bin/parsegen-tables → JSON tables (intermediate)
    → go/bin/lexgen-code, go/bin/parsegen-code     → Generated Go source in apps/go/generated/
```

The JSON intermediate format is language-independent. The same pipeline can be driven in process: see **Using the generators as a library** below.

### Key Packages

- **`go/lib/pkg/tokens/`** — Token type, location tracking
- **`go/lib/pkg/lexers/`** — `AbstractLexer` interface, EBNF lexer, LookaheadLexer
- **`go/lib/pkg/parsers/`** — `AbstractParser` interface, EBNF parser
- **`go/lib/pkg/asts/`** — AST node structure, constructors, pretty-printing
- **`go/lib/pkg/util/`** — SplitString and other helpers
- **`go/generators/pkg/lexgen/`** — NFA→DFA lexer table + Go codegen (templates/lexer.go.tmpl)
- **`go/generators/pkg/parsegen/`** — LR(1) parser table + Go codegen (templates/parser.go.tmpl)
- **`go/generators/pkg/run/`** — `LexgenTables`, `LexgenCode`, `ParsegenTables`, `ParsegenCode`
- **`apps/bnfs/`** — Grammar files
- **`apps/go/generated/pkg/lexers/`**, **`apps/go/generated/pkg/parsers/`** — Auto-generated from apps/bnfs
- **`apps/go/cmd/`** — trylex, tryparse, tryast

### BNF Grammars

Grammar files live in `apps/bnfs/` (pemdas, lisp, json, seng, statements, pascal, etc.).

### Using the generators as a library

Other modules (e.g. pgpg-experiments) use the generators in process via `go get github.com/johnkerl/pgpg/go`. Import `github.com/johnkerl/pgpg/go/generators/pkg/lexgen`, `.../parsegen`, `.../run`. No `replace` needed. Library surface is:

- **`pkg/lexgen`** and **`pkg/parsegen`**: `GenerateTables(grammar, opts)`, `EncodeTables(tables, opts)`, `DecodeTables(data)`, `GenerateCode(tables, opts)`. All behavior is controlled by options structs; no globals.
- **`pkg/run`**: `LexgenTables`, `LexgenCode`, `ParsegenTables`, `ParsegenCode` — each does read → generate → write for one pipeline step; pass `""` or `"-"` as output path to write to stdout.

See **`go/generators/LIBRARY.md`** (if present) for option types and examples.

## Profiling

```bash
./go/bin/parsegen-tables \
  -cpuprofile cpu.pprof -memprofile mem.pprof -trace trace.out \
  -o output.json grammar.bnf
go tool pprof -http=:8082 cpu.pprof
```
