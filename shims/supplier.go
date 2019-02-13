package shims

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/cloudfoundry/libbuildpack"
)

type Supplier struct {
	V2AppDir        string
	V3AppDir        string
	V2DepsDir       string
	DepsIndex       string
	V2BuildpackDir  string
	V3BuildpacksDir string
	OrderDir        string
	Installer       Installer
}

const (
	ERROR_FILE = "Error V2 buildpack After V3 buildpack"
)

func (s *Supplier) Supply() error {
	if err := s.SetUpFirstV3Buildpack(); err != nil {
		return errors.Wrap(err, "failed to set up first shimmed buildpack")
	}

	if err := s.RemoveV2DepsIndex(); err != nil {
		return errors.Wrap(err, "failed to remove v2 deps index dir")
	}

	orderFile, err := s.SaveOrderToml()
	if err != nil {
		return errors.Wrap(err, "failed to save shimmed CNB order.toml")
	}

	return s.Installer.InstallCNBS(orderFile, s.V3BuildpacksDir)
}

func (s *Supplier) SetUpFirstV3Buildpack() error {
	exists, err := v3symlinkExists(s.V2AppDir)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if err := moveContent(s.V2AppDir, s.V3AppDir); err != nil {
		return err
	}

	if err := os.Symlink(ERROR_FILE, s.V2AppDir); err != nil {
		return err
	}

	appCFPath := filepath.Join(s.V3AppDir, ".cloudfoundry")
	if err := os.MkdirAll(appCFPath, 0777); err != nil {
		fmt.Println("could not open the cloudfoundry dir")
		return err
	}

	if _, err := os.OpenFile(filepath.Join(appCFPath, libbuildpack.SENTINEL), os.O_RDONLY|os.O_CREATE, 0666); err != nil {
		return err
	}

	return nil
}

func (s *Supplier) RemoveV2DepsIndex() error {
	return os.RemoveAll(filepath.Join(s.V2DepsDir, s.DepsIndex))
}

func (s *Supplier) SaveOrderToml() (string, error) {
	orderFile := filepath.Join(s.OrderDir, fmt.Sprintf("order%s.toml", s.DepsIndex))
	if err := libbuildpack.CopyFile(filepath.Join(s.V2BuildpackDir, "order.toml"), orderFile); err != nil {
		return "", err
	}
	return orderFile, nil
}

func moveContent(source, destination string) error {
	if err := os.RemoveAll(destination); err != nil {
		return err
	}

	if err := os.Rename(source, destination); err != nil {
		return err
	}
	return nil
}

func v3symlinkExists(path string) (bool, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		return true, nil
	}
	return false, nil
}
