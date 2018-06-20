package vdf

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
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
	perFile = 1 << 30
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

func computeAndStoreGs(hash func(input *big.Int) *big.Int, B int, P_inv *big.Int, N *big.Int, gsPath string) {
	nFiles := B / perFile
	lastFile := B % perFile
	if lastFile != 0 {
		fmt.Println("number of files for gs", nFiles+1)
	} else {
		fmt.Println("number of files for gs", nFiles)
	}
	bytesPerBig := int(math.Ceil(float64(N.BitLen()) / 8))

	for i := 0; i <= nFiles; i++ {
		filename := gsPath + strconv.Itoa(i)
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

func readFileAndComputeGx(S_x []int, gsPath string, B int, N *big.Int) *big.Int {
	sort.Ints(S_x)
	// perFile := 1 << 20
	nFiles := B / perFile
	bytesPerBig := int(math.Ceil(float64(N.BitLen()) / 8))

	gx := make([]*big.Int, len(S_x))
	for i := 0; i <= nFiles; i++ {
		filename := gsPath + strconv.Itoa(i)
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

// interface
func Setup(t, B, lambda, keysize int, gsPath, NPath string) {
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

	file, _ := os.Create(NPath)
	file.Write(N.Bytes())
	file.Close()

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
	computeAndStoreGs(hashfunc, B, P_inv, N, gsPath)
	tic.Toc("compute and store gs time")
}

type Solution struct {
	Y           *big.Int   // solution
	L_x         []*big.Int // primes
	MerkleProof [][]byte   // proof of merkle tree
}

func EvalInit(t, omitHeight int) {
	tic := tictoc.NewTic()
	L := computeL(t)
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(L)
	fmt.Printf("L size: %v (%v B)\n", HumanSize(w.Len()), w.Len())
	fmt.Println("merkle tree size:", HumanSize(32*t*2))
	bitlen := 0
	for _, v := range L {
		bitlen += v.BitLen()
	}
	fmt.Println("elements in L size:", bitlen)

	lpath := "L"
	treepath := "Ltree"
	lfile, _ := os.Create(lpath)
	defer lfile.Close()
	for _, v := range L {
		data := bigToFixedLengthBytes(v, 2*log2(t))
		lfile.Write(data)
	}
	_ = MakeTreeOnDiskFromData(L, omitHeight, treepath)

	tic.Toc("evaluator init time:")
}

func Evaluate(t, B, lambda, omitHeight int, NPath string, x interface{}, gsPath, solPath string) (sol Solution) {
	fmt.Println("\nEVAL")
	tic := tictoc.NewTic()
	L_ind, S_x := generateChallenge(t, B, lambda, x)
	tic.Toc("generate challenge time:")

	N := new(big.Int)
	dat, err := ioutil.ReadFile(NPath)
	check(err)
	N.SetBytes(dat)

	y := readFileAndComputeGx(S_x, gsPath, B, N)
	tic.Toc("read and compute g_x time:")
	// fmt.Println("g_x", y)

	L := make([]*big.Int, t)

	f, err := os.Open("L")
	defer f.Close()
	check(err)
	for i := range L {
		buf := make([]byte, 2*log2(t))
		f.Read(buf)
		L[i] = big.NewInt(0).SetBytes(buf)
	}

	tic.Toc("read L time:")

	tic.Tic()
	proofs := GetBatchProofFromDisk(L_ind, "Ltree", t, omitHeight)
	tic.Toc("generate merkle proof takes:")

	L_x := make([]*big.Int, lambda)
	for i, v := range L_ind {
		buf := make([]byte, 2*log2(t))
		f.ReadAt(buf, int64(v*2*log2(t)))
		L_x[i] = big.NewInt(0).SetBytes(buf)
		// L_x[i] = L[v]
	}

	tic.Tic()
	for i, v := range L { // TODO read L from disk
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
		y.Exp(y, v, N)
	}
	tic.Toc("actual evaluate time:")

	// fmt.Println("y", y)
	sol.Y = y
	sol.L_x = L_x
	sol.MerkleProof = proofs

	file, _ := os.Create(solPath)
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(sol)
	check(err)
	return
}

func VeriInit(t, omitHeight int, rootsPath string) {
	tic := tictoc.NewTic()
	tic2 := tictoc.NewTic()
	L := computeL(t)
	tic.Toc("compute L time:")

	_, Lroots := MakeTreeFromData(L, omitHeight)
	tic.Toc("compute merkle tree time:")
	tic2.Toc("verify init time:")

	file, _ := os.Create(rootsPath)
	encoder := gob.NewEncoder(file)
	err := encoder.Encode(Lroots)
	check(err)
}

func Verify(t, B, lambda, omitHeight int, NPath string, rootsPath string, x interface{}, solPath string) bool {
	N := new(big.Int)
	dat, err := ioutil.ReadFile(NPath)
	check(err)
	N.SetBytes(dat)
	hashfun := func(input *big.Int) (hashval *big.Int) {
		return Hashfunc(input, N)
	}

	roots := new([][]byte)
	file, err := os.Open(rootsPath)
	check(err)
	decoder := gob.NewDecoder(file)
	decoder.Decode(roots)
	file.Close()

	sol := new(Solution)
	file, err = os.Open(solPath)
	check(err)
	decoder = gob.NewDecoder(file)
	decoder.Decode(sol)
	file.Close()

	fmt.Println("omitHeight:", omitHeight)
	fmt.Println("\nVERIFY")
	tic := tictoc.NewTic()
	L_ind, S_x := generateChallenge(t, B, lambda, x)
	tic.Toc("generate challenge time:")
	height := fullMerkleHeight(t)
	fmt.Println(height)

	if !VerifyBatchProof(L_ind, sol.L_x, *roots, sol.MerkleProof, height-omitHeight) {
		return false
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
		h := hashfun(big.NewInt(int64(v)))
		h_x.Mul(h_x, h)
		h_x.Mod(h_x, N)
	}
	h2 := big.NewInt(1)
	tic.Toc("compute hx time:")

	y := sol.Y
	h2.Exp(y, P_x, N)
	tic.Toc("actual verification time:")

	// fmt.Println("h", h_x)
	// fmt.Println("h2", h2)
	compare := h_x.Cmp(h2)
	return compare == 0
}
