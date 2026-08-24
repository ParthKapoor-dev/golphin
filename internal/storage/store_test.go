package storage_test

import (
	"fmt"
	"testing"

	"github.com/parthkapoor-dev/golphin/internal/storage"
)

func newTestDB(t *testing.T, maxRecords int) *storage.Db {

	// t.tempDir --> that exists during testing
	db, err := storage.GetDB(t.TempDir(), maxRecords)
	if err != nil {
		// error logging during testing
		t.Fatalf("GetDB() error : %v", err)
	}
	// t.cleanup is like defer, but during testing
	t.Cleanup(db.Close)
	return db
}

func requireValue(t *testing.T, db *storage.Db, key string, value string) {
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

func requireMissing(t *testing.T, db *storage.Db, key string) {
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

func TestSetMakesValueRetrievable(t *testing.T) {

	db := newTestDB(t, 3)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "lang", "go")
}

func TestSetReplacesExistingValue(t *testing.T) {

	db := newTestDB(t, 3)

	if err := db.Set("lang", "typescript"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "lang", "go")
}

func TestDeleteMakesKeyUnavailable(t *testing.T) {

	db := newTestDB(t, 3)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	if err := db.Delete("lang"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	requireMissing(t, db, "lang")
}

func TestDataSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	oldDb, err := storage.GetDB(dir, 2)
	if err != nil {
		t.Fatalf("GetDB() error : %v", err)
	}
	t.Cleanup(oldDb.Close)
	oldDb.Close()

	db, err := storage.GetDB(dir, 2)
	if err != nil {
		t.Fatalf("GetDB() error : %v", err)
	}
	t.Cleanup(db.Close)

	if err := db.Set("lang", "go"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	requireValue(t, db, "lang", "go")
}

func TestDeleteAcrossSegments(t *testing.T) {

	db := newTestDB(t, 2)

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

// benchmarking

func BenchmarkGetOldest(b *testing.B) {

	db, err := storage.GetDB(b.TempDir(), 10_001)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	for i := range 10_001 {
		key := fmt.Sprintf("key-%d", i)
		if err := db.Set(key, "value"); err != nil {
			b.Fatal(err)
		}
	}
	b.ResetTimer()

	for b.Loop() {
		if _, _, err := db.Get("key-0"); err != nil {
			b.Fatal(err)
		}
	}

}
