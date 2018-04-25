package vdf

import (
	"container/heap"
	"fmt"
	"math/big"
	"sync"
	"time"
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
	workers := 2
	singles := make(chan *big.Int, len(array)/2)
	big2 := make(chan *big.Int, 1)
	doubles := make(chan pair, 1)
	start := time.Now()

	go func() {
		pq := PriorityQueue(append([]*big.Int(nil), array...))
		heap.Init(&pq)
		for {
			select {
			case item, ok := <-singles:
				if !ok {
					return
				}
				fmt.Printf("read from singles: %v, time: %v\n", item.BitLen(), time.Now().Sub(start))
				// heap.Push(&pq, item)
				if item.Cmp(pq[0]) >= 0 { // input is larger than all
					if len(big2) == cap(big2) {
						replaced := false
						for i := 0; i < len(big2); i++ {
							b := <-big2
							if item.Cmp(b) > 0 {
								big2 <- item
								heap.Push(&pq, b)
								replaced = true
								break
							} else {
								big2 <- b
							}
						}
						if !replaced {
							heap.Push(&pq, item)
						}
					} else {
						big2 <- item
					}
				} else {
					t1 := time.Now()
					heap.Push(&pq, item)
					fmt.Printf("sorting takes: %v, length: %v\n", time.Now().Sub(t1), len(pq))
				}
			default:
				if len(pq) != 0 {
					select {
					case big2 <- pq[0]:
						fmt.Printf("push to big2: %v, time: %v\n", pq[0].BitLen(), time.Now().Sub(start))
						heap.Pop(&pq)
					default:
					}
				}
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() { // grab two elements from singles to feed to workers
		for i := 0; i < len(array)-1; i++ {
			a := <-big2
			b := <-big2
			doubles <- pair{a, b}
		}
		close(doubles)
		fmt.Println("reading result")
		prod = <-big2
		fmt.Println("result read")
		close(singles)
		wg.Done()
	}()

	for i := 0; i < workers; i++ { // workers multiplies two elements passed from doubles
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
					tnow := time.Now()
					elapsed2 := tnow.Sub(start)
					fmt.Printf("length (%v, %v), worker %v, %v since start.\n", a.BitLen(), b.BitLen(), i, elapsed2)
					singles <- big.NewInt(0).Mul(a, b)
				default:
					tnow := time.Now()
					elapsed2 := tnow.Sub(start)
					fmt.Printf("waiting on doubles, worker %v, %v since start.\n", i, elapsed2)
					// fmt.Printf("waiting on doubles, worker %v\n", i)
					time.Sleep(1 * time.Microsecond)
				}

				// ab, ok := <-doubles
				// if !ok {
				// 	fmt.Println("doubles closed")
				// 	return
				// }
				// a := ab.a
				// b := ab.b
				// // tnow := time.Now()
				// // elapsed2 := tnow.Sub(start)
				// fmt.Printf("length (%v, %v), worker %v, %v since start.\n", a.BitLen(), b.BitLen(), i, time.Now().Sub(start))
				// // fmt.Printf("length (%v, %v), worker %v\n", a.BitLen(), b.BitLen(), i)
				// t1 := time.Now()
				// result := big.NewInt(0).Mul(a, b)
				// fmt.Printf("compute time: %v\n", time.Now().Sub(t1))
				// fmt.Printf("length (%v, %v), worker %v finished, %v since start.\n", a.BitLen(), b.BitLen(), i, time.Now().Sub(start))
				// singles <- result
			}
		}(i)
	}
	wg.Wait()
	return
}
