package vdf

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestPrime(t *testing.T) {

	t.Error("test regular prime")
	p, _ := rand.Prime(rand.Reader, 200)

	fmt.Println(p)
	fmt.Println(isStrongPrime(p))
}

func TestStrong(t *testing.T) {
	t.Error("test strong prime")
	p := StrongPrime(200)
	fmt.Println(p)
	fmt.Println(isStrongPrime(p))
}
