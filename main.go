package main

import (
	"fmt"
	"log"
	"time"
	"vdf"

	_ "net/http/pprof"

	"net/http"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	start := time.Now()
	evaluateKey, verifyKey := vdf.Setup()
	t1 := time.Now()
	elapsed := t1.Sub(start)
	fmt.Println("setup time", elapsed)
	solution := vdf.Evaluate(evaluateKey, 1)
	t2 := time.Now()
	elapsed = t2.Sub(t1)
	fmt.Println("evaluate time", elapsed)
	success := vdf.Verify(verifyKey, solution, 1)
	t3 := time.Now()
	elapsed = t3.Sub(t2)
	fmt.Println("verify time", elapsed)
	fmt.Println("result: ", success)
	fmt.Println("finish")
}
