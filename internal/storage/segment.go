package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/parthkapoor-dev/golphin/internal/fs"
)

type segment struct {
	id    int
	file  *os.File
	count int
}

func newSegment(filepath string, dirPath string) (*segment, error) {
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

	// TODO: better way for finding the id
	segId, err := strconv.Atoi(strings.Split(strings.Split(filepath, ".txt")[0], dirPath+"/")[1])
	if err != nil {
		return nil, err
	}

	return &segment{
		file:  file,
		count: lineCount,
		id:    segId,
	}, nil
}

func (seg *segment) find(key string) (bool, string, error) {

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

func (seg *segment) replaceSegment(skips []fs.Skip) error {

	file, err := fs.ReplaceFile(skips, seg.file)
	if err != nil {
		return err
	}

	seg.file = file
	seg.count -= len(skips)
	return nil
}

func (seg *segment) compact(cache map[string]bool) error {

	var pairs []fs.Skip

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

		kv := strings.Split(line, ":")
		if len(kv) == 2 {
			if cache[kv[0]] || kv[1] == tombstone {
				pairs = append(pairs, fs.Skip{Start: pos, End: pos + int64(len(line)) + 1})
			}
			cache[kv[0]] = true
		}
	}

	if len(pairs) != 0 {
		if err := seg.replaceSegment(pairs); err != nil {
			return fmt.Errorf("replacement failed: %w", err)
		}
	}

	return nil
}

func (seg *segment) close() {
	seg.file.Close()
}
