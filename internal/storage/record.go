package storage

import "strings"

type Record struct {
	key   string
	value string
	entry string
	chunk []byte
}

const tombstone = "\000"

func encode(key, value string) *Record {

	entry := key + ":" + value + "\n"
	chunk := []byte(entry)
	return &Record{key, value, entry, chunk}
}

func decode(line string) *Record {
	line = strings.TrimSuffix(line, "\n")
	kv := strings.Split(line, ":")

	if len(kv) != 2 {
		return nil
	}

	return &Record{
		key:   kv[0],
		value: kv[1],
		entry: line,
	}

}
