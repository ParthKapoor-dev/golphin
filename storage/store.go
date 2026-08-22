package storage

import (
	"io"
	"os"
	"strings"
)

type Db struct {
	file     *os.File
	filePath string
}

func getFile(filePath string) (*os.File, error) {
	return os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
}

func GetDB(filePath string) (*Db, error) {
	file, err := getFile(filePath)
	if err != nil {
		return nil, err
	}

	return &Db{file, filePath}, nil
}

func (db *Db) Get(key string) (bool, string, error) {

	revReader, err := newReverseReader(db.file)
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

func (db *Db) Upsert(key string, value string) error {
	_, err := db.file.WriteString(key + ":" + value + "\n")
	return err
}

func (db *Db) Delete(key string) error {
	_, err := db.file.WriteString(key + ":" + "\000" + "\n")
	return err
}

func (db *Db) Close() {
	db.file.Close()
}
