package cli

import (
	"fmt"

	"github.com/parthkapoor-dev/golphin/storage"
)

type Cli struct {
	db *storage.Db
}

func NewCli(db *storage.Db) *Cli {
	return &Cli{db}
}

func (cli *Cli) read(key string) {
	found, value, err := cli.db.Get(key)
	if err != nil {
		return
	}
	if !found {
		fmt.Println("Invalid key " + key)
		return
	}
	fmt.Println("GET " + key + ": " + value)
}

func (cli *Cli) write(key string, value string) {
	err := cli.db.Upsert(key, value)
	if err != nil {
		return
	}
	fmt.Println("SET " + key + ": " + value)
}

func (cli *Cli) delete(key string) {
	err := cli.db.Delete(key)
	if err != nil {
		return
	}
	fmt.Println("DELETE " + key)
}

func (cli *Cli) Work(args []string) {
	size := len(args)

	if size == 2 && args[0] == "get" {
		cli.read(args[1])
	} else if size == 3 && args[0] == "set" {
		cli.write(args[1], args[2])
	} else if size == 2 && args[0] == "delete" {
		cli.delete(args[1])
	} else {
		println("Invalid arguments")
	}
}
