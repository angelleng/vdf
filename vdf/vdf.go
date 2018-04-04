package vdf

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
	"prime"
)

const t = 10000  // 1000000
const B = 100000 // 1 << 10
const lambda = 10

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
		primes = prime.Primes(116485863)
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
	gs = make([]*big.Int, len(hs))
	for i, v := range hs {
		gs[i] = big.NewInt(0)
		gs[i].Exp(v, P_inv, N)
	}
	return
}

func isStrongPrime(prime *big.Int) bool {
	small := big.NewInt(0)
	small.Add(prime, big.NewInt(-1))
	small.Div(small, big.NewInt(2))
	return small.ProbablyPrime(60)
}

func generateChallenge(X interface{}) (Ind_L, Ind_S []int) {
	// first turn X into bytes
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(X)
	data := w.Bytes()

	h := sha256.New()
	Ind_L = make([]int, lambda)
	for i := 0; i < lambda; i++ {
		h.Write(data)
		shasum := h.Sum(nil)
		ind := big.NewInt(0)
		ind.SetBytes(shasum)
		ind.Mod(ind, big.NewInt(t))
		Ind_L[i] = int(ind.Int64())
	}

	Ind_S = make([]int, lambda)
	for i := 0; i < lambda; i++ {
		h.Write(data)
		shasum := h.Sum(nil)
		ind := big.NewInt(0)
		ind.SetBytes(shasum)
		ind.Mod(ind, big.NewInt(B))
		Ind_S[i] = int(ind.Int64())
	}
	return
}

// interface
func Setup() (*EvalKey, *VerifyKey) {
	fmt.Println("setup")
	fmt.Printf("parameters: t: %v, B: %v, lambda: %v \n", t, B, lambda)

	L := computeL(t)
	fmt.Printf("L [%v %v %v %v ... %v %v %v] \n", L[0], L[1], L[2], L[3], L[len(L)-3], L[len(L)-2], L[len(L)-1])
	P := big.NewInt(1)
	for _, v := range L {
		P.Mul(P, v)
	}

	p := big.NewInt(1)
	q := big.NewInt(1)
	var rsaKey *rsa.PrivateKey
	var N *big.Int

	for !isStrongPrime(p) || !isStrongPrime(q) {
		rsaKey, _ = rsa.GenerateKey(cryptorand.Reader, 59)
		N = rsaKey.PublicKey.N
		p = rsaKey.Primes[0]
		q = rsaKey.Primes[1]
	}

	fmt.Println("p and q", rsaKey.Primes)
	fmt.Println("N and e", rsaKey.PublicKey)
	fmt.Println("D", rsaKey.D)
	fmt.Println("precomputed", rsaKey.Precomputed)
	tmp := big.NewInt(0)
	tmp.Add(rsaKey.Primes[0], big.NewInt(-1))
	phi := big.NewInt(0)
	phi.Add(rsaKey.Primes[1], big.NewInt(-1))
	phi.Mul(phi, tmp)
	fmt.Println("phi", phi)

	hashfunc := func(input *big.Int) (hashval *big.Int) {
		h := sha256.New()
		h.Write(input.Bytes())
		shasum := h.Sum(nil)
		hashval = big.NewInt(0)
		hashval.SetBytes(shasum)
		hashval.Mod(hashval, N)
		return
	}
	P_inv := big.NewInt(1)
	P_inv.ModInverse(P, phi)

	hs := computehs(hashfunc, B)
	gs := computegs(hs, P_inv, N)

	evaluateKey := EvalKey{N, hashfunc, gs}
	verifyKey := VerifyKey{N, hashfunc}
	return &evaluateKey, &verifyKey
}

func Evaluate(evaluateKey *EvalKey, x int) (y *big.Int) {
	N := evaluateKey.G
	gs := evaluateKey.Gs
	L := computeL(t)
	fmt.Println("evaluate")

	L_ind, S_x := generateChallenge(x)
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
	y.Exp(g_x, exp_coeff, N)

	fmt.Println("y", y)
	return
}

func Verify(verifyKey *VerifyKey, y *big.Int, x int) bool {
	fmt.Println("verify")
	hashfunc := verifyKey.H
	hs := computehs(hashfunc, B)
	L := computeL(t)

	N := verifyKey.G
	L_ind, S_x := generateChallenge(x)
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

	h2.Exp(y, P_x, N)

	fmt.Println("h", h_x)
	fmt.Println("h2", h2)

	h_x.Add(h_x, h2.Neg(h2))

	return h_x.Sign() == 0
}

//example usage
//func main() {
//	evaluateKey, verifyKey := setup()
//	solution := evaluate(evaluateKey, 1)
//	success := verify(verifyKey, solution, 1)
//	fmt.Println("result: ", success)
//	fmt.Println("finish")
//}
