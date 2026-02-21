package main

import (
	"fmt"
	"math"
	"strconv"
)

// FloatNumeric implements Numeric[float64, float64] for float arithmetic.
type FloatNumeric struct{}

func (FloatNumeric) FromString(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func (FloatNumeric) ParseExponent(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func (FloatNumeric) String(t float64) string {
	return fmt.Sprintf("%g", t)
}

func (FloatNumeric) Add(a, b float64) float64      { return a + b }
func (FloatNumeric) Subtract(a, b float64) float64 { return a - b }
func (FloatNumeric) Multiply(a, b float64) float64 { return a * b }

func (FloatNumeric) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}

func (FloatNumeric) Mod(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("modulo by zero")
	}
	return math.Mod(a, b), nil
}

func (FloatNumeric) Exponentiate(base, exp float64) (float64, error) {
	return math.Pow(base, exp), nil
}

func (FloatNumeric) ToExponent(v float64) (float64, error) { return v, nil }
func (FloatNumeric) Negate(v float64) float64              { return -v }
