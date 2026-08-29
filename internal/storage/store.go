package storage

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/parthkapoor-dev/golphin/internal/fs"
)

type Db struct {
	segments             []*segment
	maxRecordsPerSegment int
	dirPath              string
	size                 int
	idxCache             map[string]*IdxRecord
}

func (db *Db) initIndexing() error {

	db.idxCache = make(map[string]*IdxRecord)

	filePath := db.dirPath + "/_snapshot.txt"
	snapFile, err := os.Open(filePath)
	if errors.Is(err, os.ErrNotExist) {

		size := len(db.segments)
		for i := size - 1; i >= 0; i-- {
			if err := db.segments[i].indexing(db.idxCache); err != nil {
				return err
			}
		}

		return nil
	}

	if err != nil {
		return err
	}
	defer snapFile.Close()

	revReader, err := fs.NewReverseReader(snapFile)
	if err != nil {
		return err
	}
	defer revReader.Close()

	for {
		line, _, err := revReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rec := strings.Split(line, ":")
		if len(rec) == 4 {
			key := rec[0]

			segId, err := strconv.Atoi(rec[1])
			if err != nil {
				return err
			}

			// TODO: this would fail later on.
			if len(db.segments) > segId {
				return fmt.Errorf("Incorrect segId")
			}
			seg := db.segments[segId-1]

			start, err := strconv.ParseInt(rec[2], 10, 64)
			if err != nil {
				return err
			}

			end, err := strconv.ParseInt(rec[3], 10, 64)
			if err != nil {
				return err
			}

			db.idxCache[key] = newIdxRecord(key, seg, start, end)

		}
	}

	if err := os.Remove(filePath); err != nil {
		return err
	}

	return nil
}

func (db *Db) snapshot() error {

	filePath := db.dirPath + "/_snapshot.txt"
	snapFile, err := fs.EnsureFile(filePath)
	if err != nil {
		return err
	}
	defer snapFile.Close()

	for key, rec := range db.idxCache {
		chunk := []byte(key + ":" + strconv.Itoa(rec.segment.id) + ":" + strconv.FormatInt(rec.start, 10) + ":" + strconv.FormatInt(rec.end, 10) + "\n")
		fs.WriteChunk(snapFile, chunk)
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

	if err := db.initIndexing(); err != nil {
		return err
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
		if !file.IsDir() && strings.Contains(file.Name(), ".txt") && file.Name() != "_snapshot.txt" {
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
	var idxCache map[string]*IdxRecord
	idxCache = nil

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

	idxRec, exists := db.idxCache[key]
	if !exists {
		return false, "", nil
	}

	found, value, err := idxRec.segment.get(idxRec)
	if err != nil {
		return false, "", err
	}
	if !found || value == tombstone {
		return false, "", nil
	}

	return true, value, nil
}

func (db *Db) legacyGet(key string) (bool, string, error) {

	size := len(db.segments)

	for i := size - 1; i >= 0; i-- {
		found, value, err := db.segments[i].search(key)
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

	start, end, err := seg.upsert(key, value)
	if err != nil {
		return err
	}

	db.idxCache[key] = newIdxRecord(key, seg, start, end)

	return nil
}

func (db *Db) Delete(key string) error {
	seg, err := db.getSegment()
	if err != nil {
		return err
	}

	if _, _, err = seg.delete(key); err != nil {
		return err
	}

	// db.idxCache[key] = newIdxRecord(key, seg, start, end)
	delete(db.idxCache, key)

	return nil
}

func (db *Db) Close() {

	db.snapshot()

	for _, seg := range db.segments {
		seg.close()
	}
}
