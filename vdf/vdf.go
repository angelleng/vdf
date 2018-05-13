package vdf

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"math/big"
	"os"
	"prime"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/onrik/gomerkle"
)

// helper functions
func computeL(t int) (L []*big.Int) {
	var primes []uint64
	if t <= 1000000 {
		primes = prime.Primes(16485863)
	} else if t <= 100000000 {
		primes = prime.Primes(2038074751)
	} else if t <= 1000000000 {
		primes = prime.Primes(22801763513)
	} else if t <= 10000000000 {
		primes = prime.Primes(252097800629)
	}
	if len(primes) < t+1 {
		fmt.Println("error: not enough primes generated.")
	}
	L = make([]*big.Int, t)
	for i := 0; i < t; i++ {
		L[i] = big.NewInt(int64(primes[i+1]))
	}
	return
}

func computegs(hash func(*big.Int) *big.Int, B int, P_inv *big.Int, N *big.Int) (gs []*big.Int) {
	fmt.Println("start compute gs")
	start := time.Now()
	gs = make([]*big.Int, B)

	var wg sync.WaitGroup
	wg.Add(B)

	input := make(chan int, 10)

	go func() {
		for i := 0; i < B; i++ {
			input <- i
		}
		close(input)
	}()

	for worker := 0; worker < runtime.NumCPU(); worker++ {
		go func() {
			for {
				i, ok := <-input
				if ok {
					v := hash(big.NewInt(int64(i)))
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

func computeAndStoreGs(hash func(input *big.Int) *big.Int, B int, P_inv *big.Int, N *big.Int, gspath string) {
	perFile := 2 ^ 20
	nFiles := B / perFile
	lastFile := B % perFile
	fmt.Println(nFiles)
	fmt.Println(lastFile)

	for i := 0; i <= nFiles; i++ {
		fmt.Println(i)
		filename := gspath + strconv.Itoa(i)
		var thisFile int
		if i < nFiles {
			thisFile = perFile
		} else if lastFile != 0 {
			thisFile = lastFile
		} else {
			return
		}

		gs := make([]*big.Int, thisFile)

		var wg sync.WaitGroup
		wg.Add(thisFile)
		input := make(chan int, 100)

		go func() {
			for j := 0; j < thisFile; j++ {
				ind := perFile*i + j
				input <- ind
			}
			close(input)
		}()

		for worker := 0; worker < runtime.NumCPU(); worker++ {
			go func() {
				for {
					ind, ok := <-input
					if ok {
						v := hash(big.NewInt(int64(ind)))
						gs[ind%perFile] = big.NewInt(0)
						gs[ind%perFile].Exp(v, P_inv, N)
						wg.Done()
					} else {
						return
					}
				}
			}()
		}

		wg.Wait()
		file, _ := os.Create(filename)
		encoder := gob.NewEncoder(file)
		encoder.Encode(gs)
		file.Close()

	}
}

func isStrongPrime(prime *big.Int) bool {
	half := new(big.Int)
	half.Sub(prime, big.NewInt(1))
	half.Div(half, big.NewInt(2))
	return half.ProbablyPrime(20)
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

func generateTwoGoodPrimes(keysize int) (p, q *big.Int) {
	primechan := make(chan *big.Int)
	found := false
	fmt.Println("no of cores: ", runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for !found {
				// fmt.Println("trying1")
				candidate, _ := cryptorand.Prime(cryptorand.Reader, keysize)
				if isStrongPrime(candidate) {
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
type EvalKey struct {
	G  *big.Int
	H  func(*big.Int) *big.Int
	Gs []*big.Int
}

type VerifyKey struct {
	G *big.Int
	H func(*big.Int) *big.Int
}

func Hashfunc(input *big.Int, N *big.Int) (hashval *big.Int) {
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

func Setup(t, B, lambda, keysize int) (*EvalKey, *VerifyKey) {
	if lambda >= t || lambda >= B {
		err := errors.New("error, lambda should be less than t and B")
		fmt.Println(err)
	}
	fmt.Println("\nSETUP")
	fmt.Printf("parameters: -t=%v -B=%v -lambda=%v -keysize=%v \n", t, B, lambda, keysize)

	L := computeL(t)
	fmt.Printf("L [%v %v %v %v ... %v %v %v] \n", L[0], L[1], L[2], L[3], L[len(L)-3], L[len(L)-2], L[len(L)-1])

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

	hashfunc := func(input *big.Int) (hashval *big.Int) {
		return Hashfunc(input, N)
	}

	P := product1(L, phi)
	P_inv := big.NewInt(1)
	t1 := time.Now()
	P_inv.ModInverse(P, phi)
	t2 := time.Now()
	elapsed := t2.Sub(t1)
	fmt.Println("gen 1/P time: ", elapsed)

	gs := computegs(hashfunc, B, P_inv, N)
	computeAndStoreGs(hashfunc, B, P_inv, N, "gs")

	evaluateKey := EvalKey{N, hashfunc, gs}
	verifyKey := VerifyKey{N, hashfunc}
	return &evaluateKey, &verifyKey
}

type Evaluator struct {
	T      int
	B      int
	Lambda int
	L      []*big.Int
	N      *big.Int
	Gs     []*big.Int
	Ltree  gomerkle.Tree
}

type Solution struct {
	Y           *big.Int         // solution
	L_x         []*big.Int       // primes
	MerkleProof []gomerkle.Proof // proof of merkle tree
}

func (ev *Evaluator) Init(t, B, lambda int, evaluateKey *EvalKey) {
	start := time.Now()
	ev.T = t
	ev.B = B
	ev.Lambda = lambda
	ev.N = evaluateKey.G
	ev.Gs = evaluateKey.Gs
	ev.L = computeL(t)
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("evaluator init time", elapsed)

	ev.Ltree = gomerkle.NewTree(sha256.New())
	for _, v := range ev.L {
		ev.Ltree.AddData(v.Bytes())
	}
	err := ev.Ltree.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Println("tree", ev.Ltree.Height(), ev.Ltree.Root())
}

func readFileAndComputeGx(S_x []int, gspath string, B int, N *big.Int) *big.Int {
	sort.Ints(S_x)
	perFile := 2 ^ 20
	nFiles := B / perFile

	gx := make([]*big.Int, len(S_x))
	for i := 0; i <= nFiles; i++ {
		var gs []*big.Int
		filename := gspath + strconv.Itoa(i)
		file, _ := os.Open(filename)
		decoder := gob.NewDecoder(file)
		decoder.Decode(&gs)
		file.Close()
		for j, v := range S_x {
			if v >= i*perFile && v < (i+1)*perFile {
				offset := v % perFile
				gx[j] = gs[offset]
			}
		}
	}

	y := big.NewInt(1)
	for _, v := range gx {
		y.Mul(y, v)
		y.Mod(y, N)
	}
	return y
}

func (ev *Evaluator) Eval(x interface{}) (sol Solution) {
	t1 := time.Now()
	L_ind, S_x := generateChallenge(ev.T, ev.B, ev.Lambda, x)

	y := readFileAndComputeGx(S_x, "gs", ev.B, ev.N)
	// y := big.NewInt(1)
	// for _, v := range S_x {
	// 	y.Mul(y, ev.Gs[v])
	// 	y.Mod(y, ev.N)
	// }

	L_x := make([]*big.Int, ev.Lambda)
	for i, v := range L_ind {
		L_x[i] = ev.L[v]
	}

	fmt.Println("g_x", y)
	t2 := time.Now()
	elapsed1 := t2.Sub(t1)

	start := time.Now()
	for i, v := range ev.L {
		yes := true
		for _, l := range L_ind {
			if i == l {
				yes = false
				break
			}
		}
		if !yes {
			continue
		}
		y.Exp(y, v, ev.N)
	}
	end := time.Now()
	elapsed := end.Sub(start)

	fmt.Println("compute merkle proof for L_x")
	proofs := make([]gomerkle.Proof, ev.Lambda)
	for i, v := range L_ind {
		proof := ev.Ltree.GetProof(v)
		proofs[i] = proof
	}

	fmt.Println("y", y)
	fmt.Println("evaluate prepare time", elapsed1)
	fmt.Println("actual evaluate time", elapsed)
	sol.Y = y
	sol.L_x = L_x
	sol.MerkleProof = proofs
	return
}

type Verifier struct {
	T      int
	B      int
	Lambda int
	N      *big.Int
	L      []*big.Int
	Hash   func(*big.Int) *big.Int
	Lroot  []byte
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

	Ltree := gomerkle.NewTree(sha256.New())
	for _, v := range vr.L {
		Ltree.AddData(v.Bytes())
	}
	err := Ltree.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Println("tree", Ltree.Height(), Ltree.Root())

	vr.Lroot = Ltree.Root()
	fmt.Println("verifier init time", elapsed)
}

func (vr *Verifier) Verify(x interface{}, sol Solution) bool {
	t1 := time.Now()
	L_ind, S_x := generateChallenge(vr.T, vr.B, vr.Lambda, x)

	// use merkle proofs to verify L_x
	tmpTree := gomerkle.NewTree(sha256.New())
	tmpTree.AddHash(vr.Lroot)
	tmpTree.Generate()
	fmt.Println("Lroot", tmpTree.Height(), tmpTree.Root())
	for i, _ := range L_ind {
		hashv := sha256.Sum256(sol.L_x[i].Bytes())
		yes := tmpTree.VerifyProof(sol.MerkleProof[i], vr.Lroot, hashv[:])
		if !yes {
			return false
		}
	}

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

	y := sol.Y
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
