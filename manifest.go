package buildpack

import (
	"io/ioutil"

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

const defaultVersionsError = "The buildpack manifest is misconfigured for 'default_versions'. " +
	"Contact your Cloud Foundry operator/admin. For more information, see " +
	"https://docs.cloudfoundry.org/buildpacks/custom.html#specifying-default-versions"
