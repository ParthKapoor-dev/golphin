package storage

// an indexed record
type IdxRecord struct {
	key     string
	segment *segment
	start   int64
	end     int64
}

func newIdxRecord(key string, segment *segment, start, end int64) *IdxRecord {
	return &IdxRecord{key, segment, start, end}
}
