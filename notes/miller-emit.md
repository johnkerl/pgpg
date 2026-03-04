I find Miller's emit, emitp, emit1, and emitf statements confusing. Can you explain them to me?

---

This is a really elegant design insight. Since Miller now has first-class maps, you're right that the transformation logic can be decoupled from the emission step.

Here's what the split could look like:

---

## Emission: just `emit1`

Since the goal of `emitx` is "emit a map as-is, one record," that's exactly what `emit1` already does. No new emitter needed — `emit1` *is* `emitx`.

---

## The `something` transformation functions

These would be pure map-in, map-out functions — composable, testable, usable outside of emit contexts too:

### `flatten(map)` / `flatten(map, separator)`
Replaces `emitp`'s nested-map-flattening behavior:
```
emit1(flatten(@data))
```
Takes `{"red": {"large": 10}}` → `{"red:large": 10}`

### `hoist(map, key)` or `label(map, key)`
Replaces the `emit @var, "name"` behavior of promoting a scalar/map with a named key:
```
emit1(label(@sum, "total"))
```
Takes `42` → `{"total": 42}`

### `unlevel(map)` or `entries(map)`
For the "one record per top-level key" behavior of `emit`:
```
for (record in entries(@counts)) { emit1(record) }
```
Or possibly returning an array of maps that you iterate.

---

## Side-by-side comparison

| Old API | New API |
|---|---|
| `emit @sum, "sum"` | `emit1(label(@sum, "sum"))` |
| `emitp @data, "color"` | `emit1(flatten(@data))` |
| `emit1 @rec` | `emit1(@rec)` *(unchanged)* |
| `emitf @a, @b` | `emit1({"a": @a, "b": @b})` *(just a map literal)* |

---

## The key benefit

Because `flatten`, `label`, etc. are **ordinary functions**, you can also use them outside of emit — in conditionals, assignments, passing to other functions — which wasn't possible when the transformation was baked into the emit statement itself. The API becomes **orthogonal** rather than a collection of special cases.
