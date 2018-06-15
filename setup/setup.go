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
	var evalKeyPath = flag.String("evalkeypath", "eval.key", "path to evaluation key")
	var veriKeyPath = flag.String("verikeypath", "veri.key", "path to verification key")
	var gsPath = flag.String("gspath", "gs", "path to gs")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")

	flag.Parse()

	start := time.Now()
	evaluateKey, verifyKey := vdf.Setup(t, B, lambda, keysize, *gsPath)
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("setup time", elapsed)

	file, err := os.Create(*evalKeyPath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		err = encoder.Encode(evaluateKey)
		if err != nil {
			fmt.Println(err)
		}
	}
	file.Close()

	file, err = os.Create(*veriKeyPath)
	if err == nil {
		encoder := gob.NewEncoder(file)
		encoder.Encode(verifyKey)
	} else {
		fmt.Println(err)
	}
	file.Close()
	// fmt.Printf("verifier storage size: %v (%v B)\n", HumanSize(w.Len()), w.Len())
}
