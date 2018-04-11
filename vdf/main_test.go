package vdf

import (
	"fmt"
	"testing"
	"time"
)

func TestMain(test *testing.T) {
	t := 1000
	B := 100000
	lambda := 100
	keysize := 248
	start := time.Now()
	evaluateKey, verifyKey := Setup(t, B, lambda, keysize)
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("setup time", elapsed)
	solution := Evaluate(t, B, lambda, evaluateKey, 1)
	t2 := time.Now()
	elapsed = t2.Sub(t1)
	fmt.Println("evaluate time", elapsed)
	success := Verify(t, B, lambda, verifyKey, solution, 1)
	t3 := time.Now()
	elapsed = t3.Sub(t2)
	fmt.Println("verify time", elapsed)
	fmt.Println("result: ", success)
	fmt.Println("finish")
}
