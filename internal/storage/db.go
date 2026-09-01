package storage

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	record "github.com/parthkapoor-dev/golphin/internal/storage/record"
	sg "github.com/parthkapoor-dev/golphin/internal/storage/segment"
)

type Db struct {
	segments             []*sg.Segment
	maxRecordsPerSegment int
	dirPath              string
	size                 int
	idxCache             map[string]*location
}

// ======================================================
// HELPERS
// ======================================================

func (db *Db) buildIndex() error {

	db.idxCache = make(map[string]*location)
	size := len(db.segments)
	for i := size - 1; i >= 0; i-- {
		if err := db.segments[i].Indexing(func(key string, segId int, start, end int64) {
			_, exists := db.idxCache[key]
			if !exists {
				db.idxCache[key] = newLocation(key, segId, start, end)
			}
		}); err != nil {
			return err
		}
	}

	return nil
}

func (db *Db) initIndex() error {

	db.idxCache = make(map[string]*location)

	err := db.readSnapshot()

	if err != nil {
		// TODO: make sure snapshot file is deleted

		if err := db.buildIndex(); err != nil {
			return err
		}
	}
	return nil
}

func (db *Db) compact() error {

	cache := make(map[string]bool)

	size := len(db.segments)
	for i := size - 1; i >= 0; i-- {
		if err := db.segments[i].Compact(cache); err != nil {
			return err
		}
	}

	if err := db.buildIndex(); err != nil {
		return err
	}

	return nil
}

func (db *Db) addSegment() (*sg.Segment, error) {
	size := len(db.segments)
	// TODO: fix this hardcoded way for the next file name
	newIdx := size + 1
	filepath := db.dirPath + "/" + strconv.Itoa(newIdx) + ".txt"

	seg, err := sg.NewSegment(newIdx, filepath)
	if err != nil {
		return nil, err
	}

	db.segments = append(db.segments, seg)
	return seg, err
}

// Idempotent: Get or Create New Segment
func (db *Db) ensureSegment() (*sg.Segment, error) {
	size := len(db.segments)
	var seg = db.segments[size-1]
	if seg.Count >= db.maxRecordsPerSegment {
		// TODO: make this bg process; so here we just invoke compaction
		if err := db.compact(); err != nil {
			return nil, err
		}
		return db.addSegment()
	}
	return seg, nil
}

// ======================================================
// API
// ======================================================

func GetDB(dirPath string, maxRecordsPerSegment int) (*Db, error) {

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir %q: %w", dirPath, err)
	}

	var segments []*sg.Segment

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), ".txt") && file.Name() != "_snapshot.txt" {
			filepath := dirPath + "/" + file.Name()

			fileId, err := strconv.Atoi(strings.Split(file.Name(), ".txt")[0])
			if err != nil {
				return nil, fmt.Errorf("get fileId: %w", err)
			}

			seg, err := sg.NewSegment(fileId, filepath)
			if err != nil {
				return nil, err
			}

			segments = append(segments, seg)
		}
	}

	slices.SortFunc(segments, func(a, b *sg.Segment) int {
		return cmp.Compare(a.Id, b.Id)
	})

	size := 0
	var idxCache map[string]*location
	idxCache = nil

	db := &Db{
		segments,
		maxRecordsPerSegment,
		dirPath,
		size,
		idxCache,
	}

	if err := db.initIndex(); err != nil {
		return nil, err
	}

	if len(segments) == 0 {
		if _, err := db.addSegment(); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (db *Db) GetSize() (int, error) {
	var count = 0
	for _, seg := range db.segments {
		count += seg.Count
	}
	return count, nil
}

func (db *Db) Get(key string) (bool, string, error) {

	loc, exists := db.idxCache[key]
	if !exists {
		return false, "", nil
	}

	// TODO: need to fix this with a map of segments
	seg := db.segments[loc.segId-1]

	found, value, err := seg.Get(loc.start, loc.end)
	if err != nil {
		return false, "", err
	}
	if !found || value == record.Tombstone {
		return false, "", nil
	}

	return true, value, nil
}

func (db *Db) Set(key string, value string) error {
	seg, err := db.ensureSegment()
	if err != nil {
		return err
	}

	start, end, err := seg.Upsert(key, value)
	if err != nil {
		return err
	}

	db.idxCache[key] = newLocation(key, seg.Id, start, end)

	return nil
}

func (db *Db) Delete(key string) error {
	seg, err := db.ensureSegment()
	if err != nil {
		return err
	}

	if _, _, err = seg.Delete(key); err != nil {
		return err
	}

	delete(db.idxCache, key)

	return nil
}

// Keys returns all keys between fromKey and toKey, exclusive.
// NOTE: this is lexicographical search
func (db *Db) GetInBetweenKeys(fromKey string, toKey string) ([]string, error) {

	results := make([]string, 0)

	for key, loc := range db.idxCache {

		if key > fromKey && key < toKey {

			// TODO: need to fix this with a map of segments
			seg := db.segments[loc.segId-1]

			found, value, err := seg.Get(loc.start, loc.end)
			if err != nil {
				return nil, err
			}
			if !found || value == record.Tombstone {
				continue
			}

			results = append(results, value)
		}
	}

	slices.Sort(results)

	return results, nil
}

func (db *Db) Close() {

	db.writeSnapshot()

	for _, seg := range db.segments {
		seg.Close()
	}
}
