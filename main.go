package main

import (
	"fmt"
	"os"

	"github.com/parthkapoor-dev/golphin/cmd"
)

const dbDirPath = "./test"
const maxRecordsPerSegment = 3

func main() {
	args := os.Args[1:]

	if err := cmd.Run(args, dbDirPath, maxRecordsPerSegment); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
