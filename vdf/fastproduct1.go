package vdf

import (
	"math/big"
	"runtime"
	"sync"
)

func product1(array []*big.Int, mod *big.Int) (prod *big.Int) {
	type pair struct {
		a *big.Int
		b *big.Int
	}
	singles := make(chan *big.Int, len(array))
	doubles := make(chan pair, len(array)/2)

	go func() {
		for _, v := range array {
			// fmt.Println("add ", v)
			singles <- v
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		for i := 0; i < len(array)-1; i++ {
			a := <-singles
			b := <-singles
			doubles <- pair{a, b}
		}
		prod = <-singles
		close(doubles)
		wg.Done()
	}()

	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			for {
				ab, ok := <-doubles
				if !ok {
					return
				}
				a := ab.a
				b := ab.b
				result := new(big.Int).Mul(a, b)
				result.Mod(result, mod)
				singles <- result
			}
		}()
	}
	wg.Wait()
	return
}
