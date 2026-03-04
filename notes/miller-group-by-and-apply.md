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
