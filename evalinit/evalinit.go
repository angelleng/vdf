package main

import (
	"flag"
	"tictoc"
	"vdf"
)

func main() {
	var t, B, lambda, keysize, omit int

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")
	flag.IntVar(&omit, "omit", 0, "omit")
	flag.Parse()

	tic := tictoc.NewTic()
	vdf.EvalInit(t, omit)
	tic.Toc("evaluate init time")
}
