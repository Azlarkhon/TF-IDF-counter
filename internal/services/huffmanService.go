package services

import (
	"container/heap"
	"errors"
	"strings"
)

type Node struct {
	Symbol byte
	Freq   int
	Left   *Node
	Right  *Node
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int {
	return len(pq)
}

// Smaller frequency has higher priority
func (pq PriorityQueue) Less(a, b int) bool {
	return pq[a].Freq < pq[b].Freq
}

func (pq PriorityQueue) Swap(a, b int) {
	pq[a], pq[b] = pq[b], pq[a]
}

func (pq *PriorityQueue) Push(x any) {
	*pq = append(*pq, x.(*Node))
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// Time complexity: O(n)
// Space complexity: O(m)
func buildFreq(content []byte) map[byte]int {
	freq := make(map[byte]int)
	for _, b := range content {
		freq[b]++
	}
	return freq
}

// Time complexity: O(m log(m))
// Space complexity: O(m)
func buildHuffmanTree(freqMap map[byte]int) (*Node, error) {
	if len(freqMap) == 0 {
		return nil, errors.New("cannot build Huffman tree from empty data")
	}

	pq := make(PriorityQueue, 0, len(freqMap))
	heap.Init(&pq)
	for b, f := range freqMap {
		heap.Push(&pq, &Node{
			Symbol: b,
			Freq:   f,
		})
	}

	for pq.Len() > 1 {
		n1 := heap.Pop(&pq).(*Node)
		n2 := heap.Pop(&pq).(*Node)

		merged := &Node{
			Freq:  n1.Freq + n2.Freq,
			Left:  n1,
			Right: n2,
		}
		heap.Push(&pq, merged)
	}

	root := heap.Pop(&pq).(*Node)
	return root, nil
}

// Time complexity: O(m)
// Space complexity: O(m + L)
func generateCodes(node *Node, prefix string, codeMap map[byte]string) {
	if node == nil {
		return
	}

	if node.Left == nil && node.Right == nil {
		codeMap[node.Symbol] = prefix
		return
	}

	generateCodes(node.Left, prefix+"0", codeMap)
	generateCodes(node.Right, prefix+"1", codeMap)
}

// Time complexity: O(n + m log(m))
// Space complexity: O(n + m)
func HuffmanEncoding(content []byte) (string, *Node, error) {
	freqMap := buildFreq(content)

	root, err := buildHuffmanTree(freqMap)
	if err != nil {
		return "", nil, err
	}

	codeMap := make(map[byte]string)
	generateCodes(root, "", codeMap)

	var encodedBuilder strings.Builder
	for _, b := range content {
		encodedBuilder.WriteString(codeMap[b])
	}

	return encodedBuilder.String(), root, nil
}

// Time complexity: O(k)
// Space complexity: O(k + m)
func HuffmanDecoding(bits string, root *Node) ([]byte, error) {
	var out []byte
	node := root
	for _, bit := range bits {
		if node == nil {
			return nil, errors.New("invalid encoding")
		}
		if bit == '0' {
			node = node.Left
		} else {
			node = node.Right
		}

		if node.Left == nil && node.Right == nil {
			out = append(out, node.Symbol)
			node = root
		}
	}
	return out, nil
}
