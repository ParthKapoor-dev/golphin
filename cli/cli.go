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

func (cli *Cli) read(key string) error {
	found, value, err := cli.db.Get(key)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("Invalid Key: %q", key)
	}
	fmt.Println("GET " + key + ": " + value)
	return nil
}

func (cli *Cli) write(key string, value string) error {
	err := cli.db.Set(key, value)
	if err != nil {
		return err
	}
	fmt.Println("SET " + key + ": " + value)
	return nil
}

func (cli *Cli) delete(key string) error {
	err := cli.db.Delete(key)
	if err != nil {
		return err
	}
	fmt.Println("DELETE " + key)
	return nil
}

func (cli *Cli) Work(args []string) error {
	size := len(args)

	if size == 2 && args[0] == "get" {
		return cli.read(args[1])
	} else if size == 3 && args[0] == "set" {
		return cli.write(args[1], args[2])
	} else if size == 2 && args[0] == "delete" {
		return cli.delete(args[1])
	}
	return fmt.Errorf("Invalid arguments")
}
