package vdf

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"math/big"
	"prime"
	"runtime"
	"sync"
	"time"

	"github.com/onrik/gomerkle"
)

type EvalKey struct {
	G  *big.Int
	H  func(*big.Int) *big.Int
	Gs []*big.Int
}

type VerifyKey struct {
	G *big.Int
	H func(*big.Int) *big.Int
}

// helper functions
func computeL(t int) (L []*big.Int) {
	primes := prime.Primes(16485863)
	if len(primes) < t {
		primes = prime.Primes(2038074751)
	}
	if len(primes) < t {
		primes = prime.Primes(22563323963)
	}
	if len(primes) < t {
		fmt.Println("error: not enough primes generated.")
	}
	L = make([]*big.Int, t)
	for i := 0; i < t; i++ {
		L[i] = big.NewInt(int64(primes[i+1]))
	}
	return
}

func computehs(hashfunc func(*big.Int) *big.Int, B int) (hs []*big.Int) {
	hs = make([]*big.Int, B)
	for i := 0; i < B; i++ {
		hs[i] = hashfunc(big.NewInt(int64(i)))
	}
	return
}

func computegs(hs []*big.Int, P_inv *big.Int, N *big.Int) (gs []*big.Int) {
	fmt.Println("start compute gs")
	start := time.Now()
	gs = make([]*big.Int, len(hs))

	var wg sync.WaitGroup
	wg.Add(len(hs))

	input := make(chan int, 10)

	go func() {
		for i := 0; i < len(hs); i++ {
			input <- i
		}
		close(input)
	}()

	for worker := 0; worker < runtime.NumCPU(); worker++ {
		go func() {
			for {
				i, ok := <-input
				if ok {
					v := hs[i]
					gs[i] = big.NewInt(0)
					gs[i].Exp(v, P_inv, N)
					wg.Done()
				} else {
					return
				}
			}
		}()
	}

	wg.Wait()

	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println("compute gs time", elapsed)
	return
}

func isStrongPrime(prime *big.Int, L []*big.Int, P *big.Int) bool {
	resP := new(big.Int)
	resP.Sub(prime, big.NewInt(1))
	resP.Div(resP, big.NewInt(2))
	resP.Mod(resP, P)

	return resP.ProbablyPrime(20)

	for _, v := range L {
		resv := new(big.Int)
		resv.Mod(resP, v)
		if resv.Sign() == 0 {
			return false
		}
	}
	return true
}

func generateChallenge(t, B, lambda int, X interface{}) (Ind_L, Ind_S []int) {
	// first turn X into bytes
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(X)
	data := w.Bytes()

	h := sha256.New()
	Ind_L = make([]int, 0, lambda)

	for i := 0; i < lambda; {
		dupe := false
		h.Write(data)
		shasum := h.Sum(nil)
		ind := big.NewInt(0)
		ind.SetBytes(shasum)
		ind.Mod(ind, big.NewInt(int64(t)))
		index := int(ind.Int64())
		for _, v := range Ind_L {
			if v == index {
				dupe = true
				break
			}
		}
		if !dupe {
			Ind_L = append(Ind_L, index)
			i++
		}
	}

	Ind_S = make([]int, 0, lambda)
	for i := 0; i < lambda; {
		dupe := false
		h.Write(data)
		shasum := h.Sum(nil)
		ind := big.NewInt(0)
		ind.SetBytes(shasum)
		ind.Mod(ind, big.NewInt(int64(B)))
		index := int(ind.Int64())
		for _, v := range Ind_S {
			if v == index {
				dupe = true
				break
			}
		}
		if !dupe {
			Ind_S = append(Ind_S, index)
			i++
		}
	}
	return
}

func generateTwoGoodPrimes(keysize int, L []*big.Int, P *big.Int) (p, q *big.Int) {
	primechan := make(chan *big.Int)
	found := false
	fmt.Println("no of cores: ", runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for !found {
				// fmt.Println("trying1")
				candidate, _ := cryptorand.Prime(cryptorand.Reader, keysize)
				if isStrongPrime(candidate, L, P) {
					primechan <- candidate
				}
			}
		}()
	}
	p = <-primechan
	fmt.Println("have one prime")
	q = <-primechan
	found = true
	fmt.Println("have second prime")
	return
}

// interface
func Setup(t, B, lambda, keysize int) (*EvalKey, *VerifyKey) {
	if lambda >= t || lambda >= B {
		err := errors.New("error, lambda should be less than t and B")
		fmt.Println(err)
	}
	fmt.Println("\nSETUP")
	fmt.Printf("parameters: -t=%v -B=%v -lambda=%v -keysize=%v \n", t, B, lambda, keysize)

	L := computeL(t)
	fmt.Printf("L [%v %v %v %v ... %v %v %v] \n", L[0], L[1], L[2], L[3], L[len(L)-3], L[len(L)-2], L[len(L)-1])

	P := product1(L)

	p := big.NewInt(1)
	q := big.NewInt(1)
	N := new(big.Int)
	p, q = generateTwoGoodPrimes(keysize, L, P)

	N.Mul(p, q)

	fmt.Println("p and q", p, q)
	fmt.Println("N ", N)
	tmp := big.NewInt(0)
	tmp.Add(p, big.NewInt(-1))
	phi := big.NewInt(0)
	phi.Add(q, big.NewInt(-1))
	phi.Mul(phi, tmp)
	fmt.Println("phi", phi)

	hashfunc := func(input *big.Int) (hashval *big.Int) {
		h := sha256.New()
		var shasum []byte
		for len(shasum)*8 < N.BitLen() {
			h.Write(input.Bytes())
			shasum = h.Sum(shasum)
		}
		hashval = big.NewInt(0)
		hashval.SetBytes(shasum)
		hashval.Mod(hashval, N)
		return
	}

	P_inv := big.NewInt(1)
	t1 := time.Now()
	P_inv.ModInverse(P, phi)
	t2 := time.Now()
	elapsed := t2.Sub(t1)
	fmt.Println("gen 1/P time: ", elapsed)

	hs := computehs(hashfunc, B)
	gs := computegs(hs, P_inv, N)

	evaluateKey := EvalKey{N, hashfunc, gs}
	verifyKey := VerifyKey{N, hashfunc}
	return &evaluateKey, &verifyKey
}

type Evaluator struct {
	T      int
	B      int
	Lambda int
	L      []*big.Int
	P      *big.Int
	N      *big.Int
	Gs     []*big.Int
}

type Proof struct {
	y           *big.Int       // solution
	L_x         []*big.Int     // primes
	MerkleProof gomerkle.Proof // proof of merkle tree
}

func (ev *Evaluator) Init(t, B, lambda int, evaluateKey *EvalKey) {
	start := time.Now()
	ev.T = t
	ev.B = B
	ev.Lambda = lambda
	ev.N = evaluateKey.G
	ev.Gs = evaluateKey.Gs
	ev.L = computeL(t)
	ev.P = product1(ev.L)
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("evaluator init time", elapsed)
	tree := gomerkle.NewTree(sha256.New())
	fmt.Println("tree", tree)

}

func (ev *Evaluator) Eval(x int) (y *big.Int) {
	t1 := time.Now()
	L_ind, S_x := generateChallenge(ev.T, ev.B, ev.Lambda, x)
	L_x := make([]*big.Int, ev.Lambda)
	for i, v := range L_ind {
		L_x[i] = ev.L[v]
	}

	P_x := big.NewInt(1)
	for _, v := range L_x {
		P_x.Mul(P_x, v)
	}

	g_x := big.NewInt(1)
	for _, v := range S_x {
		g_x.Mul(g_x, ev.Gs[v])
		g_x.Mod(g_x, ev.N)
	}

	exp_coeff := big.NewInt(1)
	exp_coeff.Div(ev.P, P_x)

	fmt.Println("g_x", g_x)
	t2 := time.Now()
	elapsed1 := t2.Sub(t1)

	y = big.NewInt(1)

	start := time.Now()
	y.Exp(g_x, exp_coeff, ev.N)
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("y", y)
	fmt.Println("evaluate prepare time", elapsed1)
	fmt.Println("actual evaluate time", elapsed)

	return
}

func Evaluate(t, B, lambda int, evaluateKey *EvalKey, x int) (y *big.Int) {
	N := evaluateKey.G
	gs := evaluateKey.Gs
	L := computeL(t)
	fmt.Println("\nEVALUATE")

	L_ind, S_x := generateChallenge(t, B, lambda, x)
	L_x := make([]*big.Int, lambda)
	for i, v := range L_ind {
		L_x[i] = L[v]
	}

	P_x := big.NewInt(1)
	for _, v := range L_x {
		P_x.Mul(P_x, v)
	}

	g_x := big.NewInt(1)
	for _, v := range S_x {
		g_x.Mul(g_x, gs[v])
		g_x.Mod(g_x, N)
	}

	P := big.NewInt(1)
	for _, v := range L {
		P.Mul(P, v)
	}

	exp_coeff := big.NewInt(1)
	exp_coeff.Div(P, P_x)

	fmt.Println("g_x", g_x)

	y = big.NewInt(1)

	start := time.Now()
	y.Exp(g_x, exp_coeff, N)
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("y", y)
	fmt.Println("actual evaluate time", elapsed)
	return
}

type Verifier struct {
	T      int
	B      int
	Lambda int
	N      *big.Int
	L      []*big.Int
	Hash   func(*big.Int) *big.Int
}

func (vr *Verifier) Init(t, B, lambda int, verifyKey *VerifyKey) {
	start := time.Now()
	vr.T = t
	vr.B = B
	vr.Lambda = lambda
	vr.N = verifyKey.G
	vr.L = computeL(t)
	vr.Hash = verifyKey.H

	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("verifier init time", elapsed)
}

func (vr *Verifier) Verify(x int, y *big.Int) bool {
	t1 := time.Now()

	L_ind, S_x := generateChallenge(vr.T, vr.B, vr.Lambda, x)
	L_x := make([]*big.Int, vr.Lambda)
	for i, v := range L_ind {
		L_x[i] = vr.L[v]
	}

	P_x := big.NewInt(1)
	for _, v := range L_x {
		P_x.Mul(P_x, v)
	}

	h_x := big.NewInt(1)
	for _, v := range S_x {
		h := vr.Hash(big.NewInt(int64(v)))
		h_x.Mul(h_x, h)
		h_x.Mod(h_x, vr.N)
	}
	h2 := big.NewInt(1)

	t2 := time.Now()
	elapsed1 := t2.Sub(t1)

	start := time.Now()
	h2.Exp(y, P_x, vr.N)
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("h", h_x)
	fmt.Println("h2", h2)
	fmt.Println("verify prepare time", elapsed1)
	fmt.Println("actual verify time", elapsed)
	compare := h_x.Cmp(h2)
	return compare == 0
}

func Verify(t, B, lambda int, verifyKey *VerifyKey, y *big.Int, x int) bool {
	fmt.Println("\nVERIFY")
	hashfunc := verifyKey.H
	hs := computehs(hashfunc, B)
	L := computeL(t)

	N := verifyKey.G
	L_ind, S_x := generateChallenge(t, B, lambda, x)
	L_x := make([]*big.Int, lambda)
	for i, v := range L_ind {
		L_x[i] = L[v]
	}

	P_x := big.NewInt(1)
	for _, v := range L_x {
		P_x.Mul(P_x, v)
	}

	h_x := big.NewInt(1)
	for _, v := range S_x {
		h_x.Mul(h_x, hs[v])
		h_x.Mod(h_x, N)
	}
	h2 := big.NewInt(1)

	start := time.Now()
	h2.Exp(y, P_x, N)
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("h", h_x)
	fmt.Println("h2", h2)
	fmt.Println("actual verify time", elapsed)
	compare := h_x.Cmp(h2)
	return compare == 0
}
