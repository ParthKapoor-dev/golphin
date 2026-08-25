package storage

import (
	"fmt"
	"io"
)

// an indexed record
type IdxRecord struct {
	key     string
	segment *segment
	start   int64
	end     int64
}

func newIdxRecord(key string, segment *segment, start, end int64) *IdxRecord {
	return &IdxRecord{key, segment, start, end}
}

func (idxRec *IdxRecord) get() (string, error) {

	file := idxRec.segment.file
	start := idxRec.start
	end := idxRec.end

	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return "", fmt.Errorf("Unable to find pointer in file %q, pos: %d: %w", file.Name(), start, err)
	}

	chunk := make([]byte, end-start-1)

	_, err := file.Read(chunk)
	if err != nil {
		return "", fmt.Errorf("Unable to read file into chunk: %w", err)
	}

	if rec := decode(string(chunk)); rec != nil {
		return rec.value, nil
	}

	return "", fmt.Errorf("decode entry failed")
}
