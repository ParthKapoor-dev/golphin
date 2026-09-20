package avl

import (
	"cmp"
	"fmt"
)

type BST[K cmp.Ordered, V any] struct {
	root *node[K, V]
	Size int
}

func NewBst[K cmp.Ordered, V any]() *BST[K, V] {
	return &BST[K, V]{nil, 0}
}

func (bst *BST[K, V]) Find(key K) (bool, V, error) {

	var res V

	if bst.root == nil {
		return false, res, nil
	}

	node := bst.root.dfs(key)
	if node == nil {
		return false, res, nil
	}

	return true, node.value, nil
}

func (bst *BST[K, V]) Upsert(key K, value V) error {

	isInsert := true

	if bst.root == nil {
		bst.root = newNode(key, value)
	} else {
		bst.root, isInsert = bst.root.upsert(key, value)
	}

	if isInsert {
		bst.Size++
	}

	return nil
}

func (bst *BST[K, V]) Delete(key K) error {
	if bst.root == nil {
		return fmt.Errorf("no root exists")
	}

	isDeleted := false

	bst.root, isDeleted = bst.root.delete(key)

	if isDeleted == false {
		return fmt.Errorf("invalid key")
	}

	bst.Size--

	return nil
}

func (bst *BST[K, V]) FindBetween(leftKey K, rightKey K) ([]V, error) {
	if leftKey >= rightKey {
		return nil, fmt.Errorf("invalid leftKey/rightKey")
	}

	results := make([]V, 0)

	if bst.root == nil {
		return results, nil
	}

	return bst.root.findBetween(leftKey, rightKey, results), nil
}

func (bst *BST[K, V]) Iter() ([]V, error) {

	results := make([]V, 0)

	if bst.root == nil {
		return results, nil
	}

	return bst.root.iter(results), nil
}
