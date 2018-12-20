package shims

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/cloudfoundry/libbuildpack"

	"github.com/BurntSushi/toml"
)

type Detector interface {
	Detect() error
}

type Supplier struct {
	BinDir string

	V2AppDir       string
	V2DepsDir      string
	V2BuildpackDir string
	DepsIndex      string

	V3AppDir        string
	V3BuildpacksDir string
	V3LayersDir     string

	OrderMetadata string
	GroupMetadata string
	PlanMetadata  string

	Detector  Detector
	Installer Installer
}

func (s *Supplier) Supply() error {
	if err := s.Installer.InstallOnlyVersion("v3-builder", s.BinDir); err != nil {
		return err
	}

	if err := s.Installer.InstallCNBS(s.OrderMetadata, s.V3BuildpacksDir); err != nil {
		return err
	}

	if err := s.GetDetectorOutput(); err != nil {
		return err
	}

	if err := os.RemoveAll(s.V3AppDir); err != nil {
		return err
	}

	if err := os.Rename(s.V2AppDir, s.V3AppDir); err != nil {
		return err
	}

	if err := s.AddV2SupplyBuildpacks(); err != nil {
		return err
	}

	if err := s.RunLifeycleBuild(); err != nil {
		return err
	}

	if err := os.Rename(s.V3AppDir, s.V2AppDir); err != nil {
		return err
	}

	if err := s.Installer.InstallOnlyVersion("v3-launcher", s.V2DepsDir); err != nil {
		return err
	}

	return s.MoveV3Layers()
}

func (s *Supplier) MoveV3Layers() error {
	bpLayers, err := filepath.Glob(filepath.Join(s.V3LayersDir, "*"))
	if err != nil {
		return err
	}

	for _, bpLayer := range bpLayers {
		if filepath.Base(bpLayer) == "config" {
			if err := os.Rename(filepath.Join(s.V3LayersDir, "config"), filepath.Join(s.V2DepsDir, "config")); err != nil {
				return err
			}

			if err := os.Mkdir(filepath.Join(s.V2AppDir, ".cloudfoundry"), 0777); err != nil {
				return err
			}

			if err := libbuildpack.CopyFile(filepath.Join(s.V2DepsDir, "config", "metadata.toml"), filepath.Join(s.V2AppDir, ".cloudfoundry", "metadata.toml")); err != nil {
				return err
			}
		} else if err := os.Rename(bpLayer, filepath.Join(s.V2DepsDir, filepath.Base(bpLayer))); err != nil {
			return err
		}
	}

	return nil
}

func (s *Supplier) GetDetectorOutput() error {
	_, groupErr := os.Stat(s.GroupMetadata)
	_, planErr := os.Stat(s.PlanMetadata)

	if os.IsNotExist(groupErr) || os.IsNotExist(planErr) {
		return s.Detector.Detect()
	}

	return nil
}

func (s *Supplier) RunLifeycleBuild() error {
	cmd := exec.Command(
		filepath.Join(s.BinDir, "v3-builder"),
		"-app", s.V3AppDir,
		"-buildpacks", s.V3BuildpacksDir,
		"-group", s.GroupMetadata,
		"-layers", s.V3LayersDir,
		"-plan", s.PlanMetadata,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "PACK_STACK_ID=org.cloudfoundry.stacks."+os.Getenv("CF_STACK"))

	return cmd.Run()
}

func (s *Supplier) AddV2SupplyBuildpacks() error {
	myIDx, err := strconv.Atoi(s.DepsIndex)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(filepath.Join(s.V2DepsDir, s.DepsIndex)); err != nil {
		return err
	}

	for supplyDepsIndex := myIDx - 1; supplyDepsIndex >= 0; supplyDepsIndex-- {
		v2Layer := filepath.Join(s.V2DepsDir, strconv.Itoa(supplyDepsIndex))
		buildpackID := fmt.Sprintf("buildpack.%d", supplyDepsIndex)
		v3Layer := filepath.Join(s.V3LayersDir, buildpackID, "layer")

		if err := s.MoveV2Layers(v2Layer, v3Layer); err != nil {
			return err
		}

		if err := s.RenameEnvDir(v3Layer); err != nil {
			return err
		}

		if err := s.UpdateGroupTOML(buildpackID); err != nil {
			return err
		}

		if err := s.AddFakeCNBBuildpack(buildpackID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Supplier) MoveV2Layers(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
		return err
	}

	return os.Rename(src, dst)
}

func (s *Supplier) RenameEnvDir(dst string) error {
	if err := os.Rename(filepath.Join(dst, "env"), filepath.Join(dst, "env.build")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *Supplier) UpdateGroupTOML(buildpackID string) error {
	var groupMetadata group

	if _, err := toml.DecodeFile(s.GroupMetadata, &groupMetadata); err != nil {
		return err
	}

	groupMetadata.Buildpacks = append([]buildpack{{ID: buildpackID}}, groupMetadata.Buildpacks...)

	f, err := os.Create(s.GroupMetadata)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(groupMetadata)
}

func (s *Supplier) AddFakeCNBBuildpack(buildpackID string) error {
	buildpackPath := filepath.Join(s.V3BuildpacksDir, buildpackID, "latest")
	if err := os.MkdirAll(buildpackPath, 0777); err != nil {
		return err
	}

	buildpackMetadataFile, err := os.Create(filepath.Join(buildpackPath, "buildpack.toml"))
	if err != nil {
		return err
	}
	defer buildpackMetadataFile.Close()

	type buildpack struct {
		ID      string `toml:"id"`
		Name    string `toml:"name"`
		Version string `toml:"version"`
	}
	type stack struct {
		ID string `toml:"id"`
	}

	if err = toml.NewEncoder(buildpackMetadataFile).Encode(struct {
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
