package vdf

import (
	"fmt"
	"math/big"
	"math/bits"
)

func log2(x int) int {
	var r int = 0
	for ; x > 1; x >>= 1 {
		r++
	}
	return r
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func HumanSize(b int) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float32(b)/float32(div), "KMGTPE"[exp])
}

func bigToFixedLengthBytes(in *big.Int, length int) []byte {
	if length*8 < in.BitLen() {
		panic("specified length is shorter than input")
	}
	_S := bits.UintSize / 8
	var buf []byte
	realSize := len(in.Bits()) * _S
	if length >= realSize {
		buf = make([]byte, length)
	} else {
		buf = make([]byte, realSize)
	}
	i := len(buf)
	for _, d := range in.Bits() {
		for j := 0; j < _S; j++ {
			i--
			buf[i] = byte(d)
			d >>= 8
		}
	}
	return buf[len(buf)-length:]
}

func isStrongPrime(prime *big.Int) bool {
	half := new(big.Int)
	half.Sub(prime, big.NewInt(1))
	half.Div(half, big.NewInt(2))
	return half.ProbablyPrime(20)
}
