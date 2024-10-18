package storage

import (
	"errors"
	"fmt"
	"math/rand"
)

type node struct {
	value value
	prior int
	size  int
	left  *node
	right *node
}

type Treap struct {
	root *node
	mp   map[value]int
}

func newNode(val any) (*node, error) {
	new_val, err := newValue(val)
	if err != nil {
		return nil, err
	}
	return &node{
		value: new_val,
		prior: rand.Int(),
		size:  1,
		left:  nil,
		right: nil,
	}, nil
}

func NewTreap() *Treap {
	return &Treap{
		root: nil,
		mp:   make(map[value]int),
	}
}

func getSize(n *node) int {
	if n != nil {
		return n.size
	}
	return 0
}

func (trp Treap) incVal(val value) {
	if _, ok := trp.mp[val]; !ok {
		trp.mp[val] = 1
	} else {
		trp.mp[val] += 1
	}
}

func (trp Treap) decVal(val value) {
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

func (trp *Treap) PushBack(val any) {
	new_node, err := newNode(val)
	if err != nil {
		return
	}
	trp.root = merge(trp.root, new_node)
	trp.incVal(new_node.value)
}

func (trp *Treap) PushFront(val any) {
	new_node, err := newNode(val)
	if err != nil {
		return
	}
	trp.root = merge(new_node, trp.root)
	trp.incVal(new_node.value)
}

func (trp *Treap) PushBackToSet(val any) {
	new_node, err := newNode(val)
	if err != nil {
		return
	}
	if _, ok := trp.mp[new_node.value]; !ok {
		trp.PushBack(val)
		trp.incVal(new_node.value)
	}
}

func (trp *Treap) Get(index int) (any, bool) {
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

func (trp *Treap) Set(index int, valToAdd any) bool {
	if index < 0 || index > trp.GetSize()-1 {
		return false
	}
	new_val, err := newValue(valToAdd)
	if err != nil {
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

func (trp *Treap) PopFront() any {
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

func (trp *Treap) PopBack() any {
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

func (trp *Treap) EraseSection(l, r int) []any {
	var less, equal, greater *node
	less, greater = split(trp.root, l)
	equal, greater = split(greater, r-l+1)
	nodes := make([]any, 0)
	trp.traversalDelete(equal, &nodes)
	trp.root = merge(less, greater)
	return nodes
}

func (trp *Treap) traversalDelete(n *node, nodes *[]any) {
	if n != nil {
		trp.traversalDelete(n.left, nodes)
		res := n.value
		trp.decVal(res)
		*nodes = append(*nodes, res.Val)
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

func (trp *Treap) validateIndex(index int) int {
	if index < 0 {
		index = ((index % trp.GetSize()) + trp.GetSize()) % trp.GetSize()
	}
	if index+1 > trp.GetSize() {
		return trp.GetSize() - 1
	} else {
		return index
	}
}

func (trp *Treap) ValidateEraseSlice(indexes []any, isfromleft bool) (int, int, error) {
	switch len(indexes) {
	case 2:
		rt := trp.validateIndex(int(indexes[0].(float64)))
		lf := trp.validateIndex(int(indexes[1].(float64)))
		if rt > lf {
			return 0, 0, errors.New("IndexOutOfRange")
		}
		return rt, lf, nil
	case 1:
		cnt := int(indexes[0].(float64))
		if cnt <= 0 {
			return 0, 0, errors.New("IndexOutOfRange")
		}
		lf := trp.validateIndex(cnt - 1)
		if isfromleft {
			return 0, lf, nil
		}
		trpSize := trp.GetSize()
		return trpSize - lf - 1, trpSize, nil
	case 0:
		if isfromleft {
			return 0, 0, nil
		}
		trpSize := trp.GetSize()
		return trpSize - 1, trpSize, nil
	}

	return 0, 0, errors.New("IndexOutOfRange")
}

func traversal(n *node, nodes *[]value) {
	if n != nil {
		traversal(n.left, nodes)
		res := n.value
		*nodes = append(*nodes, res)
		traversal(n.right, nodes)
	}
}

func (trp *Treap) GetAllValues() []value {
	res := make([]value, 0)
	traversal(trp.root, &res)
	return res
}
