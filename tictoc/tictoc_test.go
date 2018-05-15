package tictoc

import (
	"testing"
	"time"
)

func TestTic(t *testing.T) {
	tic := NewTic()
	tic.Toc("")
	time.Sleep(1 * time.Second)
	tic.Toc("")
}
