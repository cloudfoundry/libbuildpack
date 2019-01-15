package main

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/shims"
)

func main() {
	if len(os.Args) != 6 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}

	finalizer := shims.Finalizer{
		V2AppDir:   os.Args[1],
		V3AppDir:   filepath.Join(string(filepath.Separator), "home", "vcap", "app"),
		DepsIndex:  os.Args[4],
		ProfileDir: os.Args[5],
	}
	if err := finalizer.Finalize(); err != nil {
		log.Fatal(err)
	}
}
