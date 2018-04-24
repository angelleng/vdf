package vdf

import (
	"container/heap"
	"fmt"
	"math/big"
	"sync"
)

type PriorityQueue []*big.Int

func (h PriorityQueue) Len() int           { return len(h) }
func (h PriorityQueue) Less(i, j int) bool { return h[i].Cmp(h[j]) > 0 }
func (h PriorityQueue) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *PriorityQueue) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*big.Int))
}

func (h *PriorityQueue) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func product(array []*big.Int) (prod *big.Int) {
	type pair struct {
		a *big.Int
		b *big.Int
	}
	singles := make(chan *big.Int, len(array))
	big2 := make(chan *big.Int, len(array))
	doubles := make(chan pair, len(array)/2)
	// start := time.Now()

	go func() { // insert all elements to singles
		for i := len(array) - 1; i >= 0; i-- {
			singles <- array[i]
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() { // grab two elements from singles to feed to workers
		for i := 0; i < len(array)-1; i++ {
			a := <-big2
			b := <-big2
			sent := false
			for !sent {
				select {
				case doubles <- pair{a, b}:
					sent = true
				default:
					// tnow := time.Now()
					// elapsed := tnow.Sub(start)
					// fmt.Println("waiting to send to double, time: ", elapsed)
					fmt.Println("waiting to send to double")
					// time.Sleep(2 * time.Second)
				}
			}
		}
		close(doubles)
		fmt.Println("reading result")
		prod = <-big2
		fmt.Println("result read")
		close(singles)
		wg.Done()
	}()

	go func() {
		pq := &PriorityQueue{}
		heap.Init(pq)
		for {
			if len(*pq) != 0 {
				select {
				case item, ok := <-singles:
					if !ok {
						return
					}
					heap.Push(pq, item)
				case big2 <- (*pq)[0]:
					heap.Pop(pq)
				}
			} else {
				item := <-singles
				heap.Push(pq, item)
			}
		}
	}()

	for i := 0; i < 100; i++ { // workers multiplies two elements passed from doubles
		go func(i int) {
			for {
				select {
				case ab, ok := <-doubles:
					if !ok {
						fmt.Println("doubles closed")
						return
					}
					a := ab.a
					b := ab.b
					fmt.Printf("length (%v, %v), worker %v\n", a.BitLen(), b.BitLen(), i)
					singles <- big.NewInt(0).Mul(a, b)
				default:
					//tnow := time.Now()
					//elapsed2 := tnow.Sub(start)
					// fmt.Printf("waiting on doubles, worker %v, %v since start.\n", i, elapsed2)
					// fmt.Printf("waiting on doubles, worker %v\n", i)
					// time.Sleep(1 * time.Second)
				}

				// ab, ok := <-doubles
				// if !ok {
				// 	fmt.Println("doubles closed")
				// 	return
				// }
				// a := ab.a
				// b := ab.b
				// singles <- big.NewInt(0).Mul(a, b)
			}
		}(i)
	}
	wg.Wait()
	return
}
