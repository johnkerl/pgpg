package main

import (
	"fmt"
	"strconv"

	"github.com/johnkerl/goffl/pkg/f2poly"
	"github.com/johnkerl/goffl/pkg/f2polymod"
)

// F2PolyModNumeric implements Numeric[*f2polymod.F2PolyMod, int] for F2[x]/m(x).
// The modulus polynomial m(x) is fixed when the backend is created (e.g. from a flag).
// Literals are raw hex digits only (e.g. "1fe"), no 0x/0b prefix. Matches goffl / f2poly mode.
type F2PolyModNumeric struct {
	Modulus *f2poly.F2Poly
}

// NewF2PolyModNumeric creates a backend for F2[x] mod m(x). modulus must be non-zero.
func NewF2PolyModNumeric(modulus *f2poly.F2Poly) (*F2PolyModNumeric, error) {
	if modulus == nil || modulus.Bits == 0 {
		return nil, fmt.Errorf("modulus polynomial must be non-zero")
	}
	return &F2PolyModNumeric{Modulus: modulus}, nil
}

func (b *F2PolyModNumeric) FromString(s string) (*f2polymod.F2PolyMod, error) {
	bits, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return nil, err
	}
	return f2polymod.New(&f2poly.F2Poly{Bits: bits}, b.Modulus), nil
}

// ParseExponent parses the exponent as decimal (e.g. 2**10 uses exponent 10, not 0x10).
func (b *F2PolyModNumeric) ParseExponent(s string) (int, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	if v < 0 {
		return 0, fmt.Errorf("negative exponent disallowed")
	}
	if v > 0x7fffffff {
		return 0, fmt.Errorf("exponent too large")
	}
	return int(v), nil
}

func (b *F2PolyModNumeric) String(t *f2polymod.F2PolyMod) string {
	return t.Residue.String()
}

func (b *F2PolyModNumeric) Add(a, c *f2polymod.F2PolyMod) *f2polymod.F2PolyMod      { return a.Add(c) }
func (b *F2PolyModNumeric) Subtract(a, c *f2polymod.F2PolyMod) *f2polymod.F2PolyMod { return a.Sub(c) }
func (b *F2PolyModNumeric) Multiply(a, c *f2polymod.F2PolyMod) *f2polymod.F2PolyMod { return a.Mul(c) }

func (b *F2PolyModNumeric) Divide(a, c *f2polymod.F2PolyMod) (*f2polymod.F2PolyMod, error) {
	return a.Div(c)
}

func (b *F2PolyModNumeric) Mod(a, c *f2polymod.F2PolyMod) (*f2polymod.F2PolyMod, error) {
	return nil, fmt.Errorf("modulo not defined for F2[x]/m(x) (use f2poly mode for %%)")
}

func (b *F2PolyModNumeric) Exponentiate(base *f2polymod.F2PolyMod, exp int) (*f2polymod.F2PolyMod, error) {
	return base.Pow(exp)
}

func (b *F2PolyModNumeric) ToExponent(v *f2polymod.F2PolyMod) (int, error) {
	if v.Residue.Bits > 0x7fffffff {
		return 0, fmt.Errorf("exponent too large (use small nonnegative integer)")
	}
	return int(v.Residue.Bits), nil
}

func (b *F2PolyModNumeric) Negate(v *f2polymod.F2PolyMod) *f2polymod.F2PolyMod {
	return v.Neg()
}
