package vdf

import (
	"container/heap"
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
	length := 1000000
	phi := big.NewInt(896655229516)
	fmt.Println("t=", length)

	primes := computeL(length)
	start := time.Now()
	prod := product1(primes, phi)
	fmt.Println("length of product: ", prod.BitLen())
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("time 1: ", elapsed)

	primes = computeL(length)
	P := big.NewInt(1)
	t1 = time.Now()
	for _, v := range primes {
		P.Mul(P, v)
		P.Mod(P, phi)
	}
	t2 := time.Now()
	elapsed2 := t2.Sub(t1)
	fmt.Println("time 2: ", elapsed2)
	fmt.Printf("ratio: %.2f\n", float64(elapsed2)/float64(elapsed))
	// fmt.Println(P)
	if P.Cmp(prod) != 0 {
		t.Error("boo")
	}
}

func TestProductTime(t *testing.T) {
	length := 100000000
	keysize := 512

	p, q := generateTwoGoodPrimes(keysize)
	N := new(big.Int).Mul(p, q)

	fmt.Println("p and q", p, q)
	fmt.Println("N ", N)
	tmp := big.NewInt(0)
	tmp.Add(p, big.NewInt(-1))
	phi := big.NewInt(0)
	phi.Add(q, big.NewInt(-1))
	phi.Mul(phi, tmp)
	fmt.Println("phi", phi)

	fmt.Println("t=", length)
	fmt.Println("keysize=", keysize)

	primes := computeL(length)

	start := time.Now()
	prod := product1(primes, phi)
	fmt.Println("length of product: ", prod.BitLen())
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("time 1: ", elapsed)

	primes = computeL(length)
	P := big.NewInt(1)
	t1 = time.Now()
	for _, v := range primes {
		P.Mul(P, v)
		P.Mod(P, phi)
	}
	t2 := time.Now()
	elapsed2 := t2.Sub(t1)
	fmt.Println("time 2: ", elapsed2)
	fmt.Printf("ratio: %.2f\n", float64(elapsed2)/float64(elapsed))
	// fmt.Println(P)
	if P.Cmp(prod) != 0 {
		t.Error("boo")
	}
}
func TestProductMod(t *testing.T) {
	length := 100000
	phi := big.NewInt(896655229516)
	fmt.Println("t=", length)

	primes := computeL(length)
	P := big.NewInt(1)
	t1 := time.Now()
	for _, v := range primes {
		P.Mul(P, v)
		P.Mod(P, phi)
	}
	t2 := time.Now()
	elapsed1 := t2.Sub(t1)
	fmt.Println("time 1: ", elapsed1)
	// fmt.Println(P)

	P_inv := big.NewInt(1)
	P_inv.ModInverse(P, phi)

	P = big.NewInt(1)
	t1 = time.Now()
	for _, v := range primes {
		P.Mul(P, v)
	}
	t2 = time.Now()
	elapsed2 := t2.Sub(t1)
	fmt.Println("time 2: ", elapsed2)
	P_inv_2 := big.NewInt(1)
	P_inv_2.ModInverse(P, phi)
	// fmt.Println(P)

	if P_inv.Cmp(P_inv_2) != 0 {
		t.Error("boo")
	}
}

func TestPriorityQueue(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, big.NewInt(3))
	heap.Push(pq, big.NewInt(1))
	heap.Push(pq, big.NewInt(5))
	heap.Push(pq, big.NewInt(2))
	heap.Push(pq, big.NewInt(8))
	fmt.Printf("minimum: %d\n", (*pq)[0])
	fmt.Printf("%v ", *pq)
	for pq.Len() > 0 {
		fmt.Printf("%d ", heap.Pop(pq))
		fmt.Printf("%v ", *pq)
	}
}
