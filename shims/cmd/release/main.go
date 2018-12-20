package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/shims"
)

func init() {
	if len(os.Args) != 2 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}
}

func main() {
	exit(run())
}

func run() error {
	metadataPath := filepath.Join(os.Args[1], ".cloudfoundry", "metadata.toml")
	releaser := shims.Releaser{MetadataPath: metadataPath, Writer: os.Stdout}

	if err := releaser.Release(); err != nil {
		return err
	}
	return nil
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	log.Printf("Failed release step: %s\n", err)
	os.Exit(1)
}
