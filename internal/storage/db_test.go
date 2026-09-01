package storage_test

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"

	"github.com/parthkapoor-dev/golphin/internal/storage"
)

// ======================================================
// HELPERS
// ======================================================

type testingCtx interface {
	Helper()
	Fatalf(format string, args ...any)
	Cleanup(func())
	TempDir() string
}

func newTestDB(t testingCtx, dirPath string, maxRecords int) *storage.Db {

	// t.tempDir --> that exists during testing
	db, err := storage.GetDB(dirPath, maxRecords)
	if err != nil {
		// error logging during testing
		t.Fatalf("GetDB() error : %v", err)
	}
	// t.cleanup is like defer, but during testing
	t.Cleanup(db.Close)
	return db
}

func requireValue(t testingCtx, db *storage.Db, key string, value string) {
	t.Helper()

	found, got, err := db.Get(key)
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

func requireMissing(t testingCtx, db *storage.Db, key string) {
	t.Helper()

	found, value, err := db.Get(key)
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

func randomDataCreation(t testingCtx, db *storage.Db, cache map[string]string, maxKey int, oprCnt int, deletionMultiple int) {

	for i := range oprCnt {
		key := strconv.Itoa(rand.Intn(maxKey))
		value := strconv.Itoa(rand.Intn(10 * maxKey))

		_, exists := cache[key]

		if i%deletionMultiple == 0 && exists {
			err := db.Delete(key)
			if err != nil {
				t.Fatalf("Delete() error: %v", err)
			}
			delete(cache, key)
		} else {
			err := db.Set(key, value)
			if err != nil {
				t.Fatalf("Set() error: %v", err)
			}
			cache[key] = value
		}
	}

}

func randomDataValidation(t testingCtx, db *storage.Db, cache map[string]string, maxKey int) {

	for i := range maxKey {
		key := strconv.Itoa(i)
		value, ok := cache[key]
		if ok {
			requireValue(t, db, key, value)
		} else {
			requireMissing(t, db, key)
		}
	}

	dbCnt, err := db.GetSize()
	if err != nil {
		t.Fatalf("%w", err)
	}

	if dbCnt > len(cache)+1 {
		t.Fatalf("compaction failed!")
	}

}

// ======================================================
// TESTS
// ======================================================

func TestSetMakesValueRetrievable(t *testing.T) {

	db := newTestDB(t, t.TempDir(), 3)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "lang", "go")
}

func TestRetrievableSetAfterGet(t *testing.T) {

	db := newTestDB(t, t.TempDir(), 3)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "lang", "go")

	if err := db.Set("claude", "code"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "claude", "code")
}

func TestSetReplacesExistingValue(t *testing.T) {

	db := newTestDB(t, t.TempDir(), 3)

	if err := db.Set("lang", "typescript"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "lang", "go")
}

func TestDeleteMakesKeyUnavailable(t *testing.T) {

	db := newTestDB(t, t.TempDir(), 3)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	if err := db.Delete("lang"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	requireMissing(t, db, "lang")
}

func TestDataSurvivesReopen(t *testing.T) {
	dirPath := t.TempDir()
	oldDb := newTestDB(t, dirPath, 2)
	oldDb.Close()

	db := newTestDB(t, dirPath, 2)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "lang", "go")
}

func TestDeleteAcrossSegments(t *testing.T) {

	db := newTestDB(t, t.TempDir(), 2)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	if err := db.Set("chain", "smith"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	if err := db.Delete("lang"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	requireMissing(t, db, "lang")
}

func TestRandomDataAndCompaction(t *testing.T) {

	dirPath := t.TempDir()
	cache := make(map[string]string)
	maxKey := 10

	db := newTestDB(t, dirPath, 3)

	randomDataCreation(t, db, cache, maxKey, 1000, 2)
	randomDataValidation(t, db, cache, maxKey)
}

func TestSnaphotRetrievable(t *testing.T) {

	dirPath := t.TempDir()
	cache := make(map[string]string)
	maxKey := 10

	db := newTestDB(t, dirPath, 3)
	randomDataCreation(t, db, cache, maxKey, 1000, 2)
	db.Close()

	db = newTestDB(t, dirPath, 3)
	randomDataValidation(t, db, cache, maxKey)
}

func TestNoSnapshotRecover(t *testing.T) {

	dirPath := t.TempDir()
	cache := make(map[string]string)
	maxKey := 10

	db := newTestDB(t, dirPath, 3)
	randomDataCreation(t, db, cache, maxKey, 1000, 2)
	db.Close()

	os.Remove(dirPath + "/_snapshot.txt")

	db = newTestDB(t, dirPath, 3)
	randomDataValidation(t, db, cache, maxKey)

}

func TestFaultySnapshotRecover(t *testing.T) {

	dirPath := t.TempDir()
	cache := make(map[string]string)
	maxKey := 10

	db := newTestDB(t, dirPath, 3)
	randomDataCreation(t, db, cache, maxKey, 1000, 2)
	db.Close()

	os.WriteFile(dirPath+"/_snapshot.txt", []byte("invalid:data\n"), 0644)

	db = newTestDB(t, dirPath, 3)
	requireMissing(t, db, "invalid")
	randomDataValidation(t, db, cache, maxKey)
}

func TestGetBetweenKeys(t *testing.T) {

	dirPath := t.TempDir()
	db := newTestDB(t, dirPath, 3)

	for i := range 100 {
		key := fmt.Sprintf("key-%03d", i)
		value := fmt.Sprintf("value-%d", i)
		if err := db.Set(key, value); err != nil {
			t.Fatal(err)
		}
	}

	values, err := db.GetInBetweenKeys("key-020", "key-080")
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

	db := newTestDB(b, b.TempDir(), 3)

	for i := range 100 {
		key := fmt.Sprintf("key-%d", i)
		if err := db.Set(key, "value"); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for b.Loop() {
		requireValue(b, db, "key-0", "value")
	}

}

func BenchmarkLargeInserts(b *testing.B) {

	db := newTestDB(b, b.TempDir(), 3)

	b.ResetTimer()

	for b.Loop() {
		for i := range 100 {
			key := fmt.Sprintf("key-%d", i)
			if err := db.Set(key, "value"); err != nil {
				b.Fatal(err)
			}
		}
	}

	requireValue(b, db, "key-0", "value")
}

func BenchmarkRandomDataGetAndDelete(b *testing.B) {

	dirPath := b.TempDir()
	cache := make(map[string]string)
	maxKey := 10

	db := newTestDB(b, dirPath, 3)
	randomDataCreation(b, db, cache, maxKey, 100, 2)
	db.Close()

	db = newTestDB(b, dirPath, 3)

	b.ResetTimer()
	for b.Loop() {
		randomDataValidation(b, db, cache, maxKey)
	}
}

func BenchmarkNoSnapshots(b *testing.B) {

	dirPath := b.TempDir()
	cache := make(map[string]string)
	maxKey := 10

	db := newTestDB(b, dirPath, 3)
	randomDataCreation(b, db, cache, maxKey, 100, 2)

	for b.Loop() {

		b.StopTimer()

		db.Close()
		os.Remove(dirPath + "/_snapshot.txt")
		db = newTestDB(b, dirPath, 3)

		b.StartTimer()

		randomDataValidation(b, db, cache, maxKey)
	}
}

func BenchmarkGetBetweenKeys(b *testing.B) {

	dirPath := b.TempDir()
	db := newTestDB(b, dirPath, 3)

	for i := range 100 {
		key := fmt.Sprintf("key-%03d", i)
		value := fmt.Sprintf("value-%d", i)
		if err := db.Set(key, value); err != nil {
			b.Fatal(err)
		}
	}

	var values []string
	var err error

	b.ResetTimer()
	for b.Loop() {
		values, err = db.GetInBetweenKeys("key-020", "key-080")
	}

	if len(values) != 59 {
		b.Fatalf("GetInBetweenKeys() count = %d, expected 59", len(values))
	}

	for i := 21; i < 80; i++ {
		got := values[i-21]
		expected := fmt.Sprintf("value-%d", i)
		if err != nil {
			b.Fatalf("Get() error: %v", err)
		}
		if got != expected {
			b.Fatalf("Get() value = %q, expected %q", got, expected)
		}
	}

}
