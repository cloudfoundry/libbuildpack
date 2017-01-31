package main

import (
	"fmt"
	"io/ioutil"

	"github.com/sesmith177/go-ce-test/buildpack"

	yaml "gopkg.in/yaml.v2"
)

func main() {
	var m buildpack.Manifest

	manifestData, err := ioutil.ReadFile("./manifest.yml")

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

	depName := "dotnet"

	version, err := m.DefaultVersionFor(depName)

	if err != nil {
		fmt.Printf("DefaultVersions error: %s\n", err)
		fmt.Printf("%s", buildpack.DefaultVersionsError)
		return
	}

	fmt.Printf("DefaultVersionFor %s: %s\n", depName, version)

	filtered_url, err := buildpack.FilterURI("https://user:password@example.com/file.tgz")

	if err != nil {
		fmt.Printf("FilterURI error: %s\n", err)
		return
	}

	fmt.Printf(filtered_url)
}
