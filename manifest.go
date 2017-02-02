package buildpack

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Dependency struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type ManifestEntry struct {
	Dependency Dependency `yaml:",inline"`
	URI        string     `yaml:"uri"`
	MD5        string     `yaml:"md5"`
	CFStacks   []string   `yaml:"cf_stacks"`
}

type Manifest struct {
	Language        string          `yaml:"language"`
	DefaultVersions []Dependency    `yaml:"default_versions"`
	ManifestEntries []ManifestEntry `yaml:"dependencies"`
}

func NewManifest(filename string) (*Manifest, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var m Manifest
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

const defaultVersionsError = "The buildpack manifest is misconfigured for 'default_versions'. " +
	"Contact your Cloud Foundry operator/admin. For more information, see " +
	"https://docs.cloudfoundry.org/buildpacks/custom.html#specifying-default-versions"

func (m *Manifest) DefaultVersion(depName string) (string, error) {
	var defaultVersion string
	numDefaults := 0

	for _, dep := range m.DefaultVersions {
		if depName == dep.Name {
			defaultVersion = dep.Version
			numDefaults++
		}
	}

	if numDefaults == 0 {
		return "", newBuildpackError(defaultVersionsError, "no default version for %s", depName)
	} else if numDefaults > 1 {
		return "", newBuildpackError(defaultVersionsError, "found %d default versions for %s", numDefaults, depName)
	}

	return defaultVersion, nil
}

func (m *Manifest) FetchDependency(dep Dependency, outputFile string) error {
	entry, err := m.getEntry(dep)

	if err != nil {
		return err
	}

	resp, err := http.Get(entry.URI)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		return err
	}

	fh, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fh, resp.Body)
	if err != nil {
		return err
	}

	filteredURI, err := filterURI(entry.URI)
	if err != nil {
		return err
	}

	err = checkMD5(outputFile, entry.MD5)
	if err != nil {
		os.Remove(outputFile)
		return err
	}

	fmt.Printf("Downloaded [%s]\n         to [%s]\n", filteredURI, outputFile)

	return nil
}

func (m *Manifest) getEntry(dep Dependency) (*ManifestEntry, error) {
	for _, e := range m.ManifestEntries {
		if e.Dependency == dep {
			return &e, nil
		}
	}
	return nil, newBuildpackError("FIXME", "dependency %s %s not found", dep.Name, dep.Version)
}

func checkMD5(filePath, expectedMD5 string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	hashInBytes := hash.Sum(nil)[:16]
	actualMD5 := hex.EncodeToString(hashInBytes)

	if actualMD5 != expectedMD5 {
		return newBuildpackError("FIXME", "md5 mismatch: expected: %s got: %s", expectedMD5, actualMD5)
	}
	return nil
}
