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
	var veriKeyPath = flag.String("verikeypath", "veri.key", "path to verify key")
	var veriStoragePath = flag.String("veristoragepath", "veri.storage", "path to verifier storage")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")

	flag.Parse()

	verifyKey := new(vdf.VerifyKey)
	file, _ := os.Open(*veriKeyPath)
	decoder := gob.NewDecoder(file)
	decoder.Decode(verifyKey)
	file.Close()

	t1 := time.Now()

	var verifier vdf.Verifier
	verifier.Init(t, B, lambda, verifyKey)
	t2 := time.Now()
	elapsed := t2.Sub(t1)
	fmt.Println("verify init time", elapsed)

	file, _ = os.Create(*veriStoragePath)
	encoder := gob.NewEncoder(file)
	encoder.Encode(&verifier)
	file.Close()
}
