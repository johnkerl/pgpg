# Constructive Feedback on Miller (mlr): A User Perspective

**Compiled from GitHub Issues, Discussions, Hacker News threads, Stack Exchange, and community commentary**

---

## Executive Summary

Miller (mlr) is widely praised as a powerful, multi-format data-processing tool, but its user community has surfaced a consistent set of friction points over the years. This report organizes those into themes, drawing on GitHub issue reports, community discussions, Hacker News comment threads, and direct user feedback. The goal is not to diminish the tool — which has a remarkably positive reception overall — but to give an honest catalog of where users struggle, where the tool underperforms relative to alternatives, and where design decisions carry real costs.

---

## 1. Steep Learning Curve and Discoverability

One of the most frequently cited barriers to Miller adoption is how long it takes new users to become productive with it. Multiple independent commenters have noted that despite Miller being a powerful tool, "it can be hard to understand how it operates." A blogger who built a GUI wrapper around Miller explained that the motivation was to help people "get started," noting that many users "often resort to just using spreadsheets, simpler command line tools like grep, and manual counting" because the ramp-up feels steep.

The core issue is that Miller conflates several things a new user must learn simultaneously:

- **Format flags** (`--icsv`, `--opprint`, `--c2j`, etc.) — a large and at-times confusing matrix of input/output specifiers
- **Verb semantics** — a distinct vocabulary (`put`, `filter`, `stats1`, `tee`, `emit`, etc.) that differs from anything in the standard Unix toolkit
- **The DSL** — a small programming language with its own sigils (`$fieldname`, `@oosvar`), control structures, and type system

The 2021 Miller User Survey reportedly confirmed this: one of the main themes was that "some things should be easier to find." That feedback directly drove significant documentation restructuring in Miller 6, where the focus shifted from expert-to-expert writing toward more introductory material.

Even after learning the basics, users consistently report needing to refer back to documentation for anything non-trivial. One Smashing Magazine tutorial acknowledges that Miller is harder to use than simpler tools like `cut` or `grep` "due to its verb-based syntax and field-aware operations," while simultaneously praising its power.

The tool's documentation is genuinely comprehensive, but its breadth can itself be overwhelming. Users must understand not just the verbs, but when to use `put` vs. `filter`, the difference between `stats1` and `stats2`, the distinction between CSV and CSV-lite, the `emit`/`emitf`/`emitp`/`emit1` family, and so on. For newcomers, the discoverability problem is real: there is no obvious entry point that maps from "I want to do X" to the right verb or flag combination.

---

## 2. Shell Quoting and Escaping Complexity

A persistent and well-documented pain point is the friction between Miller's DSL syntax and the shell's quoting rules. Because Miller's DSL uses the `$` sigil for field names, and because shells interpret `$` specially, users routinely end up in situations where constructing or iterating over DSL expressions from shell variables requires careful — and non-obvious — quoting gymnastics.

An early GitHub issue (#112) proposed replacing `$` with `%` for precisely this reason: "for j in 1 2 3 4; do mlr --from myfile.dat put '$output = '$j' * $i' ..." — notice the need to break out of and re-enter single quotes. The workaround of using `$` for Miller fields and `%` for shell variables would have been cleaner, but was ultimately abandoned because `%` clashes with the modulo operator (and whitespace sensitivity around `%` caused syntax errors).

The problem compounds on **Windows**, where the situation degrades significantly. The Windows terminal parses command-line arguments differently from Unix shells, and there are DSL expressions that are simply impossible to pass through the Windows command prompt without resorting to a script file. One maintainer noted in a 2023 GitHub discussion: "I had a very hard time getting all DSL expressions past the Windows terminal parser. There are some things that just can't be gotten through." The recommended workaround — writing DSL expressions to a file and using `mlr put -f` — adds ceremony that undermines Miller's value as a quick one-liner tool. Using triple double-quotes as a Windows quoting escape (`"""`) works in some cases but is visually confusing.

Even on Unix/macOS, simple expressions like putting a single quote inside a string require ugly shell escaping patterns (`'$a="It'\''s OK"'`). The documentation acknowledges this as "a little tricky" and provides the correct approach, but the experience of encountering it for the first time is friction.

---

## 3. CSV Parsing Edge Cases and Failure Modes

Miller's handling of real-world, imperfect CSV files is a recurring source of user frustration. There are several distinct sub-issues:

**RFC-4180 strictness.** Early versions of Miller did not support RFC-4180 double-quoted fields at all, which drew sharp criticism. One early issue (#4) put it directly: "You really cannot say you have a tool that is designed to support CSV, without supporting CSV." This was eventually fixed, but the tool went through a period where it had both a "csvlite" mode (no quoting support) and a proper "csv" mode, creating confusion about which to use.

**Backslash-escaped quotes.** A 2016 issue (#270) documented that backslash-style quote escaping (`\"`) — a common non-standard convention used by many real-world CSV generators — is not handled correctly. Miller strictly follows RFC-4180 (where quotes are escaped by doubling: `""`), meaning files from programs that use backslash escaping will produce parse errors. While this is technically correct behavior per the RFC, it conflicts with what users actually encounter.

**The "mismatch" error message.** One of the most common support questions in Miller's GitHub Discussions is users encountering: `mlr: CSV header/data length mismatch N != M at filename foo.csv row R`. This error, while accurate, is cryptic about what the actual problem is. In one case, a government meteorological CSV file from Australia that Excel and R both parsed fine was rejected by Miller with this error — the issue turned out to be non-standard double-quote usage in the CSV, but the error message gave no guidance about that. Commenters in the discussion thread noted that "the error message does not help, and perhaps it should be changed."

**Error resilience.** A more fundamental request (issue #523) is for Miller to optionally skip or handle malformed lines rather than halting processing. Miller's philosophy leans toward strictness ("format-compliant tooling incentivizes format-compliant data"), which the maintainer acknowledged was "well-intended" but not always helpful for users working with machine-generated, imperfect data. A `--skip-bad-lines` or similar flag has been a recurring feature request.

---

## 4. DSL Syntax and Parse Error Messages

The Miller DSL is expressive but has a few syntax choices that surprise users.

**The `^` bitwise XOR surprise.** A recent GitHub discussion showed a user receiving this error:

```
mlr: cannot parse DSL expression. Parse error on token "^" at line 1 column 20.
```

The problem was that `^` is the bitwise XOR operator in Miller, not the string-start anchor as in regex context. Users coming from awk or sed naturally try to use `^` for regex anchoring in filter expressions but get a confusing parse error.

**Parse error verbosity.** When a DSL expression fails to parse, Miller emits a very long list of "expected one of:" tokens — a raw dump of the parser's lookahead state rather than a human-readable explanation of what went wrong. For example:

```
Expected one of: { ( field_name $[ braced_field_name $[[ $[[[ full_srec oosvar_name @[ braced_oosvar_name full_oosvar all non_sigil_name float int...
```

This is technically complete but practically unhelpful for anyone who isn't already deeply familiar with Miller's grammar. Miller 6 did improve parse error messages to include line and column numbers, which was a meaningful step forward, but the "expected one of" dump remains noisy.

**The `emit` family complexity.** The distinction between `emit`, `emitp`, `emitf`, and `emit1` is confusing to users and has caused compatibility issues between Miller 5 and Miller 6. A regression was filed (issue #827) showing that `emit @in["a"]` — valid in Miller 5.3.0 — produces a parse error in Miller 6.0.0-rc. The `emit` API has been described by users as one of the harder areas to understand, with the documentation requiring careful reading to understand the difference between emitting individual variables and maps.

**UTF-8 / Latin-1 encoding inconsistency.** Issue #1358 reported an interesting edge case: field names containing UTF-8 characters that happen to also be encodeable as Latin-1 (like accented characters in the é/á/ñ range) could not be used inside `${...}` braces in the DSL in Miller 6, even though purely UTF-8 characters (like Chinese characters) worked fine. The maintainer acknowledged the inconsistency and the difficulty of resolving it due to the Go regexp library's requirements.

---

## 5. Performance Gap Relative to Specialized Alternatives

For simple, single-operation tasks on very large files, Miller is measurably slower than single-purpose tools. A 2024 GitHub issue (#1527) reported that for a 5 GB CSV file, `mlr cut` for a single column took about 5 minutes, while `xsv select` (a Rust-based CSV tool) completed the same task in 30 seconds — a roughly 10× gap.

This is not entirely surprising: Miller's streaming architecture involves parsing each row into a full key-value record, applying format-specific logic, and re-serializing, whereas tools like xsv or qsv are built in Rust with tight CSV-specific I/O paths and column-skipping optimizations. Miller pays a per-row overhead that xsv does not.

The performance gap is most pronounced in:

- **Simple column selection/filtering on large CSV files**, where Rust-based tools (xsv, qsv) are faster
- **JSON-heavy workloads**, where Miller's own benchmarks note that "JSON continues to be a CPU-intensive format"
- **Analytics-style aggregations on large datasets**, where DuckDB (which has become a popular Miller alternative) vastly outperforms due to its columnar storage and vectorized execution

The Go port (Miller 6) was acknowledged to perform roughly on par with the C-based Miller 5 for simple operations, with improvements for complex then-chains. However, the Go runtime's garbage collector introduces latency characteristics that a C or Rust implementation would not. The Miller 6 documentation itself notes "a significant slowdown" observed on Linux under certain conditions (low battery) for the Go version that did not affect the C version — indicating GC-related sensitivity.

In the Hacker News thread from March 2023, the top comment recommends ClickHouse Local as an alternative that "outperforms every other tool." DuckDB was repeatedly mentioned in the same thread by users preferring SQL-based access, with one comment noting that `select a,b,c from '*.jsonl.gz'` in DuckDB "has been a huge improvement to my workflows." These alternatives represent a real competitive challenge in the performance-sensitive segment of Miller's use case space.

---

## 6. Windows Support Is a Second-Class Experience

Windows support for Miller has been a persistent problem. The history includes:

- The early Windows releases (v5.x era) required shipping a Cygwin DLL alongside the binary
- After the Go port in Miller 6, Windows received a native binary, but Cygwin compatibility broke — Miller 6 does not understand Cygwin-style paths (e.g., `/cygdrive/c/path`) and requires native Windows paths, which complicates use in Cygwin shell scripts
- One user noted that with Miller 5.10.0 (the last C version) everything worked fine, but after upgrading to Miller 6 built with Go, Cygwin path handling was lost entirely
- On Windows Command Prompt, many DSL expressions involving single quotes, `^` characters, or certain special characters cannot be passed at all without resorting to workaround quoting schemes (triple double-quotes) that are documented but awkward

The maintainer's response to Cygwin issues has generally been to recommend switching to MSYS2 rather than fixing Cygwin support — a pragmatic choice given the complexity, but one that strands users who have existing shell scripts built around Cygwin.

---

## 7. Discoverability and Naming

Miller suffers from a name and command that are both common English words, making it genuinely hard to search for help online. Searching for "miller" returns results about paint, Arthur Miller, and the R machine-learning framework (also called mlr/mlr3). Searching for "mlr" is somewhat better but still pulls up the R package.

This is not a new problem — it was acknowledged by the maintainer as early as 2015 ("I've considered the searchability issue when deciding on a name for it, but eventually favored the day-to-day minimum-typing short name over better searchability"). The consequence is that users who encounter an error or want to look up a recipe have a harder time than users of tools with unique names (like `jq` or `awk`). This feeds back into the discoverability problem: Miller users can't easily Google their way to answers.

---

## 8. JSON Handling Limitations

While Miller supports JSON input and output, its JSON model is structurally limited by its tabular record abstraction. Miller works best with "tabular JSON" — arrays of objects at the top level. Deeply nested JSON objects, arrays-of-arrays, or JSON documents that don't fit the records-of-key-value-pairs model require explicit flattening/unflattening, and complex transformations often push users back toward `jq`.

One GitHub discussion (discussion #633) has a user who tried to nest fields into a sub-object using Miller and ultimately gave up: "I think I was just trying to use Miller in a way that it isn't suited and I found a different CSV tool that is specifically written to work with JSON." The maintainer's documentation acknowledges this explicitly, describing Miller as handling "tabular JSON" and noting that for deeply nested structures, tools like `jq` may be a better fit.

Miller's own documentation notes: "JSON is a format which supports scalars... as well as 'objects' (maps) and 'arrays' (lists), while Miller is a tool for handling tabular data only." This is a design choice that is also a limitation — users discovering it mid-task often feel misled by Miller's billing as a JSON tool.

---

## 9. Record Heterogeneity and Sparse Data

One of Miller's distinguishing features is its support for record heterogeneity — records with different field sets can coexist in a stream. But this feature, while theoretically powerful, creates practical problems:

- CSV format does not handle schema changes at all; Miller will error on a schema change in CSV mode (`mlr: exiting due to data error`). Users must use CSV-lite for truly heterogeneous CSV data.
- After operations like `join` with unpaired records or `stats1` across heterogeneous fields, the output can be "sparse" — records with missing fields — and users must explicitly use `unsparsify` to fill in gaps, a step that is easy to forget and produces confusing output when omitted.
- The distinction between CSV/TSV (rigid schema) and formats like DKVP or NIDX (flexible) is not obvious until users run into schema-change errors.

---

## 10. Memory Usage for Non-Streaming Operations

Miller markets itself as a streaming tool, and it mostly is — but several operations break this model. `sort`, `tac` (reverse cat), `stats1` with percentile aggregation, and `join` (in non-sorted mode, which is the default since Miller 6) all buffer data in memory. For very large files, this can cause significant RAM usage or even OOM conditions.

Earlier versions of Miller had a bug where memory-mapped I/O caused pages not to be released, causing crashes on files over 4 GB (this was fixed by using stdio for files above that threshold). The current documentation notes that Miller "retains only as much data as needed" for non-streaming operations, but does not always make clear to users which operations are streaming and which are not.

A man-page-style reference notes: "For very large datasets, some memory-intensive operations (e.g., sorting, joining with large keys) might consume significant RAM, as mlr often loads records into memory for processing." Users expecting fully streaming behavior throughout can be surprised.

---

## 11. The DSL vs. Verb Mismatch

Miller has two modes: **verbs** (like `sort`, `cut`, `stats1`) and **DSL expressions** (using `put` and `filter`). This dual interface is powerful but creates a subtle coherence problem: some things are only possible through the DSL, some only through verbs, and the boundary is not always logical.

For example, there is no built-in "group-by and apply function" verb for custom aggregations — users must construct that using DSL `put` with out-of-stream variables (`@oosvar`) and `emit` at the `end` block. This pattern requires understanding a fairly advanced part of the DSL (out-of-stream variables) for something that feels like a common operation.

The `emit`/`tee` interaction with the then-chain is another area where the mental model requires careful construction. DSL output statements (`emit`, `tee`, `dump`) interact with the record stream in ways that are documented but non-obvious — emitted records can go into the stream, or to a file, or to stdout, depending on syntax, and the output format follows the main command-line format flags.

---

## 12. The Go Port: Tradeoffs and Minor Regressions

The Miller 5→6 transition (C to Go) was by most accounts a success, but it introduced some friction:

- Several DSL expressions that were valid in Miller 5 produce parse errors in Miller 6 (e.g., `emit @in["a"]`, issue #827)
- Some scripts that relied on Miller 5's implicit type coercion behavior (`put -S` and `put -F` to force string/float mode) need updating, though Miller 6 was designed to handle type conversion more automatically
- The Go runtime adds a binary startup cost and GC overhead compared to the C implementation; benchmarks note that for CPU-constrained environments or battery-limited hardware, Go's GC can cause intermittent slowdowns
- Performance on JSON I/O in Miller 6 is comparable to Miller 5 on Mac but was faster on Linux in Miller 5

The `list.List` data structure used internally for some time was a Go-standard-library linked list that turned out to be a performance bottleneck; it was eventually replaced with Go slices (release 6.17.0, March 2026), delivering a notable speedup. That this optimization arrived in version 6.17.0 (years after the Go port) suggests the performance room was there all along and waiting to be found.

---

## 13. Competition from Newer Tools

Miller's competitive position has shifted meaningfully since its 2015 debut. Several tools have emerged or matured that address overlapping use cases, often with advantages in specific dimensions:

- **DuckDB** (`duckdb`): Handles CSV, JSON, and Parquet with SQL syntax; vastly faster for aggregation-heavy workloads; increasingly used as a one-liner tool. Many Hacker News commenters recommend it over Miller for analytical tasks.
- **xsv / qsv**: Rust-based CSV tools that are significantly faster for simple CSV operations (column selection, filtering, statistics). The 10× performance gap vs. `mlr cut` on large files (issue #1527) illustrates the cost of Miller's generality.
- **jq / gojq**: For JSON-only workloads, these tools are better suited to nested JSON than Miller, with more expressive query languages for tree-shaped data.
- **Nushell**: A shell that treats structured data as first-class; its pipelines operate on typed records in a way that makes Miller's then-chaining feel like a subset.

Miller's niche — multi-format streaming with a DSL, free from SQL, in a single binary — is still genuinely distinct. But the competitive landscape means users increasingly have options that are faster or more ergonomic for specific use cases.

---

## 14. Snap Package Naming Inconsistency

A minor but consistently mentioned friction point: the Snap package for Miller installs the command as `miller` rather than `mlr`. This is noted in the release pages ("Note: for Snap only, the executable is `miller`, not `mlr`") but is easy to miss and breaks scripts written for the standard `mlr` command name.

---

## Summary Table

| Issue Area | Severity | Status |
|---|---|---|
| Learning curve / discoverability | High | Partially addressed in Miller 6 docs |
| Shell quoting friction | Medium–High | Design limitation; mitigated by `-f` flag |
| CSV parsing edge cases | Medium | Ongoing; `--lazy-quotes` helps some |
| DSL parse error quality | Medium | Improved in Miller 6 (line/col numbers) |
| Performance vs. xsv/DuckDB | Medium–High | Design tradeoff; improvements ongoing |
| Windows / Cygwin support | Medium | Known limitation; MSYS2 recommended |
| Name searchability | Low–Medium | Design choice; acknowledged by maintainer |
| JSON nesting limitations | Medium | By design; documented |
| Non-streaming memory use | Low–Medium | Documented; not always surfaced clearly |
| DSL / verb coherence | Medium | Ongoing area of design refinement |
| Miller 5→6 DSL regressions | Low | Mostly addressed |
| Competitive landscape | Strategic | Not a bug, but a positioning challenge |
| Snap naming inconsistency | Low | Documented |

---

*Report compiled from: GitHub issues and discussions at github.com/johnkerl/miller; Hacker News threads (March 2023 and December 2021); community blog posts (Smashing Magazine, Hackaday, personal blogs); Miller 6 documentation (miller.readthedocs.io); and direct user quotes preserved throughout.*
