package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/shims"
)

var logger = libbuildpack.NewLogger(os.Stdout)

func init() {
	if len(os.Args) != 6 {
		log.Fatal(errors.New("incorrect number of arguments"))
	}
}

func main() {
	exit(finalize(logger))
}

func exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	logger.Error("Failed finalize step: %s", err.Error())
	os.Exit(1)
}

func finalize(logger *libbuildpack.Logger) error {
	v2AppDir := os.Args[1]
	v2CacheDir := os.Args[2]
	v2DepsDir := os.Args[3]
	v2DepsIndex := os.Args[4]
	profileDir := os.Args[5]

	tempDir, err := ioutil.TempDir("", "temp")
	if err != nil {
		return err
	}

	defer os.RemoveAll(tempDir)

	v3AppDir := shims.V3AppDir
	v3LayersDir := shims.V3LayersDir

	storedOrderDir := shims.V3StoredOrderDir
	defer os.RemoveAll(storedOrderDir)

	v3BuildpacksDir := shims.V3BuildpacksDir
	defer os.RemoveAll(v3BuildpacksDir)

	metadataDir := shims.V3MetadataDir
	if err := os.MkdirAll(metadataDir, 0777); err != nil {
		return err
	}
	defer os.RemoveAll(metadataDir)

	orderMetadata := filepath.Join(metadataDir, "order.toml")
	groupMetadata := filepath.Join(metadataDir, "group.toml")
	planMetadata := filepath.Join(metadataDir, "plan.toml")

	v3LifecycleDir := filepath.Join(tempDir, "lifecycle")
	v3LauncherDir := filepath.Join(v2DepsDir, "launcher")

	buildpackDir, err := libbuildpack.GetBuildpackDir()
	if err != nil {
		return err
	}

	manifest, err := libbuildpack.NewManifest(buildpackDir, logger, time.Now())
	if err != nil {
		return err
	}

	installer := shims.NewCNBInstaller(manifest)

	detector := shims.DefaultDetector{
		AppDir:          v3AppDir,
		V3LifecycleDir:  v3LifecycleDir,
		V3BuildpacksDir: v3BuildpacksDir,
		OrderMetadata:   orderMetadata,
		GroupMetadata:   groupMetadata,
		PlanMetadata:    planMetadata,
		Installer:       installer,
	}

	finalizer := shims.Finalizer{
		V2AppDir:        v2AppDir,
		V3AppDir:        v3AppDir,
		V2DepsDir:       v2DepsDir,
		V2CacheDir:      v2CacheDir,
		V3LayersDir:     v3LayersDir,
		V3BuildpacksDir: v3BuildpacksDir,
		DepsIndex:       v2DepsIndex,
		OrderDir:        storedOrderDir,
		OrderMetadata:   orderMetadata,
		GroupMetadata:   groupMetadata,
		PlanMetadata:    planMetadata,
		V3LifecycleDir:  v3LifecycleDir,
		V3LauncherDir:   v3LauncherDir,
		ProfileDir:      profileDir,
		Detector:        detector,
		Installer:       installer,
		Manifest:        manifest,
		Logger:          logger,
	}

	return finalizer.Finalize()
}
