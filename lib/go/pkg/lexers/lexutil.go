package lexers

// RunePredicateFunc is used by lexers (e.g. EBNFLexer) for predicates like unicode.IsSpace.
// Exported so that app-side lexers in apps/go/manual can use the same type.
type RunePredicateFunc func(rune) bool
