package main

import (
	"os"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/shims"
)

var logger = libbuildpack.NewLogger(os.Stdout)

func init() {
	if len(os.Args) != 5 {
		logger.Error("Incorrect number of arguments")
		os.Exit(1)
	}
}

func main() {
	exit(supply(logger))
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	logger.Error("Failed supply step: %s", err.Error())
	os.Exit(1)
}

func supply(logger *libbuildpack.Logger) error {
	v2AppDir := os.Args[1]
	v2CacheDir := os.Args[2]
	v2DepsDir := os.Args[3]
	depsIndex := os.Args[4]

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		return err
	}

	v3AppDir := shims.V3AppDir
	if err := os.MkdirAll(v3AppDir, 0777); err != nil {
		return err
	}

	storedOrderDir := shims.V3StoredOrderDir
	if err := os.MkdirAll(storedOrderDir, 0777); err != nil {
		return err
	}

	v3BuildpacksDir := shims.V3BuildpacksDir
	err = os.MkdirAll(v3BuildpacksDir, 0777)
	if err != nil {
		return err
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		return err
	}
	installer := shims.NewCNBInstaller(manifest)

	supplier := shims.Supplier{
		V2AppDir:        v2AppDir,
		V3AppDir:        v3AppDir,
		V2DepsDir:       v2DepsDir,
		V2CacheDir:      v2CacheDir,
		DepsIndex:       depsIndex,
		V2BuildpackDir:  buildpackDir,
		V3BuildpacksDir: v3BuildpacksDir,
		OrderDir:        storedOrderDir,
		Installer:       installer,
		Manifest:        manifest,
		Logger:          logger,
	}

	return supplier.Supply()
}
