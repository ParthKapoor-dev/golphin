package bst

import "fmt"

type BST struct {
	root *node
	Size int
}

func NewBst() *BST {
	return &BST{nil, 0}
}

func (bst *BST) Find(key string) (bool, string, error) {

	if bst.root == nil {
		return false, "", nil
	}

	n := bst.root.dfs(key)
	if n == nil {
		return false, "", nil
	}

	return true, n.info, nil
}

func (bst *BST) Upsert(key string, value string) error {

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

func (bst *BST) Delete(key string) error {

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

func (bst *BST) FindBetween(leftKey string, rightKey string) ([]string, error) {
	if leftKey >= rightKey {
		return nil, fmt.Errorf("invalid leftKey/rightKey")
	}

	if bst.root == nil {
		return nil, nil
	}

	results := make([]string, 0)

	results = bst.root.findBetween(leftKey, rightKey, results)

	return results, nil
}

func (bst *BST) Print(n *node) {

	if n == nil {
		if bst.root == nil {
			return
		}
		n = bst.root
	}

	fmt.Print(n.key + " -->")

	if n.left != nil {
		fmt.Print("  left: " + n.left.key)
	}

	if n.right != nil {
		fmt.Print("  right: " + n.right.key)
	}

	fmt.Print("\n")
}
