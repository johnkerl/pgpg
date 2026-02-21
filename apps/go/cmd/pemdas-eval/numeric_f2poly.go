package main

import (
	"fmt"
	"strconv"

	"github.com/johnkerl/goffl/pkg/f2poly"
)

// F2PolyNumeric implements Numeric[*f2poly.F2Poly, int] for polynomials over GF(2).
// Literals are raw hex digits only (e.g. "1fe", "0"), no 0x/0b prefix. Matches goffl.
// No fixed modulus; operations are in F2[x].
type F2PolyNumeric struct{}

func (F2PolyNumeric) FromString(s string) (*f2poly.F2Poly, error) {
	v, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return nil, err
	}
	return f2poly.New(v), nil
}

// ParseExponent parses the exponent as decimal (e.g. 2**10 uses exponent 10, not 0x10).
func (F2PolyNumeric) ParseExponent(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		return 0, fmt.Errorf("negative exponent disallowed for F2Poly")
	}
	if v > 0x7fffffff {
		return 0, fmt.Errorf("exponent too large for F2Poly")
	}
	return int(v), nil
}

func (F2PolyNumeric) String(t *f2poly.F2Poly) string {
	return t.String()
}

func (F2PolyNumeric) Add(a, b *f2poly.F2Poly) *f2poly.F2Poly      { return a.Add(b) }
func (F2PolyNumeric) Subtract(a, b *f2poly.F2Poly) *f2poly.F2Poly { return a.Sub(b) }
func (F2PolyNumeric) Multiply(a, b *f2poly.F2Poly) *f2poly.F2Poly { return a.Mul(b) }

func (F2PolyNumeric) Divide(a, b *f2poly.F2Poly) (*f2poly.F2Poly, error) {
	if b.Bits == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return a.Quo(b), nil
}

func (F2PolyNumeric) Mod(a, b *f2poly.F2Poly) (*f2poly.F2Poly, error) {
	if b.Bits == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}
	return a.Mod(b), nil
}

func (F2PolyNumeric) Exponentiate(base *f2poly.F2Poly, exp int) (*f2poly.F2Poly, error) {
	return base.Pow(exp)
}

func (F2PolyNumeric) ToExponent(v *f2poly.F2Poly) (int, error) {
	if v.Bits > 0x7fffffff {
		return 0, fmt.Errorf("exponent too large for F2Poly (use small nonnegative integer)")
	}
	return int(v.Bits), nil
}

func (F2PolyNumeric) Negate(v *f2poly.F2Poly) *f2poly.F2Poly {
	return v.Neg()
}
