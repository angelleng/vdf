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
	workers := 2
	singles := make(chan *big.Int, len(array)/2)
	big2 := make(chan *big.Int, 1)
	doubles := make(chan pair, 1)
	// start := time.Now()

	pq := PriorityQueue(append([]*big.Int(nil), array...))
	heap.Init(&pq)
	// biggest := make([]*big.Int, workers*2)
	// for i, _ := range biggest {
	// 	biggest[i] = heap.Pop(&pq).(*big.Int)
	// }
	// mu := &sync.Mutex{}

	// go func() {
	// 	for {
	// 		item, ok := <-singles
	// 		mu.Lock()
	// 		if !ok {
	// 			return
	// 		}
	// 		fmt.Println("read from singles", item.BitLen())
	// 		fmt.Println("lock and sort")
	// 		heap.Push(&pq, item)
	// 		mu.Unlock()
	// 	}
	// }()

	// go func() {
	// 	for {
	// 		if len(pq) != 0 {
	// 			mu.Lock()
	// 			big2 <- pq[0]
	// 			fmt.Println("push to big2", pq[0].BitLen())
	// 			heap.Pop(&pq)
	// 			mu.Unlock()
	// 		} else {
	// 			fmt.Println("pq is empty")
	// 			time.Sleep(1 * time.Microsecond)
	// 		}
	// 	}
	// }()

	go func() {
		for {
			select {
			case item, ok := <-singles:
				if !ok {
					return
				}
				// fmt.Println("read from singles", item.BitLen())
				heap.Push(&pq, item)
			default:
			}
			if len(pq) != 0 {
				select {
				case big2 <- pq[0]:
					// fmt.Println("push to big2", pq[0].BitLen())
					heap.Pop(&pq)
				default:
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
			// sent := false
			// for !sent {
			// 	select {
			// 	case doubles <- pair{a, b}:
			// 		sent = true
			// 	default:
			// 		// tnow := time.Now()
			// 		// elapsed := tnow.Sub(start)
			// 		// fmt.Println("waiting to send to double, time: ", elapsed)
			// 		// fmt.Println("waiting to send to double")
			// 		// time.Sleep(2 * time.Second)
			// 	}
			// }
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
				// select {
				// case ab, ok := <-doubles:
				// 	if !ok {
				// 		fmt.Println("doubles closed")
				// 		return
				// 	}
				// 	a := ab.a
				// 	b := ab.b
				// 	tnow := time.Now()
				// 	elapsed2 := tnow.Sub(start)
				// 	fmt.Printf("length (%v, %v), worker %v, %v since start.\n", a.BitLen(), b.BitLen(), i, elapsed2)
				// 	singles <- big.NewInt(0).Mul(a, b)
				// default:
				// 	tnow := time.Now()
				// 	elapsed2 := tnow.Sub(start)
				// 	fmt.Printf("waiting on doubles, worker %v, %v since start.\n", i, elapsed2)
				// 	// fmt.Printf("waiting on doubles, worker %v\n", i)
				// 	time.Sleep(1 * time.Millisecond)
				// }

				ab, ok := <-doubles
				if !ok {
					fmt.Println("doubles closed")
					return
				}
				a := ab.a
				b := ab.b
				// tnow := time.Now()
				// elapsed2 := tnow.Sub(start)
				// fmt.Printf("length (%v, %v), worker %v, %v since start.\n", a.BitLen(), b.BitLen(), i, elapsed2)
				// fmt.Printf("length (%v, %v), worker %v\n", a.BitLen(), b.BitLen(), i)
				singles <- big.NewInt(0).Mul(a, b)
			}
		}(i)
	}
	wg.Wait()
	return
}
