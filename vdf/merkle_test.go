package vdf

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestMerkleMakeParent(t *testing.T) {
	length := 5
	fmt.Println(length)

	L := computeL(length)
	Lhash := make([][]byte, 0)
	for _, v := range L {
		hash := sha256.Sum256(v.Bytes())
		Lhash = append(Lhash, hash[:])
	}
	parent := makeParentLevel(Lhash)
	fmt.Println(Lhash)
	fmt.Println(parent)
}

func TestMerkleProof(t *testing.T) {
	length := 5
	fmt.Println(length)

	L := computeL(length)
	Lhash := make([][]byte, 0)
	for _, v := range L {
		hash := sha256.Sum256(v.Bytes())
		Lhash = append(Lhash, hash[:])
	}
	tree, root := makeTree(Lhash, 0)
	fmt.Println(tree)
	fmt.Println(root)

	id := 4
	proof := getProof(id, tree)
	result := verifyProof(L[id].Bytes(), root[0], proof, id)
	fmt.Println(result)
}
