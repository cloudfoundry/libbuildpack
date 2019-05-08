package shims

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/BurntSushi/toml"
	"github.com/cloudfoundry/libbuildpack"
)

type buildpack struct {
	ID       string `toml:"id"`
	Name     string `toml:"name,omitempty"`
	Version  string `toml:"version"`
	Optional bool   `toml:"optional,omitempty"`
}

type group struct {
	Labels     []string    `toml:"labels"`
	Buildpacks []buildpack `toml:"buildpacks"`
}

type order struct {
	Groups []group `toml:"groups"`
}

type stack struct {
	ID string `toml:"id"`
}

type CNBInstaller struct {
	*libbuildpack.Installer
	manifest *libbuildpack.Manifest
}

func NewCNBInstaller(manifest *libbuildpack.Manifest) *CNBInstaller {
	return &CNBInstaller{libbuildpack.NewInstaller(manifest), manifest}
}

func (c *CNBInstaller) InstallCNBS(orderFile string, installDir string) error {
	o := order{}
	_, err := toml.DecodeFile(orderFile, &o)
	if err != nil {
		return err
	}

	bpSet := make(map[string]interface{})
	for _, group := range o.Groups {
		for _, bp := range group.Buildpacks {
			bpSet[bp.ID] = nil
		}
	}

	for buildpack := range bpSet {
		versions := c.manifest.AllDependencyVersions(buildpack)
		if len(versions) != 1 {
			return fmt.Errorf("unable to find a unique version of %s in the manifest", buildpack)
		}

		buildpackDest := filepath.Join(installDir, buildpack, versions[0])
		if exists, err := libbuildpack.FileExists(buildpackDest); err != nil {
			return err
		} else if exists {
			continue
		}

		err := c.InstallOnlyVersion(buildpack, buildpackDest)
		if err != nil {
			return err
		}

		err = os.Symlink(buildpackDest, filepath.Join(installDir, buildpack, "latest"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CNBInstaller) InstallLifecycle(dst string) error {
	tempDir, err := ioutil.TempDir("", "lifecycle")
	if err != nil {
		return errors.Wrap(err, "InstallLifecycle issue creating tempdir")
	}

	defer os.RemoveAll(tempDir)

	if err := c.InstallOnlyVersion(V3LifecycleDep, tempDir); err != nil {
		return err
	}

	firstDir, err := filepath.Glob(filepath.Join(tempDir, "*"))
	if err != nil {
		return err
	}

	if len(firstDir) != 1 {
		return errors.Errorf("issue unpacking lifecycle : incorrect dir format : %s", firstDir)
	}

	for _, binary := range []string{V3Detector, V3Builder, V3Launcher} {
		srcBinary := filepath.Join(firstDir[0], binary)
		dstBinary := filepath.Join(dst, binary)
		if err := os.Rename(srcBinary, dstBinary); err != nil {
			return errors.Wrapf(err, "issue copying lifecycle binary: %s", srcBinary)
		}
	}

	return nil
}
