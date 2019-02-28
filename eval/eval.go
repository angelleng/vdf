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
	var solPath = flag.String("solpath", "solution", "path to solution")
	var gsPath = flag.String("gspath", "gs", "path to gs")
	var NPath = flag.String("npath", "N", "path to N")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")
	flag.IntVar(&omitHeight, "omit", 0, "omit")

	flag.Parse()

	challenge, _ := ioutil.ReadFile(*challengePath)
	fmt.Println("challenge", challenge)

	tic := tictoc.NewTic()
	vdf.Evaluate(t, B, lambda, omitHeight, *NPath, challenge, *gsPath, *solPath)
	tic.Toc("evaluate time")
}
