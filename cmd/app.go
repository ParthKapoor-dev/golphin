package cmd

import (
	"fmt"

	"github.com/parthkapoor-dev/golphin/internal/cli"
	"github.com/parthkapoor-dev/golphin/internal/storage"
)

func Run(args []string, dbDirPath string, maxRecordsPerSegment int) error {

	db, err := storage.GetDB(dbDirPath, maxRecordsPerSegment)
	if err != nil {
		return fmt.Errorf("Open db: %w", err)
	}
	defer db.Close()

	if err := cli.NewCli(db).Work(args); err != nil {
		return err
	}

	return nil
}
