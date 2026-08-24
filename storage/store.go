package storage

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strconv"
)

const tombstone = "\000"

type Db struct {
	segments             []*segment
	maxRecordsPerSegment int
	dirPath              string
}

func (db *Db) compact() error {
	fmt.Println("Starting DB compaction")
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
	filepath := db.dirPath + "/" + strconv.Itoa(size+1) + ".txt"
	seg, err := newSegment(filepath)
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
		if !file.IsDir() {
			filepath := dirPath + "/" + file.Name()
			seg, err := newSegment(filepath)
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
		}
	}

	slices.SortFunc(segments, func(a, b *segment) int {
		return cmp.Compare(a.id, b.id)
	})

	db := &Db{
		segments,
		maxRecordsPerSegment,
		dirPath,
	}

	if len(segments) == 0 {
		if _, err := db.addSegment(); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (db *Db) Get(key string) (bool, string, error) {

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
