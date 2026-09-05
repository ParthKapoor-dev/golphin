package bst_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/parthkapoor-dev/golphin/pkg/bst"
)

type testingCtx interface {
	Helper()
	Fatalf(format string, args ...any)
	Cleanup(func())
	TempDir() string
}

func requireValue(t testingCtx, tr *bst.BST[string, string], key string, value string) {
	t.Helper()

	found, got, err := tr.Find(key)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if !found {
		t.Fatalf("Get() found = false, expected true")
	}
	if got != value {
		t.Fatalf("Get() value = %q, expected %q", got, value)
	}

}

func requireMissing(t testingCtx, tr *bst.BST[string, string], key string) {
	t.Helper()

	found, value, err := tr.Find(key)
	if err != nil {
		t.Fatalf("Get(%q) error: %v", key, err)
	}
	if found {
		t.Fatalf("Get(%q) = %q, want missing key", key, value)
	}
	if value != "" {
		t.Fatalf("Get(%q) value = %q, want empty value", key, value)
	}
}

func randomDataCreation(t testingCtx, tr *bst.BST[string, string], cache map[string]string, maxKey int, oprCnt int, deletionMultiple int) {

	for i := range oprCnt {
		key := strconv.Itoa(rand.Intn(maxKey))
		value := strconv.Itoa(rand.Intn(10 * maxKey))

		_, exists := cache[key]

		if i%deletionMultiple == 0 && exists {
			err := tr.Delete(key)
			if err != nil {
				t.Fatalf("Delete() error: %v", err)
			}
			delete(cache, key)
		} else {
			if err := tr.Upsert(key, value); err != nil {
				t.Fatalf("Upsert() error: %v", err)
			}
			cache[key] = value
		}
	}

}

func randomDataValidation(t testingCtx, tr *bst.BST[string, string], cache map[string]string, maxKey int) {

	for i := range maxKey {
		key := strconv.Itoa(i)
		value, ok := cache[key]
		if ok {
			requireValue(t, tr, key, value)
		} else {
			requireMissing(t, tr, key)
		}
	}

	if tr.Size > len(cache)+1 {
		t.Fatalf("compaction failed!")
	}

}

// ======================================================
// TESTS
// ======================================================

func TestInsertMakesValueRetrievable(t *testing.T) {

	tr := bst.NewBst[string, string]()

	if err := tr.Upsert("key-1", "value-1"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	if err := tr.Upsert("key-0", "value-0"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	if err := tr.Upsert("key-2", "value-2"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	requireValue(t, tr, "key-1", "value-1")
	requireValue(t, tr, "key-2", "value-2")
}

func TestRetrievableUpsertAfterGet(t *testing.T) {

	tr := bst.NewBst[string, string]()

	if err := tr.Upsert("lang", "go"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	requireValue(t, tr, "lang", "go")

	if err := tr.Upsert("claude", "code"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	requireValue(t, tr, "claude", "code")
}

func TestUpsertReplacesExistingValue(t *testing.T) {

	tr := bst.NewBst[string, string]()

	if err := tr.Upsert("lang", "typescript"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	if err := tr.Upsert("lang", "go"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	requireValue(t, tr, "lang", "go")
}

func TestDeleteMakesKeyUnavailable(t *testing.T) {

	tr := bst.NewBst[string, string]()

	if err := tr.Upsert("0", "root"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	if err := tr.Upsert("1", "one"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	if err := tr.Delete("1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	requireMissing(t, tr, "1")
}

func TestDeleteRootMakesKeyUnavailable(t *testing.T) {

	tr := bst.NewBst[string, string]()

	if err := tr.Upsert("0", "root"); err != nil {
		t.Fatalf("Upsert() error: %v", err)
	}

	if err := tr.Delete("0"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	requireMissing(t, tr, "0")
}

func TestRandomData(t *testing.T) {

	cache := make(map[string]string)
	maxKey := 10

	tr := bst.NewBst[string, string]()

	randomDataCreation(t, tr, cache, maxKey, 1000, 2)
	randomDataValidation(t, tr, cache, maxKey)
}

func TestGetBetweenKeys(t *testing.T) {

	tr := bst.NewBst[string, string]()

	for i := range 100 {
		key := fmt.Sprintf("key-%03d", i)
		value := fmt.Sprintf("value-%d", i)
		if err := tr.Upsert(key, value); err != nil {
			t.Fatal(err)
		}
	}

	values, err := tr.FindBetween("key-020", "key-080")
	if len(values) != 59 {
		t.Fatalf("GetInBetweenKeys() count = %d, expected 59", len(values))
	}

	for i := 21; i < 80; i++ {
		got := values[i-21]
		expected := fmt.Sprintf("value-%d", i)
		if err != nil {
			t.Fatalf("Get() error: %v", err)
		}
		if got != expected {
			t.Fatalf("Get() value = %q, expected %q", got, expected)
		}
	}

}

// ======================================================
// BENCHMARKS
// ======================================================

func BenchmarkGetOldest(b *testing.B) {

	tr := bst.NewBst[string, string]()

	for i := range 100 {
		key := fmt.Sprintf("key-%d", i)
		if err := tr.Upsert(key, "value"); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for b.Loop() {
		requireValue(b, tr, "key-0", "value")
	}

}

func BenchmarkLargeInserts(b *testing.B) {

	tr := bst.NewBst[string, string]()

	b.ResetTimer()

	for b.Loop() {
		for i := range 100 {
			key := fmt.Sprintf("key-%d", i)
			if err := tr.Upsert(key, "value"); err != nil {
				b.Fatal(err)
			}
		}
	}

	requireValue(b, tr, "key-0", "value")
}

func BenchmarkRandomDataGetAndDelete(b *testing.B) {

	cache := make(map[string]string)
	maxKey := 10

	tr := bst.NewBst[string, string]()
	randomDataCreation(b, tr, cache, maxKey, 100, 2)

	b.ResetTimer()
	for b.Loop() {
		randomDataValidation(b, tr, cache, maxKey)
	}
}

// func BenchmarkGetBetweenKeys(b *testing.B) {

// 	dirPath := b.TempDir()
// 	tr := newTesttr(b, dirPath, 3)

// 	for i := range 100 {
// 		key := fmt.Sprintf("key-%03d", i)
// 		value := fmt.Sprintf("value-%d", i)
// 		if err := tr.Upsert(key, value); err != nil {
// 			b.Fatal(err)
// 		}
// 	}

// 	var values []string
// 	var err error

// 	b.ResetTimer()
// 	for b.Loop() {
// 		values, err = tr.GetInBetweenKeys("key-020", "key-080")
// 	}

// 	if len(values) != 59 {
// 		b.Fatalf("GetInBetweenKeys() count = %d, expected 59", len(values))
// 	}

// 	for i := 21; i < 80; i++ {
// 		got := values[i-21]
// 		expected := fmt.Sprintf("value-%d", i)
// 		if err != nil {
// 			b.Fatalf("Get() error: %v", err)
// 		}
// 		if got != expected {
// 			b.Fatalf("Get() value = %q, expected %q", got, expected)
// 		}
// 	}

// }
