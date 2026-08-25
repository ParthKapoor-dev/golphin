package fs

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type Skip struct {
	Start int64
	End   int64
}

func ReplaceFile(skips []Skip, file *os.File) (*os.File, error) {

	var currentPos int64 = 0

	oldFileName := file.Name()
	oldFile := file

	newFileName := strings.Split(oldFileName, ".txt")[0] + "_new.txt"
	newFile, err := EnsureFile(newFileName)
	if err != nil {
		return file, err
	}

	// for each skip
	skipSize := len(skips)
	for i := skipSize - 1; i >= 0; i-- {
		skip := skips[i]
		bytesToCopy := skip.Start - currentPos

		if bytesToCopy > 0 {

			if _, err := oldFile.Seek(currentPos, io.SeekStart); err != nil {
				return file, fmt.Errorf("failed to seek: %w", err)
			}

			if _, err := io.CopyN(newFile, oldFile, bytesToCopy); err != nil {
				return file, fmt.Errorf("failed to copy chunk: %w", err)
			}

		}

		if skip.End > currentPos {
			currentPos = skip.End
		}
	}

	// copy whatever is left
	if _, err := oldFile.Seek(currentPos, io.SeekStart); err != nil {
		return file, fmt.Errorf("failed to seek: %w", err)
	}

	if _, err := io.Copy(newFile, oldFile); err != nil {
		return file, fmt.Errorf("failed to copy chunk: %w", err)
	}

	newFileStat, err := newFile.Stat()
	if err != nil {
		return file, fmt.Errorf("failed to stat new file: %w", err)
	}

	oldFileStat, err := oldFile.Stat()
	if err != nil {
		return file, fmt.Errorf("failed to stat old file: %w", err)
	}

	oldFile.Close()
	newFile.Close()

	fileCompactionSuccess := newFileStat.Size() < oldFileStat.Size()

	if fileCompactionSuccess {
		if err := os.Remove(oldFileName); err != nil {
			return nil, fmt.Errorf("failed to delete original src file: %w", err)
		}

		if err := os.Rename(newFileName, oldFileName); err != nil {
			return nil, fmt.Errorf("failed to rename dst file: %w", err)
		}

	} else {
		os.Remove(newFileName)
	}

	finalFile, err := EnsureFile(oldFileName)
	if err != nil {
		return nil, err
	}

	if !fileCompactionSuccess {
		return finalFile, fmt.Errorf("compaction process failed!")
	}

	return finalFile, nil
}
