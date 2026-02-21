package main

import (
	"fmt"
	"strconv"
)

// IntNumeric implements Numeric[int, int] for integer arithmetic.
type IntNumeric struct{}

func (IntNumeric) FromString(s string) (int, error) {
	v, err := strconv.ParseInt(s, 0, 64)
	return int(v), err
}

func (IntNumeric) ParseExponent(s string) (int, error) {
	v, err := strconv.ParseInt(s, 0, 64)
	return int(v), err
}

func (IntNumeric) String(t int) string {
	return strconv.Itoa(t)
}

func (IntNumeric) Add(a, b int) int      { return a + b }
func (IntNumeric) Subtract(a, b int) int { return a - b }
func (IntNumeric) Multiply(a, b int) int { return a * b }

func (IntNumeric) Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

func (IntNumeric) Mod(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("modulo by zero")
	}
	return a % b, nil
}

func (IntNumeric) Exponentiate(base, exp int) (int, error) {
	if exp < 0 {
		return 0, fmt.Errorf("negative exponent for integer power")
	}
	out := 1
	for i := 0; i < exp; i++ {
		out *= base
	}
	return out, nil
}

func (IntNumeric) ToExponent(v int) (int, error) { return v, nil }
func (IntNumeric) Negate(v int) int              { return -v }
