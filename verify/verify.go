package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"vdf"
)

func main() {
	var t, B, lambda, keysize int
	var veriKeyPath = flag.String("verikeypath", "veri.key", "path to verify key")
	var veriStoragePath = flag.String("veristoragepath", "veri.storage", "path to verifier storage")
	var challengePath = flag.String("challengepath", "challenge.txt", "path to challenge")
	var solutionPath = flag.String("solutionpath", "solution.txt", "path to solution")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")

	flag.Parse()

	verifyKey := new(vdf.VerifyKey)
	file, err := os.Open(*veriKeyPath)
	if err != nil {
		fmt.Println(err)
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(verifyKey)
	if err != nil {
		fmt.Println(err)
	}
	file.Close()
	fmt.Println("verify key", verifyKey)

	verifier := new(vdf.Verifier)
	file, _ = os.Open(*veriStoragePath)
	decoder = gob.NewDecoder(file)
	decoder.Decode(verifier)
	file.Close()
	fmt.Println("verifier", verifier)

	verifier.Hash = func(input *big.Int) (hashval *big.Int) {
		return vdf.Hashfunc(input, verifier.N)
	}

	challenge, err := ioutil.ReadFile(*challengePath)
	fmt.Println("challenge", challenge)

	var solution *big.Int
	file, _ = os.Open(*solutionPath)
	decoder = gob.NewDecoder(file)
	decoder.Decode(&solution)
	file.Close()
	fmt.Println("solution", solution)

	success := verifier.Verify(challenge, solution)
	fmt.Println("result: ", success)
}
