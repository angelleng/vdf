package vdf

import (
	"fmt"
	"testing"
)

func TestGenerateChallenge(t *testing.T) {
	ind_L, ind_S := generateChallenge(2)
	// t.Errorf("test")
	t.Log("test\n")
	t.Log(ind_L)
	fmt.Println(ind_S)
}
