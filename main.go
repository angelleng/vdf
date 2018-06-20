package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"tictoc"
	"vdf"
)

func main() {
	var t, B, lambda, keysize, omitHeight int
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	var gsPath = flag.String("gspath", "gs", "path to gs")
	var NPath = flag.String("npath", "N", "path to N")
	var rootsPath = flag.String("rootspath", "roots", "path to merkle roots")
	var solPath = flag.String("solpath", "solution", "path to solution")

	flag.IntVar(&t, "t", 100, "write memory profile to this file")
	flag.IntVar(&B, "B", 10000, "write memory profile to this file")
	flag.IntVar(&lambda, "lambda", 100, "write memory profile to this file")
	flag.IntVar(&keysize, "keysize", 2048, "write memory profile to this file")
	flag.IntVar(&omitHeight, "omit", 0, "omit")

	flag.Parse()

	fmt.Printf("parameters: ")
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	tic := tictoc.NewTic()
	vdf.Setup(t, B, lambda, keysize, *gsPath, *NPath)
	tic.Toc("setup time:")

	vdf.EvalInit(t, omitHeight)
	vdf.VeriInit(t, omitHeight, *rootsPath)

	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	// e.Encode(verifier)
	// fmt.Printf("verifier storage size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())

	// w.Reset()
	// e.Encode(evaluator)
	// fmt.Printf("evaluator storage size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())

	// w.Reset()
	// e.Encode(evaluateKey)
	// fmt.Printf("eval key size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())

	// 	w.Reset()
	// 	e.Encode(evaluator.L)
	// 	fmt.Printf("L size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())
	//
	// 	// w.Reset()
	// 	// e.Encode(evaluator.Ltree)
	// 	// fmt.Printf("merkle tree size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())
	//
	// 	bitlen := 0
	// 	for _, v := range evaluator.L {
	// 		bitlen += v.BitLen()
	// 	}
	// 	fmt.Println("elements in L size: ", bitlen)

	for challenge := 0; challenge < 1; challenge++ {
		// solution := vdf.Evaluate(t, B, lambda, evaluateKey, challenge)
		tic.Tic()
		solution2 := vdf.Evaluate(t, B, lambda, omitHeight, *NPath, challenge, *gsPath, *solPath)
		tic.Toc("evaluate time:")

		fmt.Println("")
		w.Reset()
		e.Encode(solution2.L_x)
		fmt.Printf("solution.L_x size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())
		w.Reset()
		e.Encode(solution2.Y)
		fmt.Printf("solution.Y size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())
		w.Reset()
		e.Encode(solution2.MerkleProof)
		fmt.Printf("solution.MerkleProof size: %v (%v B)\n", vdf.HumanSize(w.Len()), w.Len())
		// success := vdf.Verify(t, B, lambda, verifyKey, solution, challenge)
		tic.Tic()
		success2 := vdf.Verify(t, B, lambda, omitHeight, *NPath, *rootsPath, challenge, *solPath)
		tic.Toc("verify time:")
		fmt.Println("result: ", success2)
		fmt.Println("finish")
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}
}
