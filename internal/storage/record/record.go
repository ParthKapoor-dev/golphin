package record

import "strings"

type Record struct {
	Key   string
	Value string
	Entry string
	Chunk []byte
}

const Tombstone = "\000"

func Encode(key, value string) *Record {

	entry := key + ":" + value + "\n"
	chunk := []byte(entry)
	return &Record{key, value, entry, chunk}
}

func Decode(line string) *Record {
	line = strings.TrimSuffix(line, "\n")
	kv := strings.Split(line, ":")

	if len(kv) != 2 {
		return nil
	}

	return &Record{
		Key:   kv[0],
		Value: kv[1],
		Entry: line,
	}

}
