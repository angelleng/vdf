package vdf

import (
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestGenerateChallenge(t *testing.T) {
	ind_L, ind_S := generateChallenge(1000, 10000, 100, 2)
	t.Log("test\n")
	t.Log(ind_L)
	fmt.Println(ind_S)
}

func TestProduct(t *testing.T) {
	length := 800000
	fmt.Println("t=", length)

	primes := computeL(length)
	start := time.Now()
	product := Product(primes)
	// fmt.Println(product)
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("time 1: ", elapsed)

	primes = computeL(length)
	P := big.NewInt(1)
	t1 = time.Now()
	for _, v := range primes {
		P.Mul(P, v)
	}
	t2 := time.Now()
	elapsed = t2.Sub(t1)
	fmt.Println("time 2: ", elapsed)
	// fmt.Println(P)
	if P.Cmp(product) != 0 {
		t.Error("noo")
	}
}
