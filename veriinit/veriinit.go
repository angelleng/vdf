package main

import (
	"flag"
	"tictoc"
	"vdf"
)

func main() {
	var t, B, lambda, keysize, omitHeight int
	var rootsPath = flag.String("rootspath", "roots", "path to merkle roots")

	flag.IntVar(&t, "t", 1000000, "t")
	flag.IntVar(&B, "B", 1000000, "B")
	flag.IntVar(&lambda, "lambda", 1000, "lambda")
	flag.IntVar(&keysize, "keysize", 512, "keysize")
	flag.IntVar(&omitHeight, "omit", 0, "omit")

	flag.Parse()

	tic := tictoc.NewTic()
	vdf.VeriInit(t, omitHeight, *rootsPath)
	tic.Toc("verify init time")

}
