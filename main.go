package main

import (
	"fmt"
	"os"

	"github.com/parthkapoor-dev/golphin/storage"
)

func main() {
	args := os.Args[1:]
	fmt.Println(len(args))
	if len(args) < 2 {
		fmt.Println("Invalid arguments")
		return
	}

	db, err := storage.GetDB("test.txt")
	if err != nil {
		fmt.Println("Unable to establish db connection")
	}
	defer db.Close()

	if len(args) == 2 && args[0] == "get" {
		value, err := db.Get(args[1])
		if err != nil {
			return
		}
		fmt.Println("GET " + args[1] + ": " + value)
		return
	}

	if len(args) == 3 && args[0] == "set" {
		err := db.Upsert(args[1], args[2])
		if err != nil {
			return
		}
		fmt.Println("SET " + args[1] + ": " + args[2])
		return
	}

	fmt.Println("Invalid arguments")
}
