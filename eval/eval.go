package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	"vdf"
)

func main() {
	var t, B, lambda, keysize, omitHeight int
	var evalStoragePath = flag.String("evalstoragepath", "eval.storage", "path to public key")
	var challengePath = flag.String("challengepath", "challenge.txt", "path to challenge")
	var solutionPath = flag.String("solutionpath", "solution.txt", "path to solution")
	var gsPath = flag.String("gspath", "gs", "path to gs")
	var NPath = flag.String("npath", "N", "path to N")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")
	flag.IntVar(&omitHeight, "omit", 0, "omit")

	flag.Parse()

	evaluator := new(vdf.Evaluator)
	file, _ := os.Open(*evalStoragePath)
	decoder := gob.NewDecoder(file)
	decoder.Decode(evaluator)
	file.Close()

	challenge, _ := ioutil.ReadFile(*challengePath)
	fmt.Println("challenge", challenge)

	t1 := time.Now()

	solution := evaluator.Eval(t, B, lambda, omitHeight, *NPath, challenge, *gsPath)
	t2 := time.Now()
	elapsed := t2.Sub(t1)
	fmt.Println("evaluate time", elapsed)

	file, _ = os.Create(*solutionPath)
	encoder := gob.NewEncoder(file)
	encoder.Encode(&solution)
	file.Close()
}
