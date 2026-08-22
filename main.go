package main

import (
	"fmt"
	"os"

	"github.com/parthkapoor-dev/golphin/cli"
	"github.com/parthkapoor-dev/golphin/storage"
)

func main() {
	args := os.Args[1:]

	db, err := storage.GetDB("./test", 3)
	if err != nil {
		fmt.Println("Unable to establish db connection")
	}
	defer db.Close()

	cli := cli.NewCli(db)
	cli.Work(args)
}
