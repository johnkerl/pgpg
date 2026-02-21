package main

// Numeric is the arithmetic numeric interface. T is the value type, E is the
// exponent type for Exponentiate (e.g. int for modular, same as T for int/float).
// ParseExponent parses an exponent from a literal string (e.g. for ** right-hand
// side). For f2poly/f2polymod the exponent is decimal; for others it matches FromString.
type Numeric[T, E any] interface {
	FromString(s string) (T, error)
	ParseExponent(s string) (E, error)
	String(t T) string
	Add(a, b T) T
	Subtract(a, b T) T
	Multiply(a, b T) T
	Divide(a, b T) (T, error)
	Mod(a, b T) (T, error)
	Exponentiate(base T, exp E) (T, error)
	ToExponent(v T) (E, error)
	Negate(v T) T
}
