// The pattern to copy elsewhere:
//
//  1. Default to external tests. Only go internal when an invariant is
//     genuinely invisible from outside.
//  2. Write the invariant checker once, as a function that walks the whole
//     structure and reports *every* violation through t.Errorf (not
//     t.Fatalf). One failing run should tell you everything that's wrong.
//  3. Call that one checker after every interesting operation, rather than
//     writing bespoke assertions per test.
//  4. Assert the theoretical bound, not a magic number you measured once.
package avl

import (
	"cmp"
	"fmt"
	"math"
	"math/rand"
	"testing"
)

// ======================================================
// INVARIANT CHECKERS
// ======================================================

// checkNode walks the subtree rooted at n and verifies the three structural
// invariants an AVL tree must satisfy at every node:
func checkNode[K cmp.Ordered, V any](t *testing.T, n *node[K, V], lo, hi *K, path string) (height, count int) {
	t.Helper()

	if n == nil {
		return 0, 0
	}

	// 1. search-tree ordering
	if lo != nil && n.key <= *lo {
		t.Errorf("%s: key %v is not greater than lower bound %v", path, n.key, *lo)
	}
	if hi != nil && n.key >= *hi {
		t.Errorf("%s: key %v is not less than upper bound %v", path, n.key, *hi)
	}

	leftHt, leftCount := checkNode(t, n.left, lo, &n.key, path+".L")
	rightHt, rightCount := checkNode(t, n.right, &n.key, hi, path+".R")

	// 2. the stored height must match reality
	wantHt := 1 + max(leftHt, rightHt)
	if n.height != wantHt {
		t.Errorf("%s (key %v): height = %d, want %d (left=%d right=%d)",
			path, n.key, n.height, wantHt, leftHt, rightHt)
	}

	// 3. the AVL balance condition
	if bf := rightHt - leftHt; bf < -1 || bf > 1 {
		t.Errorf("%s (key %v): balance factor = %d, want within [-1,1] (left=%d right=%d)",
			path, n.key, bf, leftHt, rightHt)
	}

	return wantHt, leftCount + rightCount + 1
}

// maxAVLHeight is the proven upper bound on the height of an AVL tree holding
// n keys: h <= 1.4405 * log2(n+2) - 0.3277. Asserting the bound rather than a
// number you measured once means the test stays meaningful as n changes.
func maxAVLHeight(n int) float64 {
	return 1.4405*math.Log2(float64(n)+2) - 0.3277
}

// checkInvariants is the single entry point the tests call. It validates the
// whole tree and cross-checks Size and height against what was actually found.
func checkInvariants[K cmp.Ordered, V any](t *testing.T, tr *BST[K, V], wantSize int) {
	t.Helper()

	height, count := checkNode(t, tr.root, nil, nil, "root")

	if count != wantSize {
		t.Errorf("tree holds %d nodes, want %d", count, wantSize)
	}
	if tr.Len() != wantSize {
		t.Errorf("Size = %d, but tree holds %d nodes", tr.Len(), wantSize)
	}

	if wantSize > 0 {
		if limit := maxAVLHeight(wantSize); float64(height) > limit {
			t.Errorf("height = %d for %d keys, exceeds AVL bound of %.2f",
				height, wantSize, limit)
		}
	}
}

// keyOf renders a node's key for assertions, or "<nil>" for a missing child.
func keyOf[K cmp.Ordered, V any](n *node[K, V]) string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprint(n.key)
}

// ======================================================
// INVARIANTS UNDER EVERY INSERTION ORDER
// ======================================================

// Sequential and reverse-sequential are the orders that degenerate an
// unbalanced BST into a linked list, so they are the ones worth pinning.
func TestInvariantsHoldAfterEveryInsert(t *testing.T) {
	const n = 300

	orders := map[string]func(i int) int{
		"ascending":  func(i int) int { return i },
		"descending": func(i int) int { return n - i },
		"alternating": func(i int) int {
			if i%2 == 0 {
				return i
			}
			return n - i
		},
	}

	for name, keyAt := range orders {
		t.Run(name, func(t *testing.T) {
			tr := NewBst[int, string]()
			for i := range n {
				key := keyAt(i)
				if err := tr.Upsert(key, fmt.Sprint(key)); err != nil {
					t.Fatalf("Upsert(%d) error: %v", key, err)
				}
				// checked after *every* insert, so a failure names the
				// exact operation that broke the tree
				checkInvariants(t, tr, i+1)
				if t.Failed() {
					t.Fatalf("invariants broken after inserting key %d (op %d)", key, i)
				}
			}
		})
	}
}

func TestInvariantsHoldAfterEveryDelete(t *testing.T) {
	const n = 300

	tr := NewBst[int, string]()
	for i := range n {
		if err := tr.Upsert(i, fmt.Sprint(i)); err != nil {
			t.Fatalf("Upsert(%d) error: %v", i, err)
		}
	}
	checkInvariants(t, tr, n)

	// deleting in ascending order is the worst case for delete rebalancing:
	// it repeatedly strips the leftmost spine
	for i := range n {
		if err := tr.Delete(i); err != nil {
			t.Fatalf("Delete(%d) error: %v", i, err)
		}
		checkInvariants(t, tr, n-i-1)
		if t.Failed() {
			t.Fatalf("invariants broken after deleting key %d", i)
		}
	}

	if tr.root != nil {
		t.Errorf("root = %v after deleting everything, want nil", keyOf(tr.root))
	}
}

// ======================================================
// THE FOUR ROTATION CASES, EXPLICITLY
// ======================================================

// Each insertion order below triggers exactly one rotation case. All four
// converge on the same balanced shape (2 at the root, 1 and 3 as children),
// which makes the expected result easy to state.
func TestInsertRotationCases(t *testing.T) {
	cases := []struct {
		name   string
		insert []int
	}{
		{"LL - left child, left grandchild (single rotateRight)", []int{3, 2, 1}},
		{"RR - right child, right grandchild (single rotateLeft)", []int{1, 2, 3}},
		{"LR - left child, right grandchild (rotateLeftRight)", []int{3, 1, 2}},
		{"RL - right child, left grandchild (rotateRightLeft)", []int{1, 3, 2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := NewBst[int, int]()
			for _, k := range tc.insert {
				if err := tr.Upsert(k, k); err != nil {
					t.Fatalf("Upsert(%d) error: %v", k, err)
				}
			}

			checkInvariants(t, tr, 3)

			if got := keyOf(tr.root); got != "2" {
				t.Errorf("root = %s, want 2", got)
			}
			if got := keyOf(tr.root.left); got != "1" {
				t.Errorf("root.left = %s, want 1", got)
			}
			if got := keyOf(tr.root.right); got != "3" {
				t.Errorf("root.right = %s, want 3", got)
			}
			if tr.root.height != 2 {
				t.Errorf("root height = %d, want 2", tr.root.height)
			}
		})
	}
}

// Deletion can leave the rotating node's child with balance factor 0, which
// never happens during insertion. A single rotation is still correct there,
// but it is the case hand-written AVL implementations most often get wrong,
// so it gets its own test.
func TestDeleteRotationWithBalancedChild(t *testing.T) {
	tr := NewBst[int, int]()
	// builds:      2
	//             / \
	//            1   4
	//               / \
	//              3   5
	for _, k := range []int{2, 1, 4, 3, 5} {
		if err := tr.Upsert(k, k); err != nil {
			t.Fatalf("Upsert(%d) error: %v", k, err)
		}
	}
	checkInvariants(t, tr, 5)

	if got := keyOf(tr.root); got != "2" {
		t.Fatalf("setup: root = %s, want 2", got)
	}
	if bf := tr.root.right.bf(); bf != 0 {
		t.Fatalf("setup: node 4 balance factor = %d, want 0", bf)
	}

	// removing 1 leaves node 2 right-heavy, with a child whose bf is 0
	if err := tr.Delete(1); err != nil {
		t.Fatalf("Delete(1) error: %v", err)
	}

	checkInvariants(t, tr, 4)

	if got := keyOf(tr.root); got != "4" {
		t.Errorf("root = %s, want 4 after rebalance", got)
	}
}

// ======================================================
// HEIGHT STAYS LOGARITHMIC
// ======================================================

// The whole reason for choosing AVL over a plain BST. With sorted input an
// unbalanced tree reaches height n; this pins the real bound.
func TestHeightStaysLogarithmicOnSortedInput(t *testing.T) {
	for _, n := range []int{1_000, 10_000, 100_000} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			tr := NewBst[int, int]()
			for i := range n {
				if err := tr.Upsert(i, i); err != nil {
					t.Fatalf("Upsert(%d) error: %v", i, err)
				}
			}

			checkInvariants(t, tr, n)

			t.Logf("n=%d height=%d (AVL bound %.2f, perfect tree %.2f)",
				n, tr.root.height, maxAVLHeight(n), math.Log2(float64(n)+1))
		})
	}
}

// ======================================================
// SIZE ACCOUNTING
// ======================================================

func TestSizeCountsDistinctKeys(t *testing.T) {
	tr := NewBst[string, int]()

	// re-upserting a key must replace, not insert
	for i := range 5 {
		if err := tr.Upsert("same", i); err != nil {
			t.Fatalf("Upsert error: %v", err)
		}
	}
	checkInvariants(t, tr, 1)

	found, got, err := tr.Find("same")
	if err != nil {
		t.Fatalf("Find() error: %v", err)
	}
	if !found || got != 4 {
		t.Errorf("Find(same) = (%v, %d), want (true, 4)", found, got)
	}

	// deleting a key that isn't there must not move Size
	if err := tr.Delete("absent"); err == nil {
		t.Error("Delete(absent) returned nil error, want a failure")
	}
	checkInvariants(t, tr, 1)

	// and deleting the last key empties the tree
	if err := tr.Delete("same"); err != nil {
		t.Fatalf("Delete(same) error: %v", err)
	}
	checkInvariants(t, tr, 0)
}

// ======================================================
// RANDOM WORKLOAD AGAINST A MODEL
// ======================================================

// A plain Go map is an obviously-correct reference implementation ("model").
// Driving both with the same random operations and comparing is far stronger
// than any set of hand-written cases, because it explores shapes you would
// never think to write down.
//
// The seed is fixed so a failure is reproducible. If you want broader
// coverage, run with -count=N and vary the seed from the run index.
func TestInvariantsUnderRandomWorkload(t *testing.T) {
	const (
		ops     = 20_000
		keySpan = 400
		seed    = 1
	)

	tr := NewBst[int, int]()
	model := make(map[int]int)
	rng := rand.New(rand.NewSource(seed))

	for i := range ops {
		key := rng.Intn(keySpan)

		if rng.Intn(3) == 0 {
			// delete: the model tells us whether it should succeed
			_, existed := model[key]
			err := tr.Delete(key)

			if existed && err != nil {
				t.Fatalf("op %d: Delete(%d) failed on an existing key: %v", i, key, err)
			}
			if !existed && err == nil {
				t.Fatalf("op %d: Delete(%d) succeeded on a missing key", i, key)
			}
			delete(model, key)

		} else {
			value := rng.Int()
			if err := tr.Upsert(key, value); err != nil {
				t.Fatalf("op %d: Upsert(%d) error: %v", i, key, err)
			}
			model[key] = value
		}

		// full structural check periodically — every op would be O(n) each
		// time and make the test needlessly slow
		if i%500 == 0 {
			checkInvariants(t, tr, len(model))
			if t.Failed() {
				t.Fatalf("invariants broken at op %d (key %d)", i, key)
			}
		}
	}

	checkInvariants(t, tr, len(model))

	// every key the model knows about must read back identically
	for key, want := range model {
		found, got, err := tr.Find(key)
		if err != nil {
			t.Fatalf("Find(%d) error: %v", key, err)
		}
		if !found || got != want {
			t.Errorf("Find(%d) = (%v, %d), want (true, %d)", key, found, got, want)
		}
	}

	// and keys it doesn't know about must be absent
	for key := range keySpan {
		if _, inModel := model[key]; inModel {
			continue
		}
		if found, _, _ := tr.Find(key); found {
			t.Errorf("Find(%d) reported found, but the key was deleted", key)
		}
	}

	// in-order traversal must come back sorted and complete
	values, err := tr.Iter()
	if err != nil {
		t.Fatalf("Iter() error: %v", err)
	}
	if len(values) != len(model) {
		t.Errorf("Iter() returned %d values, want %d", len(values), len(model))
	}
	assertInOrderSorted(t, tr)
}

// assertInOrderSorted walks the tree directly and checks that keys come out
// ascending. Iter() only returns values, so the key ordering — the property
// that actually matters for range scans — can only be checked from inside.
func assertInOrderSorted[K cmp.Ordered, V any](t *testing.T, tr *BST[K, V]) {
	t.Helper()

	var prev *K
	var walk func(n *node[K, V])

	walk = func(n *node[K, V]) {
		if n == nil {
			return
		}
		walk(n.left)
		if prev != nil && n.key <= *prev {
			t.Errorf("in-order traversal out of order: %v came after %v", n.key, *prev)
		}
		key := n.key
		prev = &key
		walk(n.right)
	}

	walk(tr.root)
}
