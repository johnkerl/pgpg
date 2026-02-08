-----

_I'm building a parser generator and intending to make AST construction a crucial part of it. What
are some examples of use cases in which someone might want a parser generator but might not want any
AST construction?_

-----

Great question! There are several scenarios where you'd want parsing without AST construction:

## 1. **Validation/Recognition Only**
Sometimes you just need to know if input is valid according to a grammar. Examples:
- Configuration file validators that only report "valid" or "invalid"
- Protocol message validators in network services
- Input sanitizers that accept/reject but don't process

## 2. **Streaming/Event-Based Processing**
When data is too large to hold in memory or you want immediate action:
- Log file analyzers that extract metrics on-the-fly
- Data format converters that stream output while parsing input
- Real-time monitoring systems that trigger alerts during parsing

## 3. **Direct Execution/Interpretation**
Some systems execute during parsing rather than building an intermediate structure:
- Simple calculator languages that compute results immediately
- Command-line interpreters that execute commands as they're parsed
- Template engines that output rendered text directly

## 4. **Syntax Highlighting/Tokenization**
Editor tooling often just needs token identification:
- Code editors categorizing text for coloring
- Lexical analyzers feeding other tools
- Search/grep tools with syntax awareness

## 5. **Performance-Critical Applications**
When AST overhead is prohibitive:
- High-throughput data ingestion pipelines
- Embedded systems with memory constraints
- JIT compilers that translate directly to bytecode

## 6. **Incremental/Partial Parsing**
Some applications parse repeatedly on small changes:
- IDE autocomplete systems
- Live syntax checkers
- Diff tools that only need structural awareness

For your parser generator, you might consider making AST construction optional or pluggable,
allowing users to register callbacks/actions at parse points instead. This gives flexibility for
both worlds.

-----

_Thanks! I'm trying to make something that's language-independent, where the lexer-generator and
parser-generator can be done at one point in time, and have all the productions tables etc. saved
off as, say, JSON. Then another tool in another language could pick those up and use them to do
parsing._

_What I'm still trying to get my head around is how callback registration---which seems inherently
language-specific---would fit in with this kind of data-driven, multi-language vision._

-----

This is a really interesting design challenge! You're hitting on a fundamental tension: the parsing
tables are pure data and language-agnostic, but *actions* are inherently code. Here are some
approaches people have taken:

## 1. **Deferred Actions (Your AST Approach)**
Build a generic, language-agnostic AST structure in your JSON output. Each runtime implementation then traverses this AST with language-specific visitors/transformers. This is what you're already leaning toward and it cleanly separates concerns.

## 2. **Action Codes/Opcodes**
Instead of callbacks, emit symbolic action identifiers:
```json
{
  "production": "expr -> expr '+' term",
  "action": "binary_add",
  "args": [0, 2]  // which children to use
}
```
Each language runtime provides a mapping from action codes to actual functions. The JSON specifies *what* to do abstractly; each runtime decides *how*.

## 3. **Semantic Predicates as Data**
For validation/filtering use cases, you could express simple conditions declaratively:
```json
{
  "production": "number -> DIGIT+",
  "constraint": {"type": "range", "min": 0, "max": 255}
}
```
Limited but useful for common cases.

## 4. **Two-Phase Approach**
- Phase 1: Generator outputs tables + a generic AST definition
- Phase 2: Language-specific tool generates code from the AST schema

This is what Protocol Buffers and similar tools do. Your generator creates the schema; each language has its own code generator.

## 5. **Embedded DSL for Actions**
Define a simple, language-agnostic expression language in your JSON:
```json
{
  "action": {"type": "construct", "node": "BinaryOp", "fields": {"op": "$2", "left": "$1", "right": "$3"}}
}
```
Each runtime interprets this mini-language. Keeps actions declarative while allowing some computation.

## 6. **No Actions in Tables**
The most minimal approach: parsing tables only describe structure. All semantic actions live entirely in the language-specific runtime, which queries the parse tree/events. The JSON is purely structural metadata.

**My take:** For a language-agnostic system, option #1 (generic AST) combined with #2 (optional action codes for common operations) gives you the best flexibility. You get a working default (the AST) plus optimization opportunities (symbolic actions) without forcing every runtime to implement complex interpretation logic.

Does your use case lean more toward compiler-building (where ASTs are essential) or toward data processing (where streaming/validation might dominate)?

-----

_More toward compiler-building, although I do want to support data processing (perhaps I want too much)._

-----

You definitely don't want too much—these are complementary! Here's how they can coexist:

## **A Layered Architecture**

**Core (in JSON):**
- Parsing tables (LR/LALR/whatever)
- Grammar metadata (production rules, symbol info)
- Generic AST schema (node types, fields, relationships)

**Runtime implementations provide:**
1. **AST mode** (default): Builds the tree structure according to your schema
2. **Streaming mode** (opt-in): User provides handlers/visitors that fire during parsing

## **How Streaming Fits In**

Even for data processing, you still need the *structure* defined in your JSON. The difference is execution model:

```python
# Compiler mode - build full AST
ast = parser.parse(source)
compiler.compile(ast)

# Streaming mode - same parser, different interface
parser.parse_streaming(source, {
    'on_number': lambda val: process_number(val),
    'on_expr': lambda node: evaluate(node)
})
```

The parsing tables are identical. The runtime just offers two consumption patterns.

## **Practical Middle Ground**

Your JSON defines:
- **Structure**: "An `expr` node has `left`, `op`, `right` fields"
- **Not behavior**: What to *do* with that structure

Each language runtime can then offer:
- `parse_to_ast()` - builds the tree
- `parse_with_callbacks(handlers)` - streaming
- `validate_only()` - just check syntax

The compiler-building use case gets full ASTs. The data-processing use case gets events. Same tables, different wrappers.

**This is exactly what tools like ANTLR do**, by the way—they generate parse trees by default but also support listener/visitor patterns for streaming. You're on the right track!
