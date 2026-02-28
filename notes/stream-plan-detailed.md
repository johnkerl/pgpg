---
name: "Stream plan: io.Reader lexer + ParseOne"
overview: "Two-step plan: (1) Change all lexers to take io.Reader with a string-backed helper API and update CLIs to use file-backed vs string-backed readers; (2) Add accept-and-yield and ParseOne so the same .bnf supports both single-object (Parse, accept only on EOF) and multi-object (ParseOne loop) parsing, with tests and app examples."
todos: []
isProject: false
---

# Stream plan: io.Reader lexer + multi/single-object parse

## Step 1: Lexer takes `io.Reader` (then stop and chat)

Step 1 must implement **true streaming**: no `io.ReadAll`. The lexer reads incrementally from `io.Reader` so that (1) huge streams (e.g. billions of objects) stay bounded in memory, and (2) `tail -f` / infinite input can be handled as data arrives.

### 1.0 What would need to change without ReadAll (streaming implementation)

Current implementations assume the whole input is in memory as a string and use **byte-offset-based access** and **slicing**:

- **EBNFLexer** ([go/lib/pkg/lexers/ebnf_lexer.go](go/lib/pkg/lexers/ebnf_lexer.go)):
  - **peekRune()**: today does `utf8.DecodeRuneInString(lexer.inputText[lexer.tokenLocation.ByteOffset:])` — i.e. random access by byte offset into a full string.
  - **EOF**: `tokenLocation.ByteOffset >= lexer.inputLength` — requires knowing total length.
  - **Lexeme building**: It does **not** slice `inputText` for the lexeme. It advances with `peekRune` + `LocateRune` and builds tokens by **appending runes** (identifiers, integers, string literals). So the only real dependency is "next rune at current position" and "have we reached end of input?"
  - **Change for streaming**: Replace the in-memory string with a `bufio.Reader` and track position (line/column/byteOffset) as we go. Implement **peek** as: read one rune via `bufio.Reader.ReadRune()`, then `UnreadRune()` so the next read sees it again; or keep a one-rune peek buffer (read into buffer, "advance" by clearing buffer and updating position). So: no `inputLength`; EOF = `ReadRune` returns `io.EOF`. Lexeme is already built rune-by-rune, so no slicing changes.
- **Generated lexers** ([go/generators/pkg/lexgen/templates/lexer.go.tmpl](go/generators/pkg/lexgen/templates/lexer.go.tmpl)):
  - **DFA loop**: They maintain `tokenLocation` (start of token) and `scanLocation` (current probe). The inner loop calls `peekRuneAt(scanLocation.ByteOffset)` — i.e. **random access at an arbitrary offset** within the token being scanned. When no transition applies, they **backtrack** to `lastAcceptLocation` and **slice** `lexer.inputText[lexer.tokenLocation.ByteOffset:lastAcceptLocation.ByteOffset]` to form the lexeme.
  - So the generated lexer needs: (1) to read runes **ahead** of the token start (scanLocation can be many bytes past tokenLocation), and (2) to **rewind** to `lastAcceptLocation` after the scan and extract that segment. An `io.Reader` cannot be rewound.
  - **Change for streaming**: Use a **sliding buffer** of bytes/runes that is refilled from the reader as we scan. **Option B — sliding window**: Keep a byte/rune buffer; when we need more input we read from the reader and append; when we accept we take the slice `buffer[tokenStart:acceptEnd]` as the lexeme and **discard** that prefix from the buffer (no rewind). We never unread; we only advance the reader. So we need a buffer that holds "everything read but not yet consumed". For bounded token length (typical for JSON, etc.) the buffer stays bounded; for `tail -f` we only keep one token's worth. Position (line/column) is updated as we consume from the buffer. Option B is preferred for efficiency: each rune is decoded and run through the DFA once, with no re-scanning of the spill. Concrete changes: replace `inputText`/`inputLength` with a `bufio.Reader` plus a **consumed-from-buffer** rune (or byte) slice for the current token; implement `peekRuneAt(offset)` as "rune at offset within that buffer" (refill buffer from reader when needed); on accept, set lexeme from buffer and advance "consumed" region (drop prefix). EOF when reader returns EOF and buffer is exhausted.

So in short:

- **EBNFLexer**: No slicing; only "next rune" and EOF. Streaming = `bufio.Reader` + peek (read+unread or one-rune buffer) + track position; no structural change to token-building.
- **Generated lexers**: Use a **sliding-window buffer** (Option B) filled from the reader. On accept, set lexeme from buffer and discard prefix; no unread. Refill when needed. Position tracking updated as we consume. More efficient than unread (no re-scanning of spill).

### 1.1 Core lexer API (go/lib and generated)

- **AbstractLexer**: unchanged; constructors take `io.Reader`.
- **EBNFLexer**:
  - Constructors: `NewEBNFLexer(r io.Reader)` and `NewEBNFLexerWithSourceName(r io.Reader, sourceName string)`.
  - **Streaming implementation**: Hold a `*bufio.Reader` (and optional sourceName). Replace `inputText`/`inputLength` and `peekRune()` with: read runes via the reader, use a one-rune peek buffer or `ReadRune`+`UnreadRune` for peek semantics; EOF when read returns `io.EOF`; keep building lexemes by appending runes as today. Update `TokenLocation` (line/column/byteOffset) as runes are consumed.
  - **String-backed API**: `NewEBNFLexerFromString(s string) AbstractLexer` = `NewEBNFLexer(strings.NewReader(s))` (and same for WithSourceName).
- **Generated lexers**:
  - Constructor: `New{{.TypeName}}(r io.Reader) liblexers.AbstractLexer`.
  - **Streaming implementation (sliding window, Option B)**: Replace `inputText`/`inputLength` with a sliding-window buffer fed from `bufio.Reader` (sliding window, Option B). Replace `peekRuneAt(byteOffset int)` with "rune at logical offset in buffer"; refill from reader when that offset is past the end. On accept: lexeme = buffer[tokenStart:acceptEnd], then discard prefix (`buffer = buffer[acceptEnd:]`). No unread. EOF when reader is at EOF and buffer is exhausted. Position tracking updated when consuming from buffer.
  - **String-backed API**: Add in template `New{{.TypeName}}FromString(s string) liblexers.AbstractLexer { return New{{.TypeName}}(strings.NewReader(s)) }` for tests and -e.

### 1.2 Call sites that currently lex from string

- **Unit tests**
  - [go/lib/pkg/lexers/ebnf_lexer_test.go](go/lib/pkg/lexers/ebnf_lexer_test.go): replace `NewEBNFLexer(tt.input)` with `NewEBNFLexer(strings.NewReader(tt.input))` or `NewEBNFLexerFromString(tt.input)`; same for other `NewEBNFLexer(...)` calls.
  - [go/lib/pkg/parsers/ebnf_parser_test.go](go/lib/pkg/parsers/ebnf_parser_test.go): parser currently `Parse(inputText string)`. See 1.3.
  - Generated lexer/parser tests (if any in apps/go/generated or go/generators): use `NewXxxLexer(strings.NewReader(s))` or generated `NewXxxLexerFromString(s)`.
- **EBNF parser** ([go/lib/pkg/parsers/ebnf_parser.go](go/lib/pkg/parsers/ebnf_parser.go)):
  - Change `Parse(inputText string)` to `Parse(r io.Reader) (*asts.AST, error)`.
  - Inside: `parser.lexer = lexers.NewLookaheadLexer(lexers.NewEBNFLexerWithSourceName(r, parser.sourceName))` (EBNFLexer will take `io.Reader`). If EBNFLexer still needs a name and we only have a reader, we can keep `NewEBNFLexerWithSourceName(r, parser.sourceName)` and pass a type that implements `io.Reader` (e.g. we could add a named-reader wrapper or keep sourceName for errors only).
- **Manual parsers** ([go/lib/pkg/parsers/abstract_parser.go](go/lib/pkg/parsers/abstract_parser.go) and [apps/go/manual/parsers/*.go](apps/go/manual/parsers/)):
  - Change `AbstractParser` to `Parse(r io.Reader) (*asts.AST, error)`.
  - Each manual parser (AME, AMNE, PEMDAS, VIC, VBC) and EBNFParser: constructor unchanged; `Parse(r io.Reader)` builds the lexer from `r` (e.g. `lexers.NewVICLexer(r)`) and runs the parse. Manual lexers in [apps/go/manual/lexers/](apps/go/manual/lexers/) (VIC, VBC, AM, Line, CannedText, Rune, Word, SENG) need constructors changed to take `io.Reader`. For true streaming (no ReadAll), each implements the same pattern: `bufio.Reader`, read/peek runes, build lexemes by appending runes, track position; no full-input slice.
- **pemdas-eval** ([apps/go/cmd/pemdas-eval/main.go](apps/go/cmd/pemdas-eval/main.go)): currently uses `NewPEMDASIntLexer(input)` etc. with a string. Switch to `NewXxxLexer(strings.NewReader(input))` or the generated `FromString` helper.

### 1.3 Apps CLIs: file-backed vs string-backed io.Reader

- **tryparse** ([apps/go/cmd/tryparse/main.go](apps/go/cmd/tryparse/main.go)):
  - Change `run` signature from `func(string, traceOptions)` to `func(io.Reader, traceOptions)` (or keep a single type for both: `io.Reader`).
  - **File mode**: for each file, `f, _ := os.Open(filename)` and call `run(f, opts)`, then close. No `ReadFile`; pass the `*os.File` as `io.Reader`.
  - **Stdin**: call `run(os.Stdin, opts)` (no ReadAll; true stream) or keep ReadAll and `run(strings.NewReader(string(content)), opts)` for consistency with "one blob" behavior—prefer passing `os.Stdin` for file-backed semantics.
  - **-e mode**: for each expression arg, call `run(strings.NewReader(arg), opts)`.
  - **runGeneratedParser**: takes `newLexer func(io.Reader) liblexers.AbstractLexer`; inside, `lexer := newLexer(r)` then `parser.Parse(lexer, opts.astMode)`.
  - **runManualParser**: takes parser that implements `Parse(r io.Reader)`; call `parser.Parse(r)`.
- **trylex** ([apps/go/cmd/trylex/main.go](apps/go/cmd/trylex/main.go)):
  - Change `lexerMaker` to `func(io.Reader) liblexers.AbstractLexer` (and fix type to use `liblexers` if needed).
  - **-e mode**: `runLexerOnce(lexerMaker, strings.NewReader(arg))`.
  - **Stdin**: `runLexerOnce(lexerMaker, os.Stdin)`.
  - **File mode**: for each file, open with `os.Open`, then `runLexerOnce(lexerMaker, f)` and close. Use **one lexer over the whole file** (remove the current line-by-line loop with `bufio.Scanner`); the lexer sees the entire file content as one stream.
- **tryast**: if it uses the same run/lexer pattern, mirror the same `io.Reader`-based API.

Result: CLIs use file-backed `io.Reader` for file mode and stdin, and string-backed `io.Reader` (e.g. `strings.NewReader(s)`) for -e mode.

### 1.4 Regenerate and verify

- After changing the lexer template, run the apps/go/generated Makefile to regenerate all lexers (and fix any callers in apps that still pass string).
- Run `make -C go test`, `make -C apps/go test`, and manual smoke tests: `trylex -e g:json '{"a":1}'`, `tryparse -e g:json '{"a":1}'`, and file-based `trylex g:json file.json` / `tryparse g:json file.json`.

---

## Step 2: Same .bnf for single-object and multi-object parse

Goal: one grammar file, no source mods; single-object = accept only on EOF; multi-object = accept after one record and continue (ParseOne loop). Example: `{}{}` → single-object: error; multi-object: two successful JSON parses.

### 2.1 Table generator (parsegen-tables)

- **Accept-and-yield for non-EOF** ([go/generators/pkg/parsegen/tables.go](go/generators/pkg/parsegen/tables.go)):
  - After building `actions` as today, add a pass: for every state `s` that has `actions[s][eofSymbol].Type == "accept"`, and for every **terminal** in the grammar (including EOF), if terminal is **not** EOF and `actions[s][terminal]` is **not** already set, set `actions[s][terminal] = Action{Type: "accept_and_yield"}`.
  - Terminals list: derive from grammar (e.g. all symbols that appear as RHS and are terminal, or from a dedicated grammar.terminals set). Ensure we do not overwrite existing shift/reduce in that state (in the accept state they only have accept on EOF, so no conflict).
  - JSON encoding: extend [Action](go/generators/pkg/parsegen/tables.go) to include the new type; tables already have `Type` string, so `"accept_and_yield"` is enough.

### 2.2 Code generator (parsegen-code)

- **New action kind** ([go/generators/pkg/parsegen/codegen.go](go/generators/pkg/parsegen/codegen.go) and [templates/parser.go.tmpl](go/generators/pkg/parsegen/templates/parser.go.tmpl)):
  - Add `{{.TypeName}}ActionAcceptAndYield` to the action kind enum and to `actionKindLiteral` (e.g. case `"accept_and_yield"`).
  - **Parser struct**: add `stashedLookahead *tokens.Token`.
  - **Parse(lexer, astMode)** (existing): keep unchanged; only return on `ActionAccept` (EOF). Do **not** handle `AcceptAndYield` in the main `Parse` loop (so if we ever land on AcceptAndYield in Parse, treat as "no action" or error—better: in the table, only accept state has accept_and_yield, and Parse only stops on Accept; so in Parse we never look at accept_and_yield. So in the switch we only handle Shift, Reduce, Accept; if we see AcceptAndYield we could return an error like "use ParseOne for multi-record input" or simply not emit AcceptAndYield in the action map when we're in a "single-object-only" build; but the plan is one table for both, so Parse should ignore accept_and_yield by not stopping on it—then we'd loop and call lexer.Scan() again, which would consume the next token and likely cause a parse error. So the clean approach: in **Parse**, when we see AcceptAndYield, treat it as "parse error: multiple objects not allowed" or equivalent so single-object mode is strict. Alternatively, have Parse only look up actions that are Shift, Reduce, or Accept—so the table still has accept_and_yield for ParseOne, but Parse's loop only breaks on Accept. So in the generated Parse(), the switch has case Accept: return; case AcceptAndYield: return error "multi-record input; use ParseOne". That way single-document callers get a clear error if they pass multi-object input.)
  - **ParseOne(lexer, astMode) (ast *asts.AST, done bool, err error)**:
    - First token: if `parser.stashedLookahead != nil`, use it as lookahead and set `parser.stashedLookahead = nil`; else `lookahead = lexer.Scan()`.
    - Loop: same shift/reduce as Parse; on **Accept**: return `(ast, true, nil)`; on **AcceptAndYield**: set `parser.stashedLookahead = lookahead`, return `(ast, false, nil)`; do not advance past lookahead.
    - On error: return `(nil, false, err)`.
- **Backward compatibility**: Keep `Parse(lexer, astMode)`; existing callers that expect single-document behavior stay as-is. Document that multi-record input should use ParseOne in a loop.

### 2.3 Tests

- **Unit tests** (e.g. in [go/generators/pkg/parsegen](go/generators/pkg/parsegen) or [apps/go/generated](apps/go/generated) or a new test package):
  - Single-object: `Parse(lexer, "")` on input `"{}"` → success; on `"{}{}"` → error (or defined behavior: "multiple objects").
  - Multi-object: one lexer from `strings.NewReader("{}{}")` (or from string), then first `ParseOne(lexer, "")` → `(ast1, false, nil)`, second `ParseOne(lexer, "")` → `(ast2, true, nil)`. Optionally test three objects and EOF.
- **Apps examples**:
  - Use **tryparse** with g:json (or g:json-plain): single-object mode is default (Parse), so `tryparse -e g:json '{}{}'` should report a parse error.
  - Add a **multi-object mode** (e.g. flag `-multi` or new subcommand): create one lexer from the input (file or stdin or -e string), then loop ParseOne until done; print each AST. Example: `tryparse -multi g:json 2.json` or `tryparse -multi -e g:json '{} {}'` (if grammar allows space between objects) to get two parses. This gives concrete ./apps examples for both single- and multi-object.

### 2.4 Optional: grammar pragma

- Stream-plan suggests an optional flag or pragma (e.g. `!streaming` or `!multi_record`) so "existing single-document use stays accept only on EOF". With the above, we already keep Parse() as single-document (accept only on EOF); the optional part is whether we **emit** accept_and_yield at all (e.g. only when a pragma is set). For "same .bnf without source mods," the minimal approach is: always emit accept_and_yield in the table; single-object = use Parse(); multi-object = use ParseOne loop. No pragma needed unless we want to avoid adding accept_and_yield for grammars that never need multi-record (smaller tables). Defer pragma to a follow-up unless you want it in scope.

---

## Summary

- **Step 1**: Lexer API uses `io.Reader`; string-backed helper (e.g. `FromString` or `strings.NewReader`) for tests and -e; CLIs use file-backed reader for files/stdin and string-backed for -e; trylex file mode = one lexer per whole file.
- **Step 2**: Table generator adds accept_and_yield in accept state for non-EOF terminals; codegen adds AcceptAndYield, stashed lookahead, and ParseOne; Parse unchanged (single-object); add tests and -multi (or similar) example in apps for multi-object.

After Step 1, stop and chat before implementing Step 2.
