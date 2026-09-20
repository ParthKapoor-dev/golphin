package avl

// rotateRight      the left child moves up   (fixes a left-heavy node)
// rotateLeft       the right child moves up  (fixes a right-heavy node)
// rotateLeftRight  left child is right-heavy (LR case)
// rotateRightLeft  right child is left-heavy (RL case)
func (n *node[K, V]) rotate() *node[K, V] {

	bf := n.bf()

	if bf < -1 {
		// left-heavy: the left child moves up
		if n.left.bf() == 1 {
			// left child leans right -> LR case
			return n.rotateLeftRight()
		}
		return n.rotateRight()

	} else if bf > 1 {
		// right-heavy: the right child moves up
		if n.right.bf() == -1 {
			// right child leans left -> RL case
			return n.rotateRightLeft()
		}
		return n.rotateLeft()
	}

	return n
}

// rotateLeft turns the subtree left: n's right child becomes the new root.
func (n *node[K, V]) rotateLeft() *node[K, V] {

	pivot := n.right

	n.right = pivot.left
	pivot.left = n

	n.updateHt()
	pivot.updateHt()

	return pivot
}

// rotateRight turns the subtree right: n's left child becomes the new root.
func (n *node[K, V]) rotateRight() *node[K, V] {

	pivot := n.left

	n.left = pivot.right
	pivot.right = n

	n.updateHt()
	pivot.updateHt()

	return pivot
}

// rotateRightLeft handles the RL case: n is right-heavy but its right child
// leans left, so a single rotateLeft would not fix the imbalance. The right
// child's left child is lifted all the way to the top.
func (n *node[K, V]) rotateRightLeft() *node[K, V] {

	pivot := n.right.left

	newLeft := n
	newRight := n.right

	newLeft.right = pivot.left
	newRight.left = pivot.right

	pivot.left = newLeft
	pivot.right = newRight

	newLeft.updateHt()
	newRight.updateHt()
	pivot.updateHt()

	return pivot
}

// rotateLeftRight handles the LR case: n is left-heavy but its left child
// leans right, so a single rotateRight would not fix the imbalance. The left
// child's right child is lifted all the way to the top.
func (n *node[K, V]) rotateLeftRight() *node[K, V] {

	pivot := n.left.right

	newLeft := n.left
	newRight := n

	newLeft.right = pivot.left
	newRight.left = pivot.right

	pivot.left = newLeft
	pivot.right = newRight

	newLeft.updateHt()
	newRight.updateHt()
	pivot.updateHt()

	return pivot
}
