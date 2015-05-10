package main

import (
	"fmt"
	"os"
)

func ErrExit(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
