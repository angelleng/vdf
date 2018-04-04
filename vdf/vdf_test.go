package vdf

import (
	"fmt"
	"testing"
)

func TestGenerateChallenge(t *testing.T) {
	ind_L, ind_S := generateChallenge(2)
	t.Errorf("test")
	fmt.Println("test")
	fmt.Println(ind_L)
	fmt.Println(ind_S)
}
