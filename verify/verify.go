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
	var t, B, lambda, keysize, omitHeight int
	var veriKeyPath = flag.String("verikeypath", "veri.key", "path to verify key")
	var veriStoragePath = flag.String("veristoragepath", "veri.storage", "path to verifier storage")
	var challengePath = flag.String("challengepath", "challenge.txt", "path to challenge")
	var solutionPath = flag.String("solutionpath", "solution.txt", "path to solution")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")
	flag.IntVar(&omitHeight, "omit", 0, "omit")

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

	verifier := new(vdf.Verifier)
	file, _ = os.Open(*veriStoragePath)
	decoder = gob.NewDecoder(file)
	decoder.Decode(verifier)
	file.Close()

	verifier.Hash = func(input *big.Int) (hashval *big.Int) {
		return vdf.Hashfunc(input, verifier.N)
	}

	challenge, err := ioutil.ReadFile(*challengePath)

	solution := new(vdf.Solution)
	file, _ = os.Open(*solutionPath)
	decoder = gob.NewDecoder(file)
	decoder.Decode(solution)
	file.Close()
	// fmt.Println(solution)

	success := verifier.Verify(challenge, omitHeight, *solution)
	fmt.Println("result: ", success)
}
