package main

import (
	"fmt"
	"strconv"
)

// ModInt is a value in Z/nZ (modular arithmetic). N is the modulus.
type ModInt struct {
	V int // representative in [0, N)
	N int
}

// ModNumeric implements Numeric[ModInt, int]; exponent is always int.
type ModNumeric struct {
	N int
}

func NewModNumeric(n int) (*ModNumeric, error) {
	if n <= 0 {
		return nil, fmt.Errorf("modulus must be positive, got %d", n)
	}
	return &ModNumeric{N: n}, nil
}

func (b *ModNumeric) FromString(s string) (ModInt, error) {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return ModInt{}, err
	}
	return b.normalize(int(v)), nil
}

func (b *ModNumeric) ParseExponent(s string) (int, error) {
	v, err := strconv.ParseInt(s, 0, 64)
	return int(v), err
}

func (b *ModNumeric) normalize(v int) ModInt {
	r := v % b.N
	if r < 0 {
		r += b.N
	}
	return ModInt{V: r, N: b.N}
}

func (b *ModNumeric) String(t ModInt) string {
	return strconv.Itoa(t.V)
}

func (b *ModNumeric) Add(a, bVal ModInt) ModInt {
	return b.normalize(a.V + bVal.V)
}

func (b *ModNumeric) Subtract(a, bVal ModInt) ModInt {
	return b.normalize(a.V - bVal.V)
}

func (b *ModNumeric) Multiply(a, bVal ModInt) ModInt {
	return b.normalize(a.V * bVal.V)
}

func (b *ModNumeric) Divide(a, bVal ModInt) (ModInt, error) {
	inv, err := modInverse(bVal.V, b.N)
	if err != nil {
		return ModInt{}, fmt.Errorf("no modular inverse for %d mod %d", bVal.V, b.N)
	}
	return b.normalize(a.V * inv), nil
}

func (b *ModNumeric) Mod(a, bVal ModInt) (ModInt, error) {
	if bVal.V == 0 {
		return ModInt{}, fmt.Errorf("modulo by zero")
	}
	return b.normalize(a.V % bVal.V), nil
}

func (be *ModNumeric) Exponentiate(base ModInt, exp int) (ModInt, error) {
	if exp < 0 {
		inv, err := modInverse(base.V, be.N)
		if err != nil {
			return ModInt{}, err
		}
		base.V = inv
		exp = -exp
	}
	out := 1
	baseV := base.V
	for exp > 0 {
		if exp&1 == 1 {
			out = (out * baseV) % be.N
		}
		baseV = (baseV * baseV) % be.N
		exp >>= 1
	}
	if out < 0 {
		out += be.N
	}
	return ModInt{V: out, N: be.N}, nil
}

func (b *ModNumeric) ToExponent(v ModInt) (int, error) {
	return v.V, nil
}

func (b *ModNumeric) Negate(v ModInt) ModInt {
	return b.normalize(-v.V)
}

// modInverse returns x such that (a * x) % n == 1. Requires gcd(a, n) == 1.
func modInverse(a, n int) (int, error) {
	a = a % n
	if a < 0 {
		a += n
	}
	g, x, _ := extendedGCD(a, n)
	if g != 1 {
		return 0, fmt.Errorf("no inverse: gcd(%d, %d) = %d", a, n, g)
	}
	x = x % n
	if x < 0 {
		x += n
	}
	return x, nil
}

func extendedGCD(a, b int) (gcd, x, y int) {
	if b == 0 {
		return a, 1, 0
	}
	gcd, x1, y1 := extendedGCD(b, a%b)
	x = y1
	y = x1 - (a/b)*y1
	return gcd, x, y
}
