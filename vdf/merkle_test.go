package vdf

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"testing"
	"tictoc"
	"time"
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
	getBatchProof(list, tree)
}

func TestMerkleBatchVerify(t *testing.T) {
	tic := tictoc.NewTic()
	length := 1024 * 1024 * 4 * 8
	omit := 8
	num := 80
	fmt.Println("length =", length)
	fmt.Println("omit height:", omit)
	fmt.Println("num:", num)

	L := computeL(length)
	tic.Toc("compute L time:")

	tree, roots := makeTreeFromL(L, omit)
	tic.Toc("make tree time:")

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	list := make([]int, 0)
	for i := 0; i < num; {
		newInd := r1.Intn(length)
		skip := false
		for _, exist := range list {
			if exist == newInd {
				skip = true
			}
		}
		if skip {
			continue
		}
		list = append(list, newInd)
		i++
	}
	fmt.Println(list)
	tic.Tic()
	proof := getBatchProof(list, tree)
	tic.Toc("get proof time:")
	height := len(tree)
	fmt.Printf("\nVerify:\n")
	datas := make([][]byte, 0)
	for _, id := range list {
		datas = append(datas, tree[0][id])
	}

	tic.Tic()
	result := verifyBatchProof(list, datas, roots, proof, height)
	tic.Toc("verify time:")
	fmt.Println("old method proof size:", len(list)*height)
	fmt.Println("new method proof size:", len(proof))
	fmt.Println(result)
	if !result {
		t.Error("should verify true")
	}

}
