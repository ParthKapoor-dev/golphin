package storage

import (
	"bufio"
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

func (db *Db) Get(key string) (string, error) {
	scanner := bufio.NewScanner(db.file)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		kv := strings.Split(line, ":")
		if kv[0] == key {
			return kv[1], nil
		}
	}

	return "", nil
}

func (db *Db) Upsert(key string, value string) error {
	_, err := db.file.WriteString(key + ":" + value + "\n")
	return err
}

func (db *Db) Close() {
	db.file.Close()
}
