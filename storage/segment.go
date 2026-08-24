package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type segment struct {
	file  *os.File
	count int
}

// get or create file
func getFile(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
}

func newSegment(filepath string) (*segment, error) {
	file, err := getFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("open segment: %w", err)
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
	}, nil
}

func (seg *segment) find(key string) (bool, string, error) {

	revReader, err := newReverseReader(seg.file)
	if err != nil {
		return false, "", err
	}
	defer revReader.close()

	for {
		line, _, err := revReader.readLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, "", err
		}

		kv := strings.Split(line, ":")
		if kv[0] == key {
			return true, kv[1], nil
		}
	}
	return false, "", nil
}

func (seg *segment) upsert(key string, value string) error {
	if _, err := seg.file.WriteString(key + ":" + value + "\n"); err != nil {
		return fmt.Errorf("Unable to write to file %q, %q : %w", seg.file.Name(), key, err)
	}
	seg.count++
	return nil
}

func (seg *segment) delete(key string) error {
	if _, err := seg.file.WriteString(key + ":" + tombstone + "\n"); err != nil {
		return fmt.Errorf("Unable to write to file %q, %q : %w", seg.file.Name(), key, err)
	}
	seg.count++
	return nil
}

func (seg *segment) compact(cache map[string]bool) error {

	revReader, err := newReverseReader(seg.file)
	if err != nil {
		return err
	}
	defer revReader.close()

	for {
		line, pos, err := revReader.readLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		kv := strings.Split(line, ":")
		if len(kv) == 2 && cache[kv[0]] {
			// remove this line
			if err := revReader.deleteLine(pos, len(line)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (seg *segment) close() {
	seg.file.Close()
}
