package libbuildpack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
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
	ManifestRootDir string
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

	m.ManifestRootDir, err = filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (m *Manifest) DefaultVersion(depName string) (Dependency, error) {
	var defaultVersion Dependency
	var err error
	numDefaults := 0

	for _, dep := range m.DefaultVersions {
		if depName == dep.Name {
			defaultVersion = dep
			numDefaults++
		}
	}

	if numDefaults == 0 {
		err = fmt.Errorf("no default version for %s", depName)
	} else if numDefaults > 1 {
		err = fmt.Errorf("found %d default versions for %s", numDefaults, depName)
	}

	if err != nil {
		Log.Error(defaultVersionsError)
		return Dependency{}, err
	}

	return defaultVersion, nil
}

func (m *Manifest) FetchDependency(dep Dependency, outputFile string) error {
	entry, err := m.getEntry(dep)
	if err != nil {
		return err
	}

	filteredURI, err := filterURI(entry.URI)
	if err != nil {
		return err
	}

	if m.isCached() {
		r := strings.NewReplacer("/", "_", ":", "_", "?", "_", "&", "_")
		source := filepath.Join(m.ManifestRootDir, "dependencies", r.Replace(filteredURI))
		Log.Info("Copy [%s]", source)
		err = copyFile(source, outputFile)
	} else {
		Log.Info("Download [%s]", filteredURI)
		err = downloadFile(entry.URI, outputFile)
	}
	if err != nil {
		return err
	}

	err = checkMD5(outputFile, entry.MD5)
	if err != nil {
		os.Remove(outputFile)
		return err
	}
	Log.Info("to [%s]", outputFile)

	return nil
}

func (m *Manifest) getEntry(dep Dependency) (*ManifestEntry, error) {
	for _, e := range m.ManifestEntries {
		if e.Dependency == dep {
			return &e, nil
		}
	}

	Log.Error(dependencyMissingError(m, dep))
	return nil, fmt.Errorf("dependency %s %s not found", dep.Name, dep.Version)
}

func (m *Manifest) isCached() bool {
	dependenciesDir := filepath.Join(m.ManifestRootDir, "dependencies")

	_, err := os.Stat(dependenciesDir)
	if err != nil && os.IsNotExist(err) {
		return false
	}

	return true
}

func (m *Manifest) allDependencyVersions(depName string) []string {
	var depVersions []string

	for _, e := range m.ManifestEntries {
		if e.Dependency.Name == depName {
			depVersions = append(depVersions, e.Dependency.Version)
		}
	}

	return depVersions
}
