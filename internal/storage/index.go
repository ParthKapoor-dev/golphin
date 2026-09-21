package storage

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/parthkapoor-dev/golphin/pkg/avl"
	"github.com/parthkapoor-dev/golphin/pkg/bst"
)

// ======================================================
// SINGLE ENTRY LOCATION
// ======================================================

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

// ======================================================
// INDEX STORE
// ======================================================

var (
	_ orderedMap[string, *location] = (*avl.BST[string, *location])(nil)
	_ orderedMap[string, *location] = (*bst.BST[string, *location])(nil)
)

type orderedMap[K cmp.Ordered, V any] interface {
	Find(key K) (bool, V, error)
	Upsert(key K, loc V) error
	Delete(key K) error
	FindBetween(lo K, hi K) ([]V, error)
	Iter() ([]V, error)
	Len() int
}

type index struct {
	tree orderedMap[string, *location]
}

func NewIndex() index {
	b := avl.NewBst[string, *location]()
	return index{b}
}

func (idx index) get(key string) (bool, *location, error) {
	return idx.tree.Find(key)
}

func (idx index) set(key string, loc *location) error {
	return idx.tree.Upsert(key, loc)
}

func (idx index) delete(key string) error {
	return idx.tree.Delete(key)
}

func (idx index) between(from string, to string) ([]*location, error) {
	return idx.tree.FindBetween(from, to)
}

func (idx index) iter() ([]*location, error) {
	return idx.tree.Iter()
}

func (idx index) size() int {
	return idx.tree.Len()
}
