package buildpackUtils

import "errors"

type Downloader struct {
	Dependency Dependency
	OutputDir  string
	Manifest   *Manifest
}

func (d *Downloader) Run() (string, error) {
	url, err := d.Manifest.GetUrl(d.Dependency)
	return url, err
}

func (m *Manifest) GetUrl(dep Dependency) (string, error) {
	var uri string
	inManifest := false

	for _, entry := range m.ManifestEntries {
		if entry.Dependency == dep {
			uri = entry.URI
			inManifest = true
		}
	}

	if !inManifest {
		return "", errors.New("not found")
	}

	return uri, nil
}
