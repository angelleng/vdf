package vdf

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"runtime"
	"sync"
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

func makeParentLevelParallel(hashes [][]byte) [][]byte {
	if len(hashes)%2 == 1 {
		hashes = append(hashes, []byte{})
	}
	parentLevel := make([][]byte, len(hashes)/2)

	workChan := make(chan int, runtime.NumCPU()*50)
	var wg sync.WaitGroup
	wg.Add(len(hashes) / 2)
	go func() {
		for i := 0; i < len(hashes); i += 2 {
			// fmt.Println("task", i)
			workChan <- i
		}
		close(workChan)
	}()

	for worker := 0; worker < runtime.NumCPU(); worker++ {
		go func() {
			for i := range workChan {
				// fmt.Println(i)
				parentLevel[i/2] = makeParent(hashes[i], hashes[i+1])
				wg.Done()
			}
		}()
	}
	wg.Wait()
	return parentLevel
}

func makeTreeParallel(hashes [][]byte, omit int) (tree [][][]byte, roots [][]byte) {
	currentLevel := hashes
	tree = make([][][]byte, 0)
	for len(currentLevel) > 1<<uint(omit) {
		tree = append(tree, currentLevel)
		currentLevel = makeParentLevelParallel(currentLevel)
	}
	roots = currentLevel
	return
}

func makeTreeFromLParallel(L []*big.Int, omit int) (tree [][][]byte, roots [][]byte) {
	Lhashes := make([][]byte, len(L))
	workChan := make(chan int, runtime.NumCPU()*50)
	var wg sync.WaitGroup
	wg.Add(len(L))
	go func() {
		for i := 0; i < len(L); i++ {
			// fmt.Println("task", i)
			workChan <- i
		}
		close(workChan)
	}()

	for worker := 0; worker < runtime.NumCPU(); worker++ {
		go func() {
			for i := range workChan {
				hash := sha256.Sum256(L[i].Bytes())
				Lhashes[i] = hash[:]
				wg.Done()
			}
		}()
	}
	wg.Wait()
	return makeTreeParallel(Lhashes, omit)
}

func makeRoot(hashes [][]byte) []byte {
	currentLevel := hashes
	for len(currentLevel) != 1 {
		currentLevel = makeParentLevel(currentLevel)
	}
	return currentLevel[0]
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

func makeTreeFromL(L []*big.Int, omit int) (tree [][][]byte, roots [][]byte) {
	Lhashes := make([][]byte, 0)
	for _, v := range L {
		hash := sha256.Sum256(v.Bytes())
		Lhashes = append(Lhashes, hash[:])
	}
	return makeTree(Lhashes, omit)
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

// incomplete. might not need depending on memory of prover
func makeTreeOnDiskFromL(L []*big.Int, omit int, file string) {
	currentLevel := make([][]byte, 0)
	for _, v := range L {
		hash := sha256.Sum256(v.Bytes())
		currentLevel = append(currentLevel, hash[:])
	}
	for len(currentLevel) > 1<<uint(omit) {
		currentLevel = makeParentLevel(currentLevel)
	}
}

func getBatchProofFromDisk(ids []int, treefile string, t int) (proof [][]byte) {
	return
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
				var sibling []byte
				if siblingIndex == len(tree[i]) {
					sibling = []byte{}
				} else {
					sibling = tree[i][siblingIndex]
				}
				proof = append(proof, sibling)
			}
		}
	}
	return
}

func verifyBatchProof(ids []int, datas []*big.Int, roots [][]byte, proof [][]byte, height int) bool {
	currentLevelValues := make(map[int][]byte)
	currentLevelInds := make([]int, 0)
	front := 0
	for i, id := range ids {
		data := datas[i].Bytes()
		hash := sha256.Sum256(data)
		currentLevelValues[id] = hash[:]
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

func getBatchProofOld(ids []int, tree [][][]byte) (proof [][]byte) {
	proof = make([][]byte, 0)
	for _, id := range ids {
		p := getProof(id, tree)
		proof = append(proof, p...)
	}
	return
}

func verifyBatchProofOld(ids []int, datas []*big.Int, roots [][]byte, proof [][]byte, height int) bool {
	perTree := 1 << uint(height)
	for i, id := range ids {
		rootId := id / perTree
		root := roots[rootId]
		if !verifyProof(datas[i].Bytes(), root, proof[i*height:(i+1)*height], id) {
			return false
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
