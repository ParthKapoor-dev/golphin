package fs

import (
	"fmt"
	"io"
	"os"
)

// Idempotent: get or create file
func EnsureFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Open file %q: %w", filePath, err)
	}
	return file, nil
}

func GetChunk(file *os.File, start, end int64) ([]byte, error) {

	size := end - start
	chunk := make([]byte, size)

	n, err := file.ReadAt(chunk, start)
	if err != nil {
		return nil, fmt.Errorf("read file into chunk: %w", err)
	}

	if int64(n) != size {
		return nil, fmt.Errorf("expected %d bytes, read %d", size, n)
	}

	return chunk, nil
}

func WriteChunk(file *os.File, entry []byte) (int64, int64, error) {

	startPtr, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, 0, fmt.Errorf("failed seeking ptr for file %q: %w", file.Name(), err)
	}

	n, err := file.Write(entry)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to write in file %q: %w", file.Name(), err)
	}

	endPtr := startPtr + int64(n)

	return startPtr, endPtr, nil
}
