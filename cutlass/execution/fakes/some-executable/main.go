package main

import (
	"fmt"
	"os"
)

var fail string

func main() {
	fmt.Fprintf(os.Stdout, "Output on stdout\n")
	fmt.Fprintf(os.Stderr, "Output on stderr\n")
	fmt.Printf("Arguments: %v\n", os.Args)

	workingDirectory, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fmt.Printf("PWD: %s\n", workingDirectory)

	if fail == "true" {
		fmt.Fprintf(os.Stdout, "Error on stdout\n")
		fmt.Fprintf(os.Stderr, "Error on stderr\n")
		os.Exit(1)
	}
}
