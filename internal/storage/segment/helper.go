package segment

import (
	"io"

	"github.com/parthkapoor-dev/golphin/internal/fs"
)

func (seg *Segment) readlineIterator(do func(line string, pos int64) error) error {

	revReader, err := fs.NewReverseReader(seg.file)
	if err != nil {
		return err
	}
	defer revReader.Close()

	for {
		line, pos, err := revReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := do(line, pos); err != nil {
			return err
		}
	}
	return nil
}
