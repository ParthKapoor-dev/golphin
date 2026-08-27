package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/parthkapoor-dev/golphin/internal/fs"
)

type segment struct {
	id    int
	file  *os.File
	count int
}

func newSegment(fileId int, filepath string) (*segment, error) {
	file, err := fs.EnsureFile(filepath)
	if err != nil {
		return nil, err
	}

	lineCount := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		file.Close()
		return nil, fmt.Errorf("Unable to count records in file %q: %w", filepath, err)
	}

	return &segment{
		file:  file,
		count: lineCount,
		id:    fileId,
	}, nil
}

func (seg *segment) get(idxRec *IdxRecord) (bool, string, error) {

	chunk, err := fs.GetChunk(seg.file, idxRec.start, idxRec.end)
	if err != nil {
		return false, "", err
	}

	rec := decode(string(chunk))
	if rec == nil {
		return false, "", fmt.Errorf("decode entry failed")
	}

	return true, rec.value, nil
}

func (seg *segment) search(key string) (bool, string, error) {

	revReader, err := fs.NewReverseReader(seg.file)
	if err != nil {
		return false, "", err
	}
	defer revReader.Close()

	for {
		line, _, err := revReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, "", err
		}

		rec := decode(line)
		if rec != nil && rec.key == key {
			return true, rec.value, nil
		}
	}
	return false, "", nil
}

func (seg *segment) upsert(key string, value string) (int64, int64, error) {

	rec := encode(key, value)

	start, end, err := fs.WriteChunk(seg.file, rec.chunk)
	if err != nil {
		return start, end, err
	}

	seg.count++

	return start, end, nil
}

func (seg *segment) delete(key string) (int64, int64, error) {

	rec := encode(key, tombstone)

	start, end, err := fs.WriteChunk(seg.file, rec.chunk)
	if err != nil {
		return start, end, err
	}

	seg.count++

	return start, end, nil
}

func (seg *segment) compact(cache map[string]bool) error {

	var skips []fs.Skip

	revReader, err := fs.NewReverseReader(seg.file)
	if err != nil {
		return err
	}
	defer revReader.Close()

	for {
		line, pos, err := revReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rec := decode(line)
		if rec != nil {
			if cache[rec.key] || rec.value == tombstone {
				skips = append(skips, fs.Skip{Start: pos, End: pos + int64(len(line)) + 1})
			}
			cache[rec.key] = true
		}
	}

	if len(skips) != 0 {
		file, err := fs.ReplaceFile(skips, seg.file)
		if err != nil {
			return err
		}

		seg.file = file
		seg.count -= len(skips)
	}

	return nil
}

func (seg *segment) indexing(cache map[string]*IdxRecord) error {

	revReader, err := fs.NewReverseReader(seg.file)
	if err != nil {
		return err
	}
	defer revReader.Close()

	for {
		line, pos, err := revReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rec := decode(line)
		if rec != nil {
			_, exists := cache[rec.key]
			if !exists {
				cache[rec.key] = newIdxRecord(rec.key, seg, pos, pos+int64(len(line))+1)
			}
		}
	}

	return nil
}

func (seg *segment) close() {
	seg.file.Close()
}
