package shims

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/pkg/errors"
)

var (
	V3LifecycleDep   = "lifecycle"
	V3Detector       = "detector"
	V3Builder        = "builder"
	V3Launcher       = "launcher"
	V3LaunchScript   = "0_shim.sh"
	V3AppDir         = filepath.Join(string(filepath.Separator), "home", "vcap", "app")
	V3LayersDir      = filepath.Join(string(filepath.Separator), "home", "vcap", "deps")
	V3MetadataDir    = filepath.Join(string(filepath.Separator), "home", "vcap", "metadata")
	V3StoredOrderDir = filepath.Join(string(filepath.Separator), "home", "vcap", "order")
	V3BuildpacksDir  = filepath.Join(string(filepath.Separator), "home", "vcap", "cnbs")
)

type Detector interface {
	RunLifecycleDetect() error
}

type LayerMetadata struct {
	Build  bool `toml:"build"`
	Launch bool `toml:"launch"`
	Cache  bool `toml:"cache"`
}

type Finalizer struct {
	V2AppDir        string
	V3AppDir        string
	V2DepsDir       string
	V2CacheDir      string
	V3LayersDir     string
	V3BuildpacksDir string
	DepsIndex       string
	OrderDir        string
	OrderMetadata   string
	GroupMetadata   string
	PlanMetadata    string
	V3LifecycleDir  string
	V3LauncherDir   string
	ProfileDir      string
	Detector        Detector
	Installer       Installer
	Manifest        *libbuildpack.Manifest
	Logger          *libbuildpack.Logger
}

func (f *Finalizer) Finalize() error {
	if err := os.RemoveAll(f.V2AppDir); err != nil {
		return errors.Wrap(err, "failed to remove error file")
	}

	if err := f.MergeOrderTOMLs(); err != nil {
		return errors.Wrap(err, "failed to merge order metadata")
	}

	if err := f.RunV3Detect(); err != nil {
		return errors.Wrap(err, "failed to run V3 detect")
	}

	if err := f.IncludePreviousV2Buildpacks(); err != nil {
		return errors.Wrap(err, "failed to include previous v2 buildpacks")
	}

	if err := f.Installer.InstallLifecycle(f.V3LifecycleDir); err != nil {
		return errors.Wrap(err, "failed to install "+V3Builder)
	}

	if err := f.RestoreV3Cache(); err != nil {
		return errors.Wrap(err, "failed to restore v3 cache")
	}

	if err := f.RunLifeycleBuild(); err != nil {
		return errors.Wrap(err, "failed to run v3 lifecycle builder")
	}

	if err := os.Rename(filepath.Join(f.V3LifecycleDir, V3Launcher), filepath.Join(f.V3LauncherDir, V3Launcher)); err != nil {
		return errors.Wrap(err, "failed to move launcher")
	}

	if err := os.Rename(f.V3AppDir, f.V2AppDir); err != nil {
		return errors.Wrap(err, "failed to move app")
	}

	if err := f.MoveV3Layers(); err != nil {
		return errors.Wrap(err, "failed to move V3 dependencies")
	}

	if err := f.Manifest.StoreBuildpackMetadata(f.V2CacheDir); err != nil {
		return err
	}

	return f.WriteProfileLaunch()
}

func (f *Finalizer) MergeOrderTOMLs() error {
	var orders []order

	if err := parseOrderTOMLs(&orders, f.OrderDir); err != nil {
		return err
	}

	if len(orders) == 0 {
		return errors.New("no order.toml found")
	}

	return encodeTOML(f.OrderMetadata, combineOrders(orders))
}

func (f *Finalizer) RunV3Detect() error {
	_, groupErr := os.Stat(f.GroupMetadata)
	_, planErr := os.Stat(f.PlanMetadata)

	if os.IsNotExist(groupErr) || os.IsNotExist(planErr) {
		return f.Detector.RunLifecycleDetect()
	}

	return nil
}

func (f *Finalizer) IncludePreviousV2Buildpacks() error {
	myIDx, err := strconv.Atoi(f.DepsIndex)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(filepath.Join(f.V2DepsDir, f.DepsIndex)); err != nil {
		return err
	}

	for supplyDepsIndex := myIDx - 1; supplyDepsIndex >= 0; supplyDepsIndex-- {
		v2Layer := filepath.Join(f.V2DepsDir, strconv.Itoa(supplyDepsIndex))
		if _, err := os.Stat(v2Layer); os.IsNotExist(err) {
			continue
		}

		buildpackID := fmt.Sprintf("buildpack.%d", supplyDepsIndex)
		v3Layer := filepath.Join(f.V3LayersDir, buildpackID, "layer")

		if err := f.MoveV2Layers(v2Layer, v3Layer); err != nil {
			return err
		}
		if err := f.WriteLayerMetadata(v3Layer); err != nil {
			return err
		}
		if err := f.RenameEnvDir(v3Layer); err != nil {
			return err
		}
		if err := f.UpdateGroupTOML(buildpackID); err != nil {
			return err
		}
		if err := f.AddFakeCNBBuildpack(buildpackID); err != nil {
			return err
		}
	}

	return nil
}

func (f *Finalizer) MoveV3Layers() error {
	bpLayers, err := filepath.Glob(filepath.Join(f.V3LayersDir, "*"))
	if err != nil {
		return err
	}
	for _, bpLayerPath := range bpLayers {
		base := filepath.Base(bpLayerPath)
		if base == "config" {
			if err := f.moveV3Config(); err != nil {
				return err
			}
		} else {
			if err := f.moveV3Layer(bpLayerPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (f *Finalizer) RestoreV3Cache() error {
	//Copies cache over, and unused layers will get automatically cleaned up after successful build
	cnbCache := filepath.Join(f.V2CacheDir, "cnb")
	if exists, err := libbuildpack.FileExists(cnbCache); err != nil {
		return err
	} else if exists {
		return libbuildpack.MoveDirectory(cnbCache, f.V3LayersDir)
	}
	return nil
}

func (f *Finalizer) RunLifeycleBuild() error {
	cmd := exec.Command(
		filepath.Join(f.V3LifecycleDir, V3Builder),
		"-app", f.V3AppDir,
		"-buildpacks", f.V3BuildpacksDir,
		"-group", f.GroupMetadata,
		"-layers", f.V3LayersDir,
		"-plan", f.PlanMetadata,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "CNB_STACK_ID=org.cloudfoundry.stacks."+os.Getenv("CF_STACK"))

	return cmd.Run()
}

func (f *Finalizer) MoveV2Layers(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}

	return os.Rename(src, dst)
}

func (f *Finalizer) WriteLayerMetadata(path string) error {
	contents := LayerMetadata{true, true, false}
	return encodeTOML(path+".toml", contents)
}

func (f *Finalizer) ReadLayerMetadata(path string) (LayerMetadata, error) {
	contents := LayerMetadata{}
	if _, err := toml.DecodeFile(path, &contents); err != nil {
		return LayerMetadata{}, err
	}
	return contents, nil
}

func (f *Finalizer) RenameEnvDir(dst string) error {
	if err := os.Rename(filepath.Join(dst, "env"), filepath.Join(dst, "env.build")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (f *Finalizer) UpdateGroupTOML(buildpackID string) error {
	var groupMetadata group

	if _, err := toml.DecodeFile(f.GroupMetadata, &groupMetadata); err != nil {
		return err
	}

	groupMetadata.Buildpacks = append([]buildpack{{ID: buildpackID}}, groupMetadata.Buildpacks...)

	return encodeTOML(f.GroupMetadata, groupMetadata)
}

func (f *Finalizer) AddFakeCNBBuildpack(buildpackID string) error {
	buildpackPath := filepath.Join(f.V3BuildpacksDir, buildpackID, "latest")
	if err := os.MkdirAll(buildpackPath, 0777); err != nil {
		return err
	}

	buildpackMetadataFile, err := os.Create(filepath.Join(buildpackPath, "buildpack.toml"))
	if err != nil {
		return err
	}
	defer buildpackMetadataFile.Close()

	if err = encodeTOML(filepath.Join(buildpackPath, "buildpack.toml"), struct {
		Buildpack buildpack `toml:"buildpack"`
		Stacks    []stack   `toml:"stacks"`
	}{
		Buildpack: buildpack{
			ID:      buildpackID,
			Name:    buildpackID,
			Version: "latest",
		},
		Stacks: []stack{{
			ID: "org.cloudfoundry.stacks." + os.Getenv("CF_STACK"),
		}},
	}); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(buildpackPath, "bin"), 0777); err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(buildpackPath, "bin", "build"), []byte(`#!/bin/bash`), 0777)
}

func (f *Finalizer) WriteProfileLaunch() error {
	profileContents := fmt.Sprintf(
		`export CNB_STACK_ID="org.cloudfoundry.stacks.%s"
export CNB_LAYERS_DIR="$DEPS_DIR"
export CNB_APP_DIR="$HOME"
exec $HOME/.cloudfoundry/%s "$2"
`,
		os.Getenv("CF_STACK"), V3Launcher)

	return ioutil.WriteFile(filepath.Join(f.ProfileDir, V3LaunchScript), []byte(profileContents), 0666)
}

func (f *Finalizer) moveV3Config() error {
	if err := os.Rename(filepath.Join(f.V3LayersDir, "config"), filepath.Join(f.V2DepsDir, "config")); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(f.V2AppDir, ".cloudfoundry"), 0777); err != nil {
		return err
	}

	if err := libbuildpack.CopyFile(filepath.Join(f.V2DepsDir, "config", "metadata.toml"), filepath.Join(f.V2AppDir, ".cloudfoundry", "metadata.toml")); err != nil {
		return err
	}
	return nil
}

func (f *Finalizer) moveV3Layer(layersPath string) error {
	layersName := filepath.Base(layersPath)
	tomls, err := filepath.Glob(filepath.Join(layersPath, "*.toml"))
	if err != nil {
		return err
	}

	for _, tomlFile := range tomls {
		decodedToml, err := f.ReadLayerMetadata(tomlFile)
		if err != nil {
			return err
		}

		tomlSize := len(".toml")
		if decodedToml.Cache {
			layerPath := tomlFile[:len(tomlFile)-tomlSize]
			layerName := filepath.Base(layerPath)
			if err := f.cacheLayer(layerPath, layersName, layerName); err != nil {
				return err
			}
		}

	}

	if err := os.MkdirAll(filepath.Join(f.V2DepsDir, layersName), os.ModePerm); err != nil {
		return err
	}

	if err := libbuildpack.CopyDirectory(layersPath, filepath.Join(f.V2DepsDir, layersName)); err != nil {
		return err
	}

	return nil
}

func (f *Finalizer) cacheLayer(v3Path, layersName, layerName string) error {
	cacheDir := filepath.Join(f.V2CacheDir, "cnb", layersName, layerName)
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return err
	}
	err := libbuildpack.CopyFile(v3Path+".toml", cacheDir+".toml")
	if err != nil {
		return err
	}
	return libbuildpack.CopyDirectory(v3Path, cacheDir)
}
