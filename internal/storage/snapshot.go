package storage

import (
	"io"
	"os"

	"github.com/parthkapoor-dev/golphin/internal/fs"
)

const snapshotFileName = "_snapshot.txt"

func writeSnapshot(dirPath string, cache map[string]*location) error {

	filePath := dirPath + "/" + snapshotFileName
	snapFile, err := fs.EnsureFile(filePath)
	if err != nil {
		return err
	}
	defer snapFile.Close()

	for _, loc := range cache {
		chunk := loc.encodeIdx()
		if _, _, err = fs.WriteChunk(snapFile, chunk); err != nil {
			return err
		}
	}

	return nil

}

func readSnapshot(dirPath string, cache map[string]*location) error {

	filePath := dirPath + "/" + snapshotFileName

	snapFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer snapFile.Close()
	defer os.Remove(filePath)

	revReader, err := fs.NewReverseReader(snapFile)
	if err != nil {
		return err
	}
	defer revReader.Close()

	for {
		line, _, err := revReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		loc, err := decodeIdx(line)
		if err != nil {
			return err
		}

		// TODO: to check if segId exists
		cache[loc.key] = loc
	}

	if err := os.Remove(filePath); err != nil {
		return err
	}

	return nil
}
