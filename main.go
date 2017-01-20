package main

import (
	"fmt"
	"io/ioutil"

	buildpackUtils "github.com/sesmith177/go-ce-test/buildpackUtils"

	yaml "gopkg.in/yaml.v2"
)

func main() {
	var m buildpackUtils.Manifest

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
