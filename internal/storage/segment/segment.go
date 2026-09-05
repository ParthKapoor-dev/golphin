package segment

import (
	"fmt"
	"io"
	"os"

	"github.com/parthkapoor-dev/golphin/internal/fs"
	record "github.com/parthkapoor-dev/golphin/internal/storage/record"
)

type Segment struct {
	Id    int
	Count int
	file  *os.File
}

func NewSegment(fileId int, filepath string) (*Segment, error) {
	file, err := fs.EnsureFile(filepath)
	if err != nil {
		return nil, err
	}

	lineCount, err := fs.GetLineCount(file)
	if err != nil {
		return nil, fmt.Errorf("file: %q", fileId)
	}

	return &Segment{
		file:  file,
		Count: lineCount,
		Id:    fileId,
	}, nil
}

func (seg *Segment) Get(start, end int64) (bool, string, error) {

	chunk, err := fs.GetChunk(seg.file, start, end)
	if err != nil {
		return false, "", err
	}

	rec := record.Decode(string(chunk))
	if rec == nil {
		return false, "", fmt.Errorf("decode entry failed")
	}

	return true, rec.Value, nil
}

func (seg *Segment) Upsert(key string, value string) (int64, int64, error) {

	rec := record.Encode(key, value)

	start, end, err := fs.WriteChunk(seg.file, rec.Chunk)
	if err != nil {
		return start, end, err
	}

	seg.Count++

	return start, end, nil
}

func (seg *Segment) Delete(key string) (int64, int64, error) {

	rec := record.Encode(key, record.Tombstone)

	start, end, err := fs.WriteChunk(seg.file, rec.Chunk)
	if err != nil {
		return start, end, err
	}

	seg.Count++

	return start, end, nil
}

func (seg *Segment) Compact(cache map[string]bool) error {

	var skips []fs.Skip

	if err := seg.readlineIterator(func(line string, pos int64) error {
		rec := record.Decode(line)
		if rec == nil {
			return fmt.Errorf("decode failed for line: %q", line)
		}

		if cache[rec.Key] || rec.Value == record.Tombstone {
			skips = append(skips, fs.Skip{Start: pos, End: pos + int64(len(line)) + 1})
		}
		cache[rec.Key] = true

		return nil
	}); err != nil {
		return err
	}

	if len(skips) != 0 {
		file, err := fs.ReplaceFile(skips, seg.file)
		if err != nil {
			return err
		}

		seg.file = file
		seg.Count -= len(skips)
	}

	return nil
}

func (seg *Segment) Indexing(do func(string, int, int64, int64) error) error {
	return seg.readlineIterator(func(line string, pos int64) error {
		if rec := record.Decode(line); rec != nil {
			do(rec.Key, seg.Id, pos, pos+int64(len(line))+1)
			return nil
		}
		return fmt.Errorf("unable to decode line: %q", line)
	})
}

func (seg *Segment) Close() {
	seg.file.Close()
}

// ==============================================================
// LEGACY CODE
// ==============================================================

func (seg *Segment) Search(key string) (bool, string, error) {

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

		rec := record.Decode(line)
		if rec != nil && rec.Key == key {
			return true, rec.Value, nil
		}
	}
	return false, "", nil
}
