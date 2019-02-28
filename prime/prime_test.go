package prime

import (
	"fmt"
	"testing"
)

func TestPrimeLargeSize(t *testing.T) {
	primes := Primes(16485863)
	fmt.Println("length: ", len(primes))
	primes = Primes(2038074751)
	fmt.Println("length: ", len(primes))
	primes = Primes(22801763513)
	fmt.Println("length: ", len(primes))
	primes = Primes(252097800629)
	fmt.Println("length: ", len(primes))
}
