package main

import (
	"fmt"
	"vdf"
)

func main() {
	evaluateKey, verifyKey := vdf.Setup()
	solution := vdf.Evaluate(evaluateKey, 1)
	success := vdf.Verify(verifyKey, solution, 1)
	fmt.Println("result: ", success)
	fmt.Println("finish")
}
