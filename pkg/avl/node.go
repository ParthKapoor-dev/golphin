package avl

import (
	"cmp"
)

type node[K cmp.Ordered, V any] struct {
	key    K
	value  V
	height int
	left   *node[K, V]
	right  *node[K, V]
}

func newNode[K cmp.Ordered, V any](key K, info V) *node[K, V] {
	return &node[K, V]{key, info, 1, nil, nil}
}

func (n *node[K, V]) bf() int {
	ans := 0

	if n.right != nil {
		ans += n.right.height
	}

	if n.left != nil {
		ans -= n.left.height
	}

	return ans
}

func (n *node[K, V]) updateHt() {
	ans := 0

	if n.right != nil {
		ans = max(ans, n.right.height)
	}

	if n.left != nil {
		ans = max(ans, n.left.height)
	}

	n.height = ans + 1
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

func (n *node[K, V]) upsert(key K, value V) (*node[K, V], bool) {

	if n.key == key {
		n.value = value
		return n, false
	}

	isInsert := true

	if n.key > key {
		if n.left == nil {
			n.left = newNode(key, value)
		} else {
			n.left, isInsert = n.left.upsert(key, value)
		}
	}

	if n.key < key {
		if n.right == nil {
			n.right = newNode(key, value)
		} else {
			n.right, isInsert = n.right.upsert(key, value)
		}
	}

	n.updateHt()
	return n.rotate(), isInsert
}

func (n *node[K, V]) delete(key K) (*node[K, V], bool) {

	isDeleted := false

	if n.key == key {

		if n.left == nil {
			return n.right, true
		}

		if n.right == nil {
			return n.left, true
		}

		tmp := n.right

		for tmp.left != nil {
			tmp = tmp.left
		}

		n.key = tmp.key
		n.value = tmp.value

		n.right, isDeleted = n.right.delete(tmp.key)

	} else if n.key > key && n.left != nil {
		n.left, isDeleted = n.left.delete(key)

	} else if n.key < key && n.right != nil {
		n.right, isDeleted = n.right.delete(key)
	}

	n.updateHt()
	return n.rotate(), isDeleted
}

func (n *node[K, V]) findBetween(mini K, maxi K, results []V) []V {
	if n == nil {
		return results
	}

	if n.key > mini {
		results = n.left.findBetween(mini, maxi, results)
	}

	if n.key > mini && n.key < maxi {
		results = append(results, n.value)
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
	results = append(results, n.value)
	results = n.right.iter(results)

	return results
}
