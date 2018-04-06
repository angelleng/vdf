package vdf

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestStrong(t *testing.T) {
	t.Error("test strong prime")
	p := StrongPrime(2048)
	fmt.Println(p)
}

func TestPrime(t *testing.T) {

	t.Error("test regular prime")
	p, _ := rand.Prime(rand.Reader, 2048)
	fmt.Println(p)
}
