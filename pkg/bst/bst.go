package bst

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

	n := bst.root.dfs(key)
	if n == nil {
		return false, res, nil
	}

	return true, n.info, nil
}

func (bst *BST[K, V]) Upsert(key K, value V) error {

	new := newNode(key, value)

	if bst.root == nil {
		bst.root = new
		return nil
	}

	nearest := bst.root.dfsNearest(new.key)
	if nearest == nil {
		return fmt.Errorf("nearest came back null")
	}

	// update path
	if nearest.key == key {
		nearest.info = value
		return nil
	}

	// insertion path
	if err := nearest.add(new); err != nil {
		return err
	}

	bst.Size++

	return nil
}

func (bst *BST[K, V]) Delete(key K) error {

	if bst.root.key == key {
		if bst.root.left == nil {
			bst.root = bst.root.right

			bst.Size--
			return nil
		}
		if bst.root.right == nil {
			bst.root = bst.root.left

			bst.Size--
			return nil
		}
	}

	if err := bst.root.delete(nil, key); err != nil {
		return err
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
