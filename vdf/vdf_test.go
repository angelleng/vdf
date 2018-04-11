package vdf

import (
	"fmt"
	"testing"
)

func TestGenerateChallenge(t *testing.T) {
	ind_L, ind_S := generateChallenge(1000, 10000, 100, 2)
	t.Log("test\n")
	t.Log(ind_L)
	fmt.Println(ind_S)
}
