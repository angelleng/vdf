package vdf

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/bits"
	"os"
	"prime"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"tictoc"
	"time"
)

const (
	perFile    = 1 << 30
	omitHeight = 0
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
	nFiles := B / perFile
	lastFile := B % perFile
	if lastFile != 0 {
		fmt.Println("number of files for gs", nFiles+1)
	} else {
		fmt.Println("number of files for gs", nFiles)
	}
	bytesPerBig := int(math.Ceil(float64(N.BitLen()) / 8))

	for i := 0; i <= nFiles; i++ {
		filename := gspath + strconv.Itoa(i)
		file, _ := os.Create(filename)
		var thisFile int
		if i < nFiles {
			thisFile = perFile
		} else if lastFile != 0 {
			thisFile = lastFile
		} else {
			return
		}

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
						gi := big.NewInt(0).Exp(v, P_inv, N)
						data := bigToFixedLengthBytes(gi, bytesPerBig)
						file.WriteAt(data, int64((ind%perFile)*bytesPerBig))
						wg.Done()
					} else {
						return
					}
					if ind%10000000 == 0 {
						file.Sync()
					}
				}
			}()
		}
		wg.Wait()
		file.Close()
	}
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

func readFileAndComputeGx(S_x []int, gspath string, B int, N *big.Int) *big.Int {
	sort.Ints(S_x)
	// perFile := 1 << 20
	nFiles := B / perFile
	bytesPerBig := int(math.Ceil(float64(N.BitLen()) / 8))

	gx := make([]*big.Int, len(S_x))
	for i := 0; i <= nFiles; i++ {
		filename := gspath + strconv.Itoa(i)
		file, _ := os.Open(filename)
		for j, v := range S_x {
			if v >= i*perFile && v < (i+1)*perFile {
				offset := v % perFile
				buf := make([]byte, bytesPerBig)
				_, err := file.Seek(int64(offset*bytesPerBig), 0)
				if err != nil {
					panic(err)
				}
				_, err = file.Read(buf)
				if err != nil {
					panic(err)
				}
				gx[j] = big.NewInt(0).SetBytes(buf)
			}
		}
		file.Close()
	}

	y := big.NewInt(1)
	for _, v := range gx {
		y.Mul(y, v)
		y.Mod(y, N)
	}
	return y
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
				candidate, _ := cryptorand.Prime(cryptorand.Reader, keysize)
				if isStrongPrime(candidate) {
					primechan <- candidate
				}
			}
		}()
	}
	p = <-primechan
	// fmt.Println("have one prime")
	q = <-primechan
	found = true
	// fmt.Println("have second prime")
	return
}

// interface
type EvalKey struct {
	G *big.Int
	H func(*big.Int) *big.Int
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

	tic := tictoc.NewTic()
	L := computeL(t)
	tic.Toc("compute L time:")
	// fmt.Printf("L [%v %v %v %v ... %v %v %v] \n", L[0], L[1], L[2], L[3], L[len(L)-3], L[len(L)-2], L[len(L)-1])

	p, q := generateTwoGoodPrimes(keysize)
	tic.Toc("find p, q time:")
	N := new(big.Int).Mul(p, q)

	// fmt.Println("p and q", p, q)
	// fmt.Println("N ", N)
	tmp := big.NewInt(0)
	tmp.Add(p, big.NewInt(-1))
	phi := big.NewInt(0)
	phi.Add(q, big.NewInt(-1))
	phi.Mul(phi, tmp)
	// fmt.Println("phi", phi)

	hashfunc := func(input *big.Int) (hashval *big.Int) {
		return Hashfunc(input, N)
	}

	tic.Tic()
	P := product1(L, phi)
	tic.Toc("compute P time:")
	P_inv := big.NewInt(1)
	P_inv.ModInverse(P, phi)

	tic.Tic()
	computeAndStoreGs(hashfunc, B, P_inv, N, "gs")
	tic.Toc("compute and store gs time")

	evaluateKey := EvalKey{N, hashfunc}
	verifyKey := VerifyKey{N, hashfunc}
	return &evaluateKey, &verifyKey
}

type Evaluator struct {
	T        int
	B        int
	Lambda   int
	N        *big.Int
	lfile    os.File
	treefile os.File
	// L      []*big.Int
	// Ltree  gomerkle.Tree
}

type Solution struct {
	Y           *big.Int   // solution
	L_x         []*big.Int // primes
	MerkleProof [][][]byte // proof of merkle tree
}

func log2(x int) int {
	var r int = 0
	for ; x > 1; x >>= 1 {
		r++
	}
	return r
}

func (ev *Evaluator) Init(t, B, lambda int, evaluateKey *EvalKey) {
	tic := tictoc.NewTic()
	ev.T = t
	ev.B = B
	ev.Lambda = lambda
	ev.N = evaluateKey.G
	L := computeL(t)

	lpath := "L"
	treepath := "Ltree"
	lfile, _ := os.Create(lpath)
	treefile, _ := os.Create(treepath)
	defer lfile.Close()
	defer treefile.Close()
	for _, v := range L {
		data := bigToFixedLengthBytes(v, 2*log2(t))
		lfile.Write(data)
	}
	// assume t is power of 2

	tic.Toc("evaluator init time:")
}

func (ev *Evaluator) Eval(x interface{}) (sol Solution) {
	fmt.Println("\nEVAL")
	tic := tictoc.NewTic()
	L_ind, S_x := generateChallenge(ev.T, ev.B, ev.Lambda, x)
	tic.Toc("generate challenge time:")

	y := readFileAndComputeGx(S_x, "gs", ev.B, ev.N)
	tic.Toc("read and compute g_x time:")
	// fmt.Println("g_x", y)

	L := computeL(ev.T)
	tic.Toc("compute L time:")

	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(L)
	fmt.Printf("L size: %v (%v B)\n", HumanSize(w.Len()), w.Len())
	fmt.Println("merkle tree size:", HumanSize(32*ev.T*2))

	bitlen := 0
	for _, v := range L {
		bitlen += v.BitLen()
	}
	fmt.Println("elements in L size:", bitlen)

	tic.Tic()
	tree, _ := makeTreeFromL(L, omitHeight)
	tic.Toc("generate merkle tree takes:")

	proofs := make([][][]byte, 0)
	for _, ind := range L_ind {
		proof := getProof(ind, tree)
		proofs = append(proofs, proof)
	}
	tic.Toc("generate merkle proof takes:")

	L_x := make([]*big.Int, ev.Lambda)
	for i, v := range L_ind {
		L_x[i] = L[v]
	}

	tic.Tic()
	for i, v := range L {
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
	tic.Toc("actual evaluate time:")

	// fmt.Println("y", y)
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
	Hash   func(*big.Int) *big.Int
	Lroots [][]byte
}

func (vr *Verifier) Init(t, B, lambda int, verifyKey *VerifyKey) {
	vr.T = t
	vr.B = B
	vr.Lambda = lambda
	vr.N = verifyKey.G
	tic := tictoc.NewTic()
	tic2 := tictoc.NewTic()
	L := computeL(t)
	tic.Toc("compute L time:")
	vr.Hash = verifyKey.H

	_, vr.Lroots = makeTreeFromL(L, omitHeight)
	tic.Toc("compute merkle tree time:")
	tic2.Toc("verify init time:")
}

func (vr *Verifier) Verify(x interface{}, sol Solution) bool {
	fmt.Println("omitHeight:", omitHeight)
	fmt.Println("\nVERIFY")
	tic := tictoc.NewTic()
	L_ind, S_x := generateChallenge(vr.T, vr.B, vr.Lambda, x)
	tic.Toc("generate challenge time:")
	height := fullMerkleHeight(vr.T)
	fmt.Println(height)

	perTree := 1 << uint(height-omitHeight)
	// perTree := (1 + vr.T) / len(vr.Lroots)
	fmt.Println(perTree)
	for i, v := range L_ind {
		rootId := v / perTree
		// fmt.Println(v, rootId)
		root := vr.Lroots[rootId]
		if !verifyProof(sol.L_x[i].Bytes(), root, sol.MerkleProof[i], v) {
			return false
		}
	}

	tic.Toc("verify merkle proof time:")

	L_x := sol.L_x
	P_x := big.NewInt(1)
	for _, v := range L_x {
		P_x.Mul(P_x, v)
	}
	tic.Toc("compute Px time:")

	h_x := big.NewInt(1)
	for _, v := range S_x {
		h := vr.Hash(big.NewInt(int64(v)))
		h_x.Mul(h_x, h)
		h_x.Mod(h_x, vr.N)
	}
	h2 := big.NewInt(1)
	tic.Toc("compute hx time:")

	y := sol.Y
	h2.Exp(y, P_x, vr.N)
	tic.Toc("actual verification time:")

	// fmt.Println("h", h_x)
	// fmt.Println("h2", h2)
	compare := h_x.Cmp(h2)
	return compare == 0
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
