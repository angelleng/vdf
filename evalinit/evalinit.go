package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"os"
	"time"
	"vdf"
)

func main() {
	var t, B, lambda, keysize int
	var evalKeyPath = flag.String("evalkeypath", "eval.key", "path to public key")
	var evalStoragePath = flag.String("evalstoragepath", "eval.storage", "path to public key")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")

	flag.Parse()

	evaluateKey := new(vdf.EvalKey)
	file, _ := os.Open(*evalKeyPath)
	decoder := gob.NewDecoder(file)
	decoder.Decode(evaluateKey)
	file.Close()

	t1 := time.Now()

	evaluator := new(vdf.Evaluator)
	evaluator.Init(t, B, lambda, evaluateKey)

	t2 := time.Now()
	elapsed := t2.Sub(t1)
	fmt.Println("evaluate init time", elapsed)

	file, _ = os.Create(*evalStoragePath)
	encoder := gob.NewEncoder(file)
	encoder.Encode(evaluator)
	file.Close()
}
