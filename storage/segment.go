package storage

import (
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
		return nil, err
	}

	return &segment{
		file:  file,
		count: 0,
	}, nil
}

func (seg *segment) find(key string) (bool, string, error) {

	revReader, err := newReverseReader(seg.file)
	if err != nil {
		return false, "", err
	}
	defer revReader.close()

	for {
		line, err := revReader.readLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return false, "", err
		}

		kv := strings.Split(line, ":")
		if kv[0] == key {
			if kv[1] == "\000" {
				return false, "", nil
			}
			return true, kv[1], nil
		}
	}
	return false, "", nil
}

func (seg *segment) upsert(key string, value string) error {
	_, err := seg.file.WriteString(key + ":" + value + "\n")
	return err
}

func (seg *segment) delete(key string) error {
	_, err := seg.file.WriteString(key + ":" + "\000" + "\n")
	return err
}

func (seg *segment) close() {
	seg.file.Close()
}
