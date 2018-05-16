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

func (t *Tic) Toc(name string) {
	t2 := time.Now()
	elapsed := t2.Sub(t.start)
	t.start = t2
	fmt.Println(name, elapsed)
}
