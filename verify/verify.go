package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"tictoc"
	"vdf"
)

func main() {
	var t, B, lambda, keysize, omitHeight int
	var challengePath = flag.String("challengepath", "challenge.txt", "path to challenge")
	var NPath = flag.String("npath", "N", "path to N")
	var rootsPath = flag.String("rootspath", "roots", "path to merkle roots")
	var solPath = flag.String("solpath", "solution", "path to solution")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")
	flag.IntVar(&omitHeight, "omit", 0, "omit")

	flag.Parse()

	challenge, _ := ioutil.ReadFile(*challengePath)

	tic := tictoc.NewTic()
	success := vdf.Verify(t, B, lambda, omitHeight, *NPath, *rootsPath, challenge, *solPath)
	tic.Toc("verify time")
	fmt.Println("result: ", success)
}
