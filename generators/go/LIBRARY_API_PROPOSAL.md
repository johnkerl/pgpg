# Proposal: Library-Friendly API for lexgen/parsegen

This document proposes a simple, elegant codegen API so this package and others can use lexgen/parsegen as libraries (e.g. from `go generate` or custom tooling). Fresh dev, no compatibility constraints—optimize for clarity and usability.

---

## Design priorities

- **One clear way** to do each thing: options structs, no overloads or legacy shims.
- **No global state:** every behavior controlled by arguments to the call.
- **Obvious surface area:** tables step and codegen step each have a single entrypoint with an options struct.
- **Easy to drive:** optional runner that does read → generate → write so consumers can do one call per pipeline step.

---

## Target API

### 1. parsegen: no globals, options only

- **Remove** `var SortOutput`. Encoding behavior is per call only.

- **Tables (EBNF → JSON):**
  - `type ParseTableOptions struct { SourceName string }`  
    (optional; used for error messages. Empty = "").
  - `GenerateTables(grammar string, opts *ParseTableOptions) (*Tables, error)`  
    Single entrypoint. Replaces `GenerateTablesFromEBNF` and `GenerateTablesFromEBNFWithSourceName` with one function; `opts` can be nil and `SourceName` then treated as "".

  - `type EncodeOptions struct { Sort bool }`  
    `Sort`: deterministic JSON key order (default true when nil).
  - `EncodeTables(tables *Tables, opts *EncodeOptions) ([]byte, error)`  
    Single entrypoint. No one-arg overload; callers pass `nil` for default options.

- **Code (tables → Go):**
  - `type ParseCodegenOptions struct { Package string; Type string; Format bool }`  
    `Format`: run `go/format.Source` (default true).
  - `GenerateCode(tables *Tables, opts ParseCodegenOptions) ([]byte, error)`  
    Single entrypoint. Replaces `GenerateGoParserCode` and `GenerateGoParserCodeRaw`; raw output is `Format: false`.

- **Decode:** Keep `DecodeTables(data []byte) (*Tables, error)` as-is (no options needed).

---

### 2. lexgen: same shape as parsegen

- **Tables:**
  - `type LexTableOptions struct { SourceName string }`
  - `GenerateTables(grammar string, opts *LexTableOptions) (*Tables, error)`  
    Replaces the two existing EBNF entrypoints.

  - `type EncodeOptions struct { Sort bool }`  
    For consistency with parsegen, even if today lexgen always encodes deterministically. Enables future fast path or parity.
  - `EncodeTables(tables *Tables, opts *EncodeOptions) ([]byte, error)`  
    Single entrypoint; `nil` opts = default (e.g. Sort: true).

- **Code:**
  - `type LexCodegenOptions struct { Package string; Type string; Format bool }`
  - `GenerateCode(tables *Tables, opts LexCodegenOptions) ([]byte, error)`  
    Single entrypoint; replaces `GenerateGoLexerCode` and `GenerateGoLexerCodeRaw`.

- **Decode:** Keep `DecodeTables(data []byte) (*Tables, error)` as-is.

---

### 3. Naming consistency across packages

| Step        | parsegen              | lexgen                |
|------------|------------------------|------------------------|
| Grammar → tables | `GenerateTables(grammar, opts)` | `GenerateTables(grammar, opts)` |
| Tables → JSON    | `EncodeTables(tables, opts)`    | `EncodeTables(tables, opts)`    |
| JSON → tables   | `DecodeTables(data)`           | `DecodeTables(data)`           |
| Tables → Go     | `GenerateCode(tables, opts)`    | `GenerateCode(tables, opts)`   |

Same verb and pattern in both packages: **GenerateTables**, **EncodeTables**, **DecodeTables**, **GenerateCode**. Options structs carry package-specific flags (SourceName, Sort, Package, Type, Format).

---

### 4. Runner (optional but recommended)

A small package—e.g. `pkg/run` or `pkg/driver`—that does file I/O so the other package (or a `go generate` binary) can stay minimal:

- `LexgenTables(ctx context.Context, inputPath, outputPath string, opts *LexTableOptions) error`  
  Read BNF → `lexgen.GenerateTables` → `lexgen.EncodeTables` → write JSON. Options can include `SourceName` (e.g. inputPath) and `EncodeOptions` (e.g. Sort).

- `LexgenCode(ctx context.Context, tablesPath, outputPath string, opts LexCodegenOptions) error`  
  Read JSON → `lexgen.DecodeTables` → `lexgen.GenerateCode` → write Go.

- Same for parsegen: `ParsegenTables`, `ParsegenCode`.

One call per pipeline step; options structs passed through. `ctx` for cancellation and future-proofing. If you prefer to keep the generator module I/O-free, document the three-step flow (read → generate → write) and skip the runner; otherwise the runner is the most usable single addition.

---

### 5. Implementation notes (no code yet)

- **parsegen:** Move deterministic JSON logic into a helper. `Tables.MarshalJSON` either disappears or always uses that helper (deterministic). `EncodeTables(tables, opts)` uses the helper when `opts == nil || opts.Sort`, and `json.Marshal` via type alias when `opts != nil && !opts.Sort`. Then delete `SortOutput`.
- **lexgen:** If today there is no “fast” encoding path, `EncodeOptions` can still be added with `Sort` ignored (always deterministic) until you add a fast path; or omit `EncodeOptions` for lexgen until needed. Proposal keeps the same shape as parsegen for clarity.
- **CLIs:** Each cmd parses flags, builds the appropriate options struct(s), and calls the single library entrypoint. No globals, no branching on “which overload.”
- **Tests and callers:** One function per concern; options structs; zero magic.

---

### 6. Summary

| Package  | Remove / replace | Add |
|----------|-------------------|-----|
| parsegen | Remove `SortOutput`. Replace `GenerateTablesFromEBNF*`, `EncodeTables`, `GenerateGoParserCode*` with single-entrypoint API. | `ParseTableOptions`, `EncodeOptions`, `ParseCodegenOptions`; `GenerateTables`, `EncodeTables(tables, opts)`, `GenerateCode(tables, opts)`. |
| lexgen   | Replace `GenerateTablesFromEBNF*`, `EncodeTables`, `GenerateGoLexerCode*` with single-entrypoint API. | `LexTableOptions`, `EncodeOptions`, `LexCodegenOptions`; `GenerateTables`, `EncodeTables(tables, opts)`, `GenerateCode(tables, opts)`. |
| cmd/*    | Have each binary build options from flags and call the single API. | — |
| Optional | — | `pkg/run`: `LexgenTables`, `LexgenCode`, `ParsegenTables`, `ParsegenCode` (file I/O wrappers). |

Result: a simple, consistent, options-based interface with no globals and no compatibility shims—easy to use from this package or another.
