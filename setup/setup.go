package main

import (
	"flag"
	"fmt"
	"time"
	"vdf"
)

func main() {
	var t, B, lambda, keysize int
	var gsPath = flag.String("gspath", "gs", "path to gs")
	var NPath = flag.String("npath", "N", "path to N")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")

	flag.Parse()

	start := time.Now()
	vdf.Setup(t, B, lambda, keysize, *gsPath, *NPath)
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("setup time", elapsed)
}
