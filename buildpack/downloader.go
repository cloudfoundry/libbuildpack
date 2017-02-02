package buildpack

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

func (d *Downloader) Fetch(dep Dependency, filename string) (string, error) {
	url, err := d.Manifest.GetUrl(dep)

	if err != nil {
		return "", err
	}

	resp, err := http.Get(url)
	defer resp.Body.Close()

	if err != nil {
		return "", err
	}

	blob, _ := ioutil.ReadAll(resp.Body)

	dest := path.Join(d.OutputDir, filename)

	err = os.MkdirAll(d.OutputDir, os.ModePerm)

	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(dest, blob, os.ModePerm)

	if err != nil {
		return "", err
	}

	fmt.Printf("Downloaded [%s] to %s\n", url, dest)
	return dest, nil
}

func (m *Manifest) GetUrl(dep Dependency) (string, error) {
	entry, err := m.GetEntry(dep)

	if err != nil {
		return "", err
	}

	return entry.URI, nil
}

func (m *Manifest) GetEntry(dep Dependency) (ManifestEntry, error) {
	var entry ManifestEntry
	inManifest := false

	for _, e := range m.ManifestEntries {
		if e.Dependency == dep {
			entry = e
			inManifest = true
			break
		}
	}

	if !inManifest {
		otherVersions := m.AllDependencyVersions(dep)

		fmt.Printf("DEPENDENCY MISSING IN MANIFEST:\n\n")

		if otherVersions == nil {
			fmt.Printf("Dependency %s is not provided by this buildpack\n", dep.Name)
		} else {
			fmt.Printf("Version %s of dependency %s is not supported by this buildpack\n", dep.Version, dep.Name)
			fmt.Printf("The versions of %s supported in this buildpack are:\n", dep.Name)

			for _, ver := range otherVersions {
				fmt.Printf("- %s\n", ver)
			}
		}

		return entry, fmt.Errorf("dependency %s %s not found", dep.Name, dep.Version)
	}

	return entry, nil
}

func (m *Manifest) AllDependencyVersions(dep Dependency) []string {
	var depVersions []string

	for _, e := range m.ManifestEntries {
		if e.Dependency.Name == dep.Name {
			depVersions = append(depVersions, e.Dependency.Version)
		}
	}

	return depVersions
}
