package treap

import (
	"fmt"
	"math/rand"
)

type node struct {
	value int
	prior int
	size  int
	left  *node
	right *node
}

type Treap struct {
	root *node
	mp   map[int]int
}

func newNode(val int) *node {
	return &node{
		value: val,
		prior: rand.Int(),
		size:  1,
		left:  nil,
		right: nil,
	}
}

func NewTreap() *Treap {
	return &Treap{
		root: nil,
		mp:   make(map[int]int),
	}
}

func getSize(n *node) int {
	if n != nil {
		return n.size
	}
	return 0
}

func (trp Treap) incVal(val int) {
	if _, ok := trp.mp[val]; !ok {
		trp.mp[val] = 1
	} else {
		trp.mp[val] += 1
	}
}

func (trp Treap) decVal(val int) {
	if cnt := trp.mp[val]; cnt != 1 {
		trp.mp[val] -= 1
	} else {
		delete(trp.mp, val)
	}
}

func update(n *node) {
	if n != nil {
		n.size = getSize(n.left) + 1 + getSize(n.right)
	}
}

func merge(a, b *node) *node {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.prior > b.prior {
		a.right = merge(a.right, b)
		update(a)
		return a
	} else {
		b.left = merge(a, b.left)
		update(b)
		return b
	}
}

func split(n *node, k int) (*node, *node) {
	if n == nil {
		return nil, nil
	}
	if getSize(n.left) < k {
		a, b := split(n.right, k-getSize(n.left)-1)
		n.right = a
		update(n)
		return n, b
	} else {
		a, b := split(n.left, k)
		n.left = b
		update(n)
		return a, n
	}

}

func (trp *Treap) PushBack(val int) {
	trp.root = merge(trp.root, newNode(val))
	trp.incVal(val)
}

func (trp *Treap) PushFront(val int) {
	trp.root = merge(newNode(val), trp.root)
	trp.incVal(val)
}

func (trp *Treap) PushBackToSet(val int) {
	if _, ok := trp.mp[val]; !ok {
		trp.PushBack(val)
		trp.incVal(val)
	}
}

func (trp *Treap) Get(index int) (int, bool) {
	if index < 0 || index > trp.GetSize()-1 {
		return -1, false
	}
	var less, equal, greater *node
	less, greater = split(trp.root, index)
	equal, greater = split(greater, 1)
	res := equal.value
	trp.root = merge(merge(less, equal), greater)
	return res, true
}

func (trp *Treap) Set(index, new_val int) bool {
	if index < 0 || index > trp.GetSize()-1 {
		return false
	}
	var less, equal, greater *node
	less, greater = split(trp.root, index)
	equal, greater = split(greater, 1)
	prev_val := equal.value
	equal.value = new_val
	trp.decVal(prev_val)
	trp.incVal(new_val)
	trp.root = merge(merge(less, equal), greater)
	return true
}

func (trp *Treap) PopFront() int {
	if trp.GetSize() < 1 {
		return -1
	}
	var less, equal, greater *node
	less, greater = split(trp.root, 0)
	equal, greater = split(greater, 1)
	res := equal.value
	trp.decVal(res)
	trp.root = merge(less, greater)
	return res
}

func (trp *Treap) PopBack() int {
	if trp.GetSize() < 1 {
		return -1
	}
	var less, equal, greater *node
	less, greater = split(trp.root, getSize(trp.root)-1)
	equal, greater = split(greater, 1)
	res := equal.value
	trp.decVal(res)
	trp.root = merge(less, greater)
	return res
}

func (trp *Treap) EraseSection(l, r int) []int {
	var less, equal, greater *node
	less, greater = split(trp.root, l)
	equal, greater = split(greater, r-l+1)
	nodes := make([]int, 0)
	trp.traversalDelete(equal, &nodes)
	trp.root = merge(less, greater)
	return nodes
}

func (trp *Treap) traversalDelete(n *node, nodes *[]int) {
	if n != nil {
		trp.traversalDelete(n.left, nodes)
		res := n.value
		trp.decVal(res)
		*nodes = append(*nodes, res)
		trp.traversalDelete(n.right, nodes)
	}
}

func (trp *Treap) GetSize() int {
	return getSize(trp.root)
}

func print(n *node) {
	if n != nil {
		print(n.left)
		fmt.Println(n.value)
		print(n.right)
	}
}

func (trp *Treap) Print() {
	print(trp.root)
}

func (trp *Treap) ValidateSlice(index int) int {
	if index < 0 {
		index = ((index % trp.GetSize()) + trp.GetSize()) % trp.GetSize()
	}
	if index+1 > trp.GetSize() {
		return trp.GetSize() - 1
	} else {
		return index
	}
}
