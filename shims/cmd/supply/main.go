package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
	v2DepsDir := os.Args[3]
	depsIndex := os.Args[4]

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		return err
	}

	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	v3LayersDir := filepath.Join(string(filepath.Separator), "home", "vcap", "deps")
	err = os.MkdirAll(v3LayersDir, 0777)
	if err != nil {
		return err
	}
	defer os.RemoveAll(v3LayersDir)

	v3AppDir := filepath.Join(string(filepath.Separator), "home", "vcap", "app")

	v3BuildpacksDir := filepath.Join(tempDir, "cnbs")
	err = os.MkdirAll(v3BuildpacksDir, 0777)
	if err != nil {
		return err
	}
	defer os.RemoveAll(v3BuildpacksDir)

	orderMetadata := filepath.Join(buildpackDir, "order.toml")
	groupMetadata := filepath.Join(tempDir, "group.toml")
	planMetadata := filepath.Join(tempDir, "plan.toml")
	binDir := filepath.Join(tempDir, "bin")

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		return err
	}

	installer := shims.NewCNBInstaller(manifest)

	detector := shims.DefaultDetector{
		BinDir: binDir,

		V2AppDir: v2AppDir,

		V3BuildpacksDir: v3BuildpacksDir,

		OrderMetadata: orderMetadata,
		GroupMetadata: groupMetadata,
		PlanMetadata:  planMetadata,

		Installer: installer,
	}

	supplier := shims.Supplier{
		BinDir: binDir,

		V2AppDir:       v2AppDir,
		V2DepsDir:      v2DepsDir,
		V2BuildpackDir: buildpackDir,
		DepsIndex:      depsIndex,

		V3AppDir:        v3AppDir,
		V3BuildpacksDir: v3BuildpacksDir,
		V3LayersDir:     v3LayersDir,

		OrderMetadata: orderMetadata,
		GroupMetadata: groupMetadata,
		PlanMetadata:  planMetadata,

		Detector:  detector,
		Installer: installer,
	}

	return supplier.Supply()
}
