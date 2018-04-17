package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
	"vdf"
)

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

func main() {

	var t, B, lambda, keysize int
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	flag.IntVar(&t, "t", 100, "write memory profile to this file")
	flag.IntVar(&B, "B", 10000, "write memory profile to this file")
	flag.IntVar(&lambda, "lambda", 100, "write memory profile to this file")
	flag.IntVar(&keysize, "keysize", 2048, "write memory profile to this file")

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	start := time.Now()
	evaluateKey, verifyKey := vdf.Setup(t, B, lambda, keysize)
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("setup time", elapsed)

	var evaluator vdf.Evaluator
	var verifier vdf.Verifier
	evaluator.Init(t, B, lambda, evaluateKey)
	verifier.Init(t, B, lambda, verifyKey)

	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(verifier)
	data := w.Bytes()
	// fmt.Println("verifier storage size: ", len(data))
	fmt.Println("verifier storage size: ", HumanSize(len(data)))

	w2 := new(bytes.Buffer)
	e2 := gob.NewEncoder(w2)
	e2.Encode(evaluator)
	data2 := w2.Bytes()
	// fmt.Println("evaluator storage size: ", len(data2))
	fmt.Println("evaluator storage size: ", HumanSize(len(data2)))

	for challenge := 0; challenge < 1; challenge++ {
		// solution := vdf.Evaluate(t, B, lambda, evaluateKey, challenge)
		t1 = time.Now()
		solution2 := evaluator.Eval(challenge)
		t2 := time.Now()
		elapsed = t2.Sub(t1)
		fmt.Println("evaluate time", elapsed)

		// success := vdf.Verify(t, B, lambda, verifyKey, solution, challenge)
		success2 := verifier.Verify(challenge, solution2)
		t3 := time.Now()
		elapsed = t3.Sub(t2)
		fmt.Println("verify time", elapsed)
		// fmt.Println("result: ", success)
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
