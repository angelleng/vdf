package vdf

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func StrongPrime(bits int) *big.Int {
	p := new(big.Int)
	for {
		s, _ := rand.Prime(rand.Reader, bits/2-3)
		t, _ := rand.Prime(rand.Reader, bits/2-10)

		fmt.Println(s)
		// fmt.Println(s.BitLen())
		fmt.Println(t)
		// fmt.Println(t.BitLen())
		r := big.NewInt(1)
		a := new(big.Int)
		a.Lsh(t, 1)
		// fmt.Println(a)
		for !r.ProbablyPrime(20) {
			r.Add(r, a)
		}
		fmt.Println(r)
		// fmt.Println(r.BitLen())

		p0 := new(big.Int)
		p0.Sub(r, big.NewInt(2))
		p0.Exp(s, p0, r)
		// p0.Lsh(p0, 1)
		p0.Mul(p0, big.NewInt(2))
		p0.Mul(p0, s)
		p0.Add(p0, big.NewInt(-1))
		// fmt.Println(p0)
		// fmt.Println(p0.BitLen())

		b := new(big.Int)
		p.Set(p0)
		b.Mul(r, s)
		b.Lsh(b, 1)
		for !p.ProbablyPrime(20) && p.BitLen() <= bits {
			p.Add(p, b)
		}
		// fmt.Println(p)
		// fmt.Println(p.BitLen())

		if p.ProbablyPrime(20) && p.BitLen() == bits {

			tmp := big.NewInt(1)
			tmp.Sub(p, tmp)
			tmp.Mod(tmp, r)
			fmt.Println(tmp)

			tmp2 := big.NewInt(1)
			tmp2.Add(p, tmp2)
			tmp2.Mod(tmp2, s)
			fmt.Println(tmp2)
			return p
		}

	}
	return p
}
