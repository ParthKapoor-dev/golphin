package storage

import "strings"

type Record struct {
	key   string
	value string
	entry string
}

const tombstone = "\000"

func encode(key, value string) *Record {

	entry := key + ":" + value + "\n"
	return &Record{key, value, entry}
}

func decode(line string) *Record {
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
