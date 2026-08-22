package storage

import (
	"fmt"
	"io"
	"os"
)

// chunkSize
const chunkSize = 4096

type ReverseReader struct {
	file *os.File
	pos  int64
	buf  []byte
}

func newReverseReader(file *os.File) (*ReverseReader, error) {

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("RevReader - File doesn't exists: %w", err)
	}

	return &ReverseReader{
		file: file,
		pos:  info.Size(),
	}, nil
}

func (r *ReverseReader) readLine() (string, error) {

	for {

		// found the new line, return now
		for i := len(r.buf) - 1; i >= 0; i-- {
			if r.buf[i] == '\n' {
				line := r.buf[i+1:]
				r.buf = r.buf[:i]

				return string(line), nil
			}
		}
		// the buffer didn't have newline

		// if postion is 0, and buffer is empty
		// meaning we are beginning of the file
		if r.pos == 0 {
			// no chunk for us to fetch
			if len(r.buf) == 0 {
				return "", io.EOF
			}
			// return the first line
			line := r.buf
			r.buf = nil
			return string(line), nil
		}

		// keep position at above chunk start
		readSize := int64(chunkSize)
		if r.pos < readSize {
			readSize = r.pos
		}
		r.pos -= readSize

		// find the file pointer for this position
		_, err := r.file.Seek(r.pos, io.SeekStart)
		if err != nil {
			return "", fmt.Errorf("Unable to find pointer in file %q, pos: %d: %w", r.file.Name(), r.pos, err)
		}

		// copy the next chunk from file position to readsize length into chunk
		chunk := make([]byte, readSize)

		n, err := r.file.Read(chunk)
		if err != nil {
			return "", fmt.Errorf("Unable to read file into chunk: %w", err)
		}

		// append it in the buffer
		r.buf = append(r.buf, chunk[:n]...)
	}
}

func (r *ReverseReader) close() {
}
