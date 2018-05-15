package tictoc

import (
	"fmt"
	"time"
)

type Tic struct {
	start time.Time
}

func NewTic() (tic *Tic) {
	tic = &Tic{time.Now()}
	return
}

func (t *Tic) Tic() {
	t.start = time.Now()
}

// func (t *Tic) Tic(name string) {
// 	t.start = time.Now()
// 	fmt.Println("starting to", name)
// }

// func (t *Tic) Toc() {
// 	t2 := time.Now()
// 	elapsed := t2.Sub(t.start)
// 	t.start = t2
// 	fmt.Println("duration:", elapsed)
// }

func (t *Tic) Toc(name string) {
	t2 := time.Now()
	elapsed := t2.Sub(t.start)
	t.start = t2
	fmt.Println(name, elapsed)
}
