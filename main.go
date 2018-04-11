package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"
	"vdf"
)

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
	solution := vdf.Evaluate(t, B, lambda, evaluateKey, 1)
	t2 := time.Now()
	elapsed = t2.Sub(t1)
	fmt.Println("evaluate time", elapsed)
	success := vdf.Verify(t, B, lambda, verifyKey, solution, 1)
	t3 := time.Now()
	elapsed = t3.Sub(t2)
	fmt.Println("verify time", elapsed)
	fmt.Println("result: ", success)
	fmt.Println("finish")

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
