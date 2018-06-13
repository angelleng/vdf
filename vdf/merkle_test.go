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

func TestMerklePath(t *testing.T) {
	length := 128
	fmt.Println(length)

	path := merklePath(56, length)
	fmt.Println(path)
}

func TestMerkleBatchProof(t *testing.T) {
	length := 128
	fmt.Println(length)
	L := computeL(length)
	tree, _ := makeTreeFromL(L, 0)

	list := []int{56, 10, 3, 90, 20}
	for _, v := range list {
		path := merklePath(v, length)
		fmt.Println(path)
	}
	getProofForAList(list, tree)
}

func TestMerkleBatchVerify(t *testing.T) {
	length := 129
	fmt.Println(length)
	L := computeL(length)
	tree, roots := makeTreeFromL(L, 0)

	list := []int{1, 2, 3}
	for _, v := range list {
		path := merklePath(v, length)
		fmt.Println(path)
	}
	proof := getProofForAList(list, tree)
	height := len(tree)
	fmt.Printf("\nVerify:")
	datas := make([][]byte, 0)
	for _, id := range list {
		datas = append(datas, tree[0][id])
	}
	result := verifyBatchProof(list, datas, roots, proof, height)
	fmt.Println(result)

}
