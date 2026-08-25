package storage

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Db struct {
	segments             []*segment
	maxRecordsPerSegment int
	dirPath              string
	size                 int
	idxCache             map[string]*IdxRecord
}

func (db *Db) initIndexing() error {

	size := len(db.segments)
	for i := size - 1; i >= 0; i-- {
		db.segments[i].indexing(db.idxCache)
	}

	return nil
}

func (db *Db) compact() error {
	cache := make(map[string]bool)
	size := len(db.segments)
	for i := size - 1; i >= 0; i-- {
		if err := db.segments[i].compact(cache); err != nil {
			return err
		}
	}

	return nil
}

func (db *Db) addSegment() (*segment, error) {
	size := len(db.segments)
	// TODO: fix this hardcoded way for the next file name
	newIdx := size + 1
	filepath := db.dirPath + "/" + strconv.Itoa(newIdx) + ".txt"

	seg, err := newSegment(newIdx, filepath)
	if err != nil {
		return nil, err
	}

	db.segments = append(db.segments, seg)
	return seg, err
}

func (db *Db) getSegment() (*segment, error) {
	size := len(db.segments)
	var seg = db.segments[size-1]
	if seg.count >= db.maxRecordsPerSegment {
		// TODO: make this bg process; so here we just invoke compaction
		if err := db.compact(); err != nil {
			return nil, err
		}
		return db.addSegment()
	}
	return seg, nil
}

func GetDB(dirPath string, maxRecordsPerSegment int) (*Db, error) {

	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir %q: %w", dirPath, err)
	}

	var segments []*segment

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), ".txt") {
			filepath := dirPath + "/" + file.Name()

			fileId, err := strconv.Atoi(strings.Split(file.Name(), ".txt")[0])
			if err != nil {
				return nil, fmt.Errorf("get fileId: %w", err)
			}

			seg, err := newSegment(fileId, filepath)
			if err != nil {
				return nil, err
			}

			segments = append(segments, seg)
		}
	}

	slices.SortFunc(segments, func(a, b *segment) int {
		return cmp.Compare(a.id, b.id)
	})

	size := 0
	idxCache := make(map[string]*IdxRecord)

	db := &Db{
		segments,
		maxRecordsPerSegment,
		dirPath,
		size,
		idxCache,
	}

	if err := db.initIndexing(); err != nil {
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
		count += seg.count
	}
	return count, nil
}

func (db *Db) Get(key string) (bool, string, error) {

	if err := db.initIndexing(); err != nil {
		return false, "", err
	}

	idxRec, exists := db.idxCache[key]
	if !exists {
		return false, "", nil
	}

	value, err := idxRec.get()
	if err != nil || value == tombstone {
		return false, "", err
	}

	return true, value, nil
}

func (db *Db) legacyGet(key string) (bool, string, error) {

	size := len(db.segments)

	for i := size - 1; i >= 0; i-- {
		found, value, err := db.segments[i].find(key)
		if err != nil {
			return false, "", err
		}
		if !found {
			continue
		}
		if value == tombstone {
			return false, "", nil
		}
		return true, value, nil
	}

	return false, "", nil
}

func (db *Db) Set(key string, value string) error {
	seg, err := db.getSegment()
	if err != nil {
		return err
	}
	return seg.upsert(key, value)
}

func (db *Db) Delete(key string) error {
	seg, err := db.getSegment()
	if err != nil {
		return err
	}
	return seg.delete(key)
}

func (db *Db) Close() {
	for _, seg := range db.segments {
		seg.close()
	}
}
