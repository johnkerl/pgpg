# Test grammars

## AME

This is add and multiply of integers, with equal precedence.

```
# Lex
whitespace: ignored
number: "[0-9]+"

# Parse
Term
    : Number
    | Number "+" Term
    | Number "*" Term
;
Number: number
```

## AMNE

This is add and multiply of integers, with unequal precedence.

```
# Lex
whitespace: ignored
number: "[0-9]+"

# Parse
Term
    : MulTerm
    | Number "+" MulTerm
;
MulTerm
    : Number
    | Number "*" MulTerm
;
Number: number
```

## VPEMDAS

TO DO: type up the grammar.

Examples:
```
x = 1 + 2 * 3 # Assign
x # Evaluate
y = x + 2 * y
```

## SENG

Simple English statements:

* subject/object/verb or imperative object/verb
* articles, adjectives, adverbs
* explicit terminal wordlists -- no morphological type-inference e.g. ends with `-ly` meaning it must be an adverb

TO DO: type up the grammar

Examples:

```
Dog eats food.
The dog eats the food.
The brown dog tastes the very new food.
Take the old book.
# _Maybe_ conjunctions at higher and lower levels: _
Take the old book and read it.
Take the green and grey book.
```

## CSV/DKVP

Comma-separated and DKVP files (Miller).

## Miller DSL

This is an ultimate goal.
