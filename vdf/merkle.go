// code adapted from github.com/kwonalbert/pospace/
package vdf

import (
	"crypto/sha256"
	"fmt"
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

// might not need depending on memory of prover
func computeAndStoreTree(hashes [][]byte, file string) {
}

func merklePath(id int, total int) (path []int) {
	path = make([]int, 0)
	for i := total; i > 1; i >>= 1 {
		path = append(path, id)
		id >>= 1
	}
	return
}

func getBatchProof(ids []int, tree [][][]byte) (proof [][]byte) {
	height := len(tree)
	fmt.Println("tree height", height)
	for i := 0; i < height; i++ {
		availNodes := make([]int, 0)
		for j := 0; j < len(ids); j++ {
			newNode := ids[j] >> uint(i)
			add := true
			for _, node := range availNodes {
				if node == newNode {
					add = false
				}
			}
			if add {
				availNodes = append(availNodes, newNode)
			}
		}
		for _, node := range availNodes {
			var siblingIndex int
			if node%2 == 0 {
				siblingIndex = node + 1
			} else {
				siblingIndex = node - 1
			}
			add := true
			for _, node2 := range availNodes {
				if siblingIndex == node2 {
					add = false
				}
			}
			if add {
				proof = append(proof, tree[i][siblingIndex])
			}
		}
	}
	return
}

func verifyBatchProof(ids []int, datas [][]byte, roots [][]byte, proof [][]byte, height int) bool {
	currentLevelValues := make(map[int][]byte)
	currentLevelInds := make([]int, 0)
	front := 0
	for i, id := range ids {
		currentLevelValues[id] = datas[i]
		currentLevelInds = append(currentLevelInds, id)
	}

	for i := 0; i < height; i++ {
		siblings := make([]int, 0)
		for _, ind := range currentLevelInds {
			var siblingIndex int
			if ind%2 == 0 {
				siblingIndex = ind + 1
			} else {
				siblingIndex = ind - 1
			}
			add := true
			for _, node := range currentLevelInds {
				if siblingIndex == node {
					add = false
				}
			}
			if add {
				siblings = append(siblings, siblingIndex)
				currentLevelValues[siblingIndex] = proof[front]
				front++
			}
		}
		nextLevelValues := make(map[int][]byte)
		nextLevelInds := make([]int, 0)
		for _, node := range currentLevelInds {
			_, ok := nextLevelValues[node/2]
			if !ok {
				child1, ok1 := currentLevelValues[node/2*2]
				child2, ok2 := currentLevelValues[node/2*2+1]
				if ok1 && ok2 {
					nextLevelValues[node/2] = makeParent(child1, child2)
				} else {
					fmt.Println(ok1, ok2)
					panic("read map error")
				}
				nextLevelInds = append(nextLevelInds, node/2)
			}
		}
		currentLevelValues = nextLevelValues
		currentLevelInds = nextLevelInds
	}
	for _, ind := range currentLevelInds {
		for i, v := range roots[ind] {
			if v != currentLevelValues[ind][i] {
				return false
			}
		}
	}
	return true
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
