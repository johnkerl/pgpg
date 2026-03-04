_Can you make some suggestions of what a "group by and apply" UX might look like?_

---

Sure! This is one of the more awkward spots in Miller's current UX, so there's real room to think about it.

**Current state of things**

Right now, to do a custom aggregation grouped by a field, you write something like:

```
mlr put '
  @sum[$color] += $quantity;
  end { emit @sum, "color" }
' data.csv
```

That works, but it requires knowing about out-of-stream variables (`@`), the map-indexing pattern, and `emit` — a lot of machinery for a common task.

**Some possible UX directions**

**Option A: A `group-apply` verb with a DSL block**

```
mlr group-apply -g color '{ $quantity = sum($quantity); $rate = mean($rate) }' data.csv
```

The idea: `-g` specifies the grouping key(s), and the block uses familiar aggregate function names. The verb handles the accumulation and emission implicitly. This is the most "Miller-flavored" approach — it fits the existing verb + DSL pattern.

**Option B: Aggregation functions inside `put` that are group-aware**

Rather than a new verb, extend `put` so that certain functions are recognized as accumulators and automatically handled across the group:

```
mlr put -g color '$quantity = sum($quantity); $rate = mean($rate)' data.csv
```

The `-g` flag tells `put` to buffer and reduce by group rather than transform record-by-record. This is probably the most ergonomic for existing users since it looks like `put` they already know.

**Option C: A SQL-style `summarize` verb**

```
mlr summarize -g color -f quantity:sum,rate:mean data.csv
```

This is closer to how `stats1` already works, but with a cleaner syntax for specifying field+aggregator pairs. The downside is it's less flexible for custom expressions.

**Option D: Pipe-style with a reduce keyword**

```
mlr group-by color | reduce '{ $quantity = sum($quantity) }' data.csv
```

Borrowing from the `then`-chain philosophy but making group-by and reduce explicit separate steps. Feels natural if you think of it as "first group, then collapse."

---

My intuition is that **Option B** is the most consistent with Miller's existing design — it leverages the familiar `put` verb and just adds a flag that changes its execution mode from record-at-a-time to group-at-a-time. The mental model shift is minimal: "put normally transforms each record; put -g collapses groups." It also naturally extends to multi-key grouping (`-g color,shape`) without any new syntax.

The main design question would be: what does the output record look like? Does it retain all fields (with grouped fields constant and aggregated fields replaced), or just the group keys plus the explicitly computed fields? The latter is probably cleaner.

---

_In your report you mentioned:_

> Miller has two modes: verbs (like `sort`, `cut`, `stats1`) and DSL expressions (using `put` and `filter`). This dual interface is powerful but creates a subtle coherence problem: some things are only possible through the DSL, some only through verbs, and the boundary is not always logical.
> 
> For example, there is no built-in "group-by and apply function" verb for custom aggregations — users must construct that using DSL `put` with out-of-stream variables (`@oosvar`) and `emit` at the `end` block. This pattern requires understanding a fairly advanced part of the DSL (out-of-stream variables) for something that feels like a common operation.
> 
> The `emit`/`tee` interaction with the then-chain is another area where the mental model requires careful construction. DSL output statements (`emit`, `tee`, `dump`) interact with the record stream in ways that are documented but non-obvious — emitted records can go into the stream, or to a file, or to stdout, depending on syntax, and the output format follows the main command-line format flags.

_Can you mention a few more examples for DSL improvement in addition to group-by-and-apply, and the emit syntax?_

---

Sure. Here are a few more areas where the DSL/verb seam creates friction:

**Deduplication with custom logic**

`mlr uniq -f color` is easy. But if you want "keep the first record per group where some condition holds," you're back to out-of-stream variables:

```
mlr put 'if (!@seen[$color] && $quantity > 50) { @seen[$color] = 1 } else { filter false }' data.csv
```

That `filter false` inside a `put` block is itself a surprise — users often don't realize you can mix filtering logic into `put`. A `dedup -g color -w '$quantity > 50'` verb would be more discoverable.

**Running/cumulative aggregates**

`mlr step -a delta,ewma -f quantity` exists, but it covers only a small set of pre-baked functions. If you want a running median, or a rolling window of anything not in `step`'s fixed menu, you fall off the verb cliff into DSL territory with manual window management using arrays in out-of-stream variables. There's no general `window` or `rolling` verb.

**Conditional field renaming or reshaping**

`mlr rename` is fine for static renames. But "rename fields matching a pattern" or "rename fields based on their values or position" requires `put` with a `for (k, v in $*)` loop — a fairly advanced idiom that involves iterating over the record as a map. Something like `mlr rename -r 's/^prefix_//'` (regex rename) was a long-requested feature that took years to land.

**Multi-record lookahead/lookbehind**

Anything that requires looking at the *next* record while processing the current one is effectively impossible without buffering everything manually. `mlr step` handles a few specific cases (differences, moving averages), but "emit a record only if the next record meets condition X" requires accumulating records in an array and flushing in `end` — which also breaks streaming. There's no clean verb-level abstraction for this.

**Joining against a computed or in-memory table**

`mlr join` works against a file on disk. If you want to join against a table you've *derived* mid-pipeline — e.g., first compute per-color averages, then join each record against those averages — you have to either write to a temp file and re-invoke Miller, or do the whole thing in one DSL pass using out-of-stream variables as a lookup table. The two-pass pattern is genuinely awkward:

```
mlr put '@avg[$color] += $quantity; @count[$color] += 1;
  end { for (c, s in @avg) { @avg[c] = s / @count[c] } }
  then put '$norm = $quantity / @avg[$color]' data.csv
```

A `join -against-computed` or the ability to reference a prior verb's aggregation output in a later verb would close this gap.

---

The common thread across most of these is the same one you identified for group-by-and-apply: **the verb layer is thin, and the DSL layer requires knowing about out-of-stream variables, `for` loops over `$*`, and `emit` semantics** — a fairly steep jump from "I know the basic verbs" to "I can do this moderately complex thing." The gap between those two levels of fluency is wide, and there's a lot of useful functionality sitting just on the far side of it.
