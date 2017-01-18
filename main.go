package main

import (
	"fmt"
	"io/ioutil"

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
	CFStacks   []string   `yam:"cf_stacks"`
}

type URLToDependencyMap struct {
	Match   string `yaml:"match"`
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Manifest struct {
	Language        string               `yaml:"language"`
	DefaultVersions []Dependency         `yaml:"default_versions"`
	URLMaps         []URLToDependencyMap `yaml:"url_to_dependency_map"`
	ManifestEntries []ManifestEntry      `yaml:"dependencies"`
	ExcludeFiles    []string             `yaml:"exclude_files"`
}

func main() {
	var m Manifest

	manifestData, err := ioutil.ReadFile("/Users/pivotal/workspace/dotnet-core-buildpack/manifest.yml")

	err = yaml.Unmarshal(manifestData, &m)

	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("language: %s\n", m.Language)

	fmt.Printf("default_versions:\n")
	for _, dep := range m.DefaultVersions {
		fmt.Printf("  - name: %s\n    version: %s\n", dep.Name, dep.Version)
	}

	fmt.Printf("url_to_dependency_map:\n")
	for _, um := range m.URLMaps {
		fmt.Printf("  - match: %s\n    name: %s\n    version: %s\n", um.Match, um.Name, um.Version)
	}

	fmt.Printf("dependencies:\n")
	for _, entry := range m.ManifestEntries {
		fmt.Printf("  - name: %s\n    version: %s\n    uri: %s\n    md5: %s\n", entry.Dependency.Name, entry.Dependency.Version, entry.URI, entry.MD5)
	}

}
