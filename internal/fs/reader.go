package fs

import (
	"fmt"
	"io"
	"os"
)

// chunkSize
const chunkSize = 4096

type ReverseReader struct {
	file     *os.File
	fileSize int64
	pos      int64
	buf      []byte
}

func NewReverseReader(file *os.File) (*ReverseReader, error) {
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("RevReader - File doesn't exists: %w", err)
	}

	return &ReverseReader{
		file:     file,
		pos:      info.Size(),
		fileSize: info.Size(),
	}, nil
}

func (r *ReverseReader) ReadLine() (string, int64, error) {

	for {

		// check for new line at every byte in buffer in reverse
		bufSize := len(r.buf)
		for i := bufSize - 1; i >= 0; i-- {
			// found the new line, return now
			// the new line shouldn't be the last `\n`; coz that will return empty line
			if r.buf[i] == '\n' {
				if i+1 == bufSize {
					r.buf = r.buf[:i]
					continue
				}
				line := r.buf[i+1:]
				r.buf = r.buf[:i]
				startPos := r.pos + int64(i+1)

				return string(line), startPos, nil
			}
		}
		// the buffer didn't have newline

		// if postion is 0: meaning we are beginning of the file
		if r.pos == 0 {
			// no chunk for us to fetch
			if len(r.buf) == 0 {
				return "", -1, io.EOF
			}
			// return the first line
			line := r.buf
			r.buf = nil
			return string(line), 0, nil
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
			return "", -1, fmt.Errorf("Unable to find pointer in file %q, pos: %d: %w", r.file.Name(), r.pos, err)
		}

		// copy the next chunk from file position to readsize length into chunk
		chunk := make([]byte, readSize)

		n, err := r.file.Read(chunk)
		if err != nil {
			return "", -1, fmt.Errorf("Unable to read file into chunk: %w", err)
		}

		// append it in the buffer
		r.buf = append(r.buf, chunk[:n]...)
	}
}

// not used anywhere yet. but good to keep
func (r *ReverseReader) DeleteLine(startPos int64, stringLen int) error {

	lineLen := int64(stringLen) + 1
	endPos := startPos + lineLen
	bytesToMove := r.fileSize - endPos

	// read and move the remainder bytes
	if bytesToMove > 0 {
		remainder := make([]byte, bytesToMove)
		_, err := r.file.ReadAt(remainder, endPos)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read remainder: %w", err)
		}

		_, err = r.file.WriteAt(remainder, startPos)
		if err != nil {
			return fmt.Errorf("failed to write remainder: %w", err)
		}
	}

	// truncate the extra bytes
	newFileSize := r.fileSize - lineLen
	if err := r.file.Truncate(newFileSize); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	r.fileSize = newFileSize
	return nil
}

func (r *ReverseReader) Close() {
}
