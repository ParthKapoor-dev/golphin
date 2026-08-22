package main

import (
	"fmt"
	"os"

	"github.com/parthkapoor-dev/golphin/cli"
	"github.com/parthkapoor-dev/golphin/storage"
)

func main() {
	args := os.Args[1:]

	if err := run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {

	db, err := storage.GetDB("./test", 3)
	if err != nil {
		return fmt.Errorf("Open db: %w", err)
	}
	defer db.Close()

	if err := cli.NewCli(db).Work(args); err != nil {
		return err
	}

	return nil
}
