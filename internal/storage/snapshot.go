package storage

import (
	"fmt"
	"io"
	"os"

	"github.com/parthkapoor-dev/golphin/internal/fs"
)

const snapshotFileName = "_snapshot.txt"
const snapMarker = "**MARKER**"

func (db *Db) writeSnapshot() error {

	filePath := db.dirPath + "/" + snapshotFileName
	snapFile, err := fs.EnsureFile(filePath)
	if err != nil {
		return err
	}
	defer snapFile.Close()

	for _, loc := range db.cache {
		chunk := loc.encodeIdx()
		if _, _, err = fs.WriteChunk(snapFile, chunk); err != nil {
			return err
		}
	}

	// mark EOF to verify that snapshot was correctly created
	if _, _, err = fs.WriteChunk(snapFile, []byte(snapMarker)); err != nil {
		return err
	}

	return nil
}

func (db *Db) readSnapshot() error {

	filePath := db.dirPath + "/" + snapshotFileName

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

	// verify the EOF marker by reading the last line
	line, _, err := revReader.ReadLine()
	if err != nil {
		if err == io.EOF {
			return fmt.Errorf("snapshot is empty")
		}
		return err
	}

	if line != snapMarker {
		return fmt.Errorf("snapshot marker not found")
	}

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
		db.cache[loc.key] = loc

	}

	return nil
}
