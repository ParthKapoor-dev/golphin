package storage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type segment struct {
	id    int
	file  *os.File
	count int
}

type Pair struct {
	start int64
	end   int64
}

// get or create file
func getFile(filePath string) (*os.File, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("Open file %q: %w", filePath, err)
	}
	return file, nil
}

func newSegment(filepath string, dirPath string) (*segment, error) {
	file, err := getFile(filepath)
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

func (seg *segment) replaceSegment(pairs []Pair) error {

	var currentPos int64 = 0

	oldFileName := seg.file.Name()
	oldFile := seg.file

	newFileName := strings.Split(oldFileName, ".txt")[0] + "_new.txt"
	newFile, err := getFile(newFileName)
	if err != nil {
		return err
	}

	// for each skip
	skipSize := len(pairs)
	for i := skipSize - 1; i >= 0; i-- {
		skip := pairs[i]
		bytesToCopy := skip.start - currentPos

		if bytesToCopy > 0 {

			if _, err := oldFile.Seek(currentPos, io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek: %w", err)
			}

			if _, err := io.CopyN(newFile, oldFile, bytesToCopy); err != nil {
				return fmt.Errorf("failed to copy chunk: %w", err)
			}

		}

		if skip.end > currentPos {
			currentPos = skip.end
		}
	}

	// copy whatever is left
	if _, err := oldFile.Seek(currentPos, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	if _, err := io.Copy(newFile, oldFile); err != nil {
		return fmt.Errorf("failed to copy chunk: %w", err)
	}

	newFileStat, err := newFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat new file: %w", err)
	}

	oldFileStat, err := oldFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat old file: %w", err)
	}

	oldFile.Close()
	newFile.Close()

	fileCompactionSuccess := newFileStat.Size() < oldFileStat.Size()

	if fileCompactionSuccess {
		if err := os.Remove(oldFileName); err != nil {
			return fmt.Errorf("failed to delete original src file: %w", err)
		}

		if err := os.Rename(newFileName, oldFileName); err != nil {
			return fmt.Errorf("failed to rename dst file: %w", err)
		}

	} else {
		os.Remove(newFileName)
	}

	finalFile, err := getFile(oldFileName)
	if err != nil {
		return err
	}
	seg.file = finalFile

	if !fileCompactionSuccess {
		return fmt.Errorf("compaction process failed!")
	}

	seg.count -= skipSize
	return nil
}

func (seg *segment) compact(cache map[string]bool) error {

	var pairs []Pair

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
		if len(kv) == 2 {
			if cache[kv[0]] || kv[1] == tombstone {
				// we shouldn't delete in place. but create a new file
				// lets get all positions that we don't want
				pairs = append(pairs, Pair{start: pos, end: pos + int64(len(line)) + 1})
			}
			if !cache[kv[0]] || kv[1] == tombstone {
				cache[kv[0]] = true
			}

		}
	}

	if len(pairs) != 0 {
		// create new file that we need to replace this segment with
		if err := seg.replaceSegment(pairs); err != nil {
			return fmt.Errorf("replacement failed: %w", err)
		}
	}

	return nil
}

func (seg *segment) close() {
	seg.file.Close()
}
