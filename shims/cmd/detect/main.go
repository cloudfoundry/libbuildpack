package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/shims"
)

var logger = libbuildpack.NewLogger(os.Stderr)

func init() {
	if len(os.Args) != 2 {
		logger.Error("Incorrect number of arguments")
		os.Exit(1)
	}
}

func main() {
	exit(detect(logger))
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	logger.Error("Failed detect step: %s", err.Error())
	os.Exit(1)
}

func detect(logger *libbuildpack.Logger) error {
	v2AppDir := os.Args[1]

	v2BuildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		return err
	}

	tempDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		return errors.Wrap(err, "unable to create temp dir")
	}
	defer os.RemoveAll(tempDir)

	v3BuildpacksDir := shims.V3BuildpacksDir
	if err := os.MkdirAll(v3BuildpacksDir, 0777); err != nil {
		return err
	}

	metadataDir := shims.V3MetadataDir
	if err := os.MkdirAll(metadataDir, 0777); err != nil {
		return err
	}
	orderMetadata := filepath.Join(v2BuildpackDir, "order.toml")
	groupMetadata := filepath.Join(metadataDir, "group.toml")
	planMetadata := filepath.Join(metadataDir, "plan.toml")

	manifest, err := libbuildpack.NewManifest(v2BuildpackDir, logger, time.Now())
	if err != nil {
		return err
	}

	installer := shims.NewCNBInstaller(manifest)

	detector := shims.DefaultDetector{
		V3LifecycleDir:  tempDir,
		AppDir:          v2AppDir,
		V3BuildpacksDir: v3BuildpacksDir,
		OrderMetadata:   orderMetadata,
		GroupMetadata:   groupMetadata,
		PlanMetadata:    planMetadata,
		Installer:       installer,
	}

	return detector.Detect()
}
