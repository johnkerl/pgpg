# Using the generators as a library

Other Go packages (e.g. a `go generate` driver or custom tooling) can use the lexer and parser generators in process instead of invoking the `lexgen-tables`, `lexgen-code`, `parsegen-tables`, and `parsegen-code` binaries.

## Module dependency

Add the generators module to your `go.mod`:

```go
require github.com/johnkerl/pgpg/generators/go v0.0.0
require github.com/johnkerl/pgpg/lib v0.0.0  // indirect, pulled in by generators

replace github.com/johnkerl/pgpg/lib => /path/to/pgpg/lib
replace github.com/johnkerl/pgpg/generators/go => /path/to/pgpg/generators/go
```

For local development inside the same repo, use relative `replace` directives so both modules point at your working copy. Once the module is published (e.g. to a module proxy), you can depend on a version and drop the `replace` for the generators; you may still need a `replace` for `lib` if it is not published separately.

## Library surface

- **`github.com/johnkerl/pgpg/generators/go/pkg/lexgen`** — Lexer table generation and Go codegen. Options structs: `LexTableOptions`, `EncodeOptions`, `LexCodegenOptions`. Entrypoints: `GenerateTables`, `GenerateTablesFromReader` (grammar from `io.Reader`), `EncodeTables`, `DecodeTables`, `GenerateCode`.
- **`github.com/johnkerl/pgpg/generators/go/pkg/parsegen`** — Parser table generation and Go codegen. Options structs: `ParseTableOptions`, `EncodeOptions`, `ParseCodegenOptions`. Entrypoints: `GenerateTables`, `GenerateTablesFromReader` (grammar from `io.Reader`), `EncodeTables`, `DecodeTables`, `GenerateCode`.
- **`github.com/johnkerl/pgpg/generators/go/pkg/run`** — File I/O wrappers: one call per pipeline step. `LexgenTables`, `LexgenCode`, `ParsegenTables`, `ParsegenCode`. Pass `context.Context` for cancellation; use `""` or `"-"` as output path to write to stdout.

All options may be `nil` where pointers are used; that means “use defaults” (e.g. deterministic JSON, formatted Go). No global state; every behavior is controlled by the arguments to the call.

## Option types (quick reference)

| Package  | Tables (grammar → JSON)     | Code (tables → Go)        |
|----------|-----------------------------|----------------------------|
| lexgen   | `LexTableOptions{SourceName}`; `EncodeOptions{Sort}` | `LexCodegenOptions{Package, Type, Format}` |
| parsegen | `ParseTableOptions{SourceName}`; `EncodeOptions{Sort}` | `ParseCodegenOptions{Package, Type, Format}` |

`Sort: true` (or nil) gives deterministic JSON key order; `Sort: false` is faster but nondeterministic. `Format: true` runs `go/format.Source` on generated Go; `Format: false` returns unformatted source.

## Example: low-level (in-memory)

```go
import (
    "github.com/johnkerl/pgpg/generators/go/pkg/lexgen"
)

func generateLexer(grammar string, sourceName string) (goCode []byte, err error) {
    tables, err := lexgen.GenerateTables(grammar, &lexgen.LexTableOptions{SourceName: sourceName})
    if err != nil {
        return nil, err
    }
    return lexgen.GenerateCode(tables, lexgen.LexCodegenOptions{
        Package: "mypkg",
        Type:    "MyLexer",
        Format:  true,
    })
}
```

When the grammar comes from an `io.Reader` (e.g. HTTP body, open file, `bytes.Buffer`), use `GenerateTablesFromReader` instead of `GenerateTables`; same options. Same in parsegen: `GenerateTablesFromReader(r, opts)`.

## Example: runner (file paths, one call per step)

```go
import (
    "context"
    "github.com/johnkerl/pgpg/generators/go/pkg/run"
    "github.com/johnkerl/pgpg/generators/go/pkg/lexgen"
)

func generateFromFiles(ctx context.Context, bnfPath, jsonPath, goPath string) error {
    if err := run.LexgenTables(ctx, bnfPath, jsonPath, nil); err != nil {
        return err
    }
    return run.LexgenCode(ctx, jsonPath, goPath, lexgen.LexCodegenOptions{
        Package: "mypkg",
        Type:    "MyLexer",
        Format:  true,
    })
}
```

## go generate

You can depend on this module and call `run.LexgenTables` / `run.LexgenCode` (and the parsegen equivalents) from a small `main` package that is invoked by `//go:generate go run ./cmd/codegen` (or similar). That keeps your generate script in Go and avoids shell or Make for the pipeline.
