// code adapted from github.com/kwonalbert/pospace/
package vdf

import (
	"crypto/sha256"
	"math/big"
)

func makeParent(child1 []byte, child2 []byte) []byte {
	parent := sha256.Sum256(append(child1, child2...))
	return parent[:]
}

func makeParentLevel(hashes [][]byte) [][]byte {
	if len(hashes)%2 == 1 {
		hashes = append(hashes, []byte{})
	}
	parentLevel := make([][]byte, 0)
	for i := 0; i < len(hashes); i += 2 {
		parentLevel = append(parentLevel, makeParent(hashes[i], hashes[i+1]))
	}
	return parentLevel
}

func makeRoot(hashes [][]byte) []byte {
	currentLevel := hashes
	for len(currentLevel) != 1 {
		currentLevel = makeParentLevel(currentLevel)
	}
	return currentLevel[0]
}

func makeTreeFromL(L []*big.Int, omit int) (tree [][][]byte, roots [][]byte) {
	Lhashes := make([][]byte, 0)
	for _, v := range L {
		hash := sha256.Sum256(v.Bytes())
		Lhashes = append(Lhashes, hash[:])
	}
	return makeTree(Lhashes, omit)
}

func makeTree(hashes [][]byte, omit int) (tree [][][]byte, roots [][]byte) {
	currentLevel := hashes
	tree = make([][][]byte, 0)
	for len(currentLevel) > 1<<uint(omit) {
		tree = append(tree, currentLevel)
		currentLevel = makeParentLevel(currentLevel)
	}
	roots = currentLevel
	return
}

func fullMerkleHeight(n int) int {
	var r int = 0
	for x := n; x > 1; x >>= 1 {
		r++
	}
	if n > 1<<uint(r) {
		r++
	}
	return r
}

func computeAndStoreTree(hashes [][]byte, file string) {
}

func getProof(id int, tree [][][]byte) (proof [][]byte) {
	currentIndex := id
	var siblingIndex int
	for _, level := range tree {
		// fmt.Println("level size:", len(level))
		// fmt.Println("currentIndex:", currentIndex)
		if currentIndex%2 == 0 {
			siblingIndex = currentIndex + 1
		} else {
			siblingIndex = currentIndex - 1
		}
		// fmt.Println("sibling:", siblingIndex)
		var sibling []byte
		if siblingIndex == len(level) {
			sibling = []byte{}
		} else {
			sibling = level[siblingIndex]
		}
		proof = append(proof, sibling)
		currentIndex = currentIndex / 2
	}
	return
}

func verifyProof(data []byte, root []byte, proof [][]byte, id int) bool {
	hash := sha256.Sum256(data)
	leaf := hash[:]
	currentHash := leaf
	currentIndex := id
	for _, p := range proof {
		if currentIndex%2 == 1 {
			currentHash = makeParent(p, currentHash)
		} else {
			currentHash = makeParent(currentHash, p)
		}
		currentIndex = currentIndex / 2
	}
	for i := 0; i < len(root); i++ {
		if root[i] != currentHash[i] {
			return false
		}
	}
	return true
}
