package bst

import (
	"cmp"
	"fmt"
)

type node[K cmp.Ordered, V any] struct {
	key   K
	info  V
	left  *node[K, V]
	right *node[K, V]
}

func newNode[K cmp.Ordered, V any](key K, info V) *node[K, V] {
	return &node[K, V]{key, info, nil, nil}
}

func (n *node[K, V]) dfs(key K) *node[K, V] {

	if n.key == key {
		return n
	}

	if key < n.key && n.left != nil {
		return n.left.dfs(key)
	} else if key > n.key && n.right != nil {
		return n.right.dfs(key)
	}

	return nil
}

func (n *node[K, V]) dfsNearest(key K) *node[K, V] {

	if n.key == key {
		return n
	}

	if key < n.key && n.left != nil {
		return n.left.dfsNearest(key)
	}
	if key > n.key && n.right != nil {
		return n.right.dfsNearest(key)
	}

	return n
}

func (n *node[K, V]) add(new *node[K, V]) error {

	if new == nil {
		return fmt.Errorf("null new node")
	}
	if new.key == n.key {
		return fmt.Errorf("same key for new and old node")
	}

	if new.key < n.key && n.left == nil {
		n.left = new
		return nil
	}

	if new.key > n.key && n.right == nil {
		n.right = new
		return nil
	}

	return fmt.Errorf("cannot add to this node")
}

func (n *node[K, V]) delete(prev *node[K, V], target K) error {

	if target == n.key {

		if n.left == nil {
			if prev.left == n {
				prev.left = n.right
			} else {
				prev.right = n.right
			}
			// delete the pointer

		} else if n.right == nil {
			if prev.left == n {
				prev.left = n.left
			} else {
				prev.right = n.left
			}

		} else {

			// find smallest in the right of the chain
			tmp := n.right
			tmpPrev := n
			for tmp.left != nil {
				tmpPrev = tmp
				tmp = tmp.left
			}

			// replace n to succ's key, value
			n.key = tmp.key
			n.info = tmp.info
			if tmpPrev == n {
				tmpPrev.right = tmp.right
			} else {
				tmpPrev.left = tmp.right
			}

			tmp = nil
		}

		return nil
	}

	if target < n.key && n.left != nil {
		return n.left.delete(n, target)
	}

	if target > n.key && n.right != nil {
		return n.right.delete(n, target)
	}

	return fmt.Errorf("target doesn't exists")
}

func (n *node[K, V]) findBetween(mini K, maxi K, results []V) []V {
	if n == nil {
		return results
	}

	if n.key > mini {
		results = n.left.findBetween(mini, maxi, results)
	}

	if n.key > mini && n.key < maxi {
		results = append(results, n.info)
	}

	if n.key < maxi {
		results = n.right.findBetween(mini, maxi, results)
	}

	return results
}

func (n *node[K, V]) iter(results []V) []V {
	if n == nil {
		return results
	}

	results = n.left.iter(results)
	results = append(results, n.info)
	results = n.right.iter(results)

	return results
}
