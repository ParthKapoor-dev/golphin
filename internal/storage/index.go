package storage

import (
	"fmt"
	"strconv"
	"strings"
)

type location struct {
	key   string
	segId int
	start int64
	end   int64
}

func newLocation(key string, segId int, start, end int64) *location {
	return &location{key, segId, start, end}
}

func decodeIdx(entry string) (*location, error) {

	rec := strings.Split(entry, ":")
	if len(rec) == 4 {
		key := rec[0]

		segId, err := strconv.Atoi(rec[1])
		if err != nil {
			return nil, err
		}

		start, err := strconv.ParseInt(rec[2], 10, 64)
		if err != nil {
			return nil, err
		}

		end, err := strconv.ParseInt(rec[3], 10, 64)
		if err != nil {
			return nil, err
		}

		return &location{key, segId, start, end}, nil
	}

	return nil, fmt.Errorf("invalid index record of length: %d", len(rec))
}

func (loc *location) encodeIdx() []byte {
	return []byte(
		loc.key + ":" + strconv.Itoa(loc.segId) + ":" +
			strconv.FormatInt(loc.start, 10) + ":" +
			strconv.FormatInt(loc.end, 10) + "\n")
}

type index map[string]*location

func NewIndex() index {
	return make(index)
}

func (idx index) get(key string) (*location, bool) {
	loc, exists := idx[key]
	return loc, exists
}

func (idx index) set(key string, loc *location) {
	idx[key] = loc
}

func (idx index) delete(key string) {
	delete(idx, key)
}

func (idx index) iter(do func(*location) error) error {

	for _, loc := range idx {
		if err := do(loc); err != nil {
			return err
		}
	}

	return nil
}
