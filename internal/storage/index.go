package storage

import (
	"fmt"
	"strconv"
	"strings"
)

// From now on, this files represent Index, not an indexed record
// this is not a layer but helper for store

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
